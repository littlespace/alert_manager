package aggregator

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins/processors/aggregator/groupers"
	"time"
)

const AGG_CHECK_INTERVAL = 2 * time.Minute

// alertGroup represents a set of grouped alerts for a given grouper
type alertGroup struct {
	groupedAlerts []*models.Alert
	grouper       groupers.Grouper
}

// aggAlert generates an aggregate alert for a given alert group based on defined config.
func (ag alertGroup) aggAlert() *models.Alert {
	rule, _ := ah.Config.GetAggregationRuleConfig(ag.grouper.Name())
	desc := ""
	aggLabels := models.Labels{"device": []string{}, "entity": []string{}, "site": []string{}}
	for _, o := range ag.groupedAlerts {
		desc += o.Description + "\n"
		aggLabels["entity"] = append(aggLabels["entity"].([]string), o.Entity)
		if o.Device.Valid {
			aggLabels["device"] = append(aggLabels["device"].([]string), o.Device.String)
		}
		if o.Site.Valid {
			aggLabels["site"] = append(aggLabels["site"].([]string), o.Site.String)
		}
	}
	agg := models.NewAlert(
		rule.Alert.Name,
		desc,
		"Various",
		rule.Alert.Config.Source,
		"aggregated",
		ag.groupedAlerts[0].ExternalId,
		time.Now(),
		rule.Alert.Config.Severity,
		true)

	agg.Labels = aggLabels
	agg.AddTags(rule.Alert.Config.Tags...)
	if rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire {
		agg.SetAutoExpire(rule.Alert.Config.ExpireAfter)
	}
	if rule.Alert.Config.AutoClear != nil {
		agg.AutoClear = *rule.Alert.Config.AutoClear
	}
	return agg
}

func (ag alertGroup) saveAgg(tx models.Txn) (*models.Alert, error) {
	agg := ag.aggAlert()
	var newId int64
	newId, err := tx.NewAlert(agg)
	if err != nil {
		return nil, fmt.Errorf("Unable to insert agg alert: %v", err)
	}
	agg.Id = newId
	// update the agg IDs of all the original alerts
	var origIds []int64
	for _, o := range ag.groupedAlerts {
		origIds = append(origIds, o.Id)
		tx.NewRecord(o.Id, fmt.Sprintf("Alert aggregated into alert %d", agg.Id))
	}
	tx.NewRecord(agg.Id, fmt.Sprintf("Aggregated alert created from source alerts %v", origIds))
	err = tx.InQuery(models.QueryUpdateAggId, agg.Id, origIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update agg Ids: %v", err)
	}
	return agg, nil
}

var groupedChan = make(chan *alertGroup)

type Aggregator struct {
	Notif   chan *ah.AlertEvent
	grouper *Grouper
	db      models.Dbase

	statAggsActive stats.Stat
	statError      stats.Stat
}

func (a *Aggregator) Name() string {
	return "aggregator"
}

func (a *Aggregator) handleGrouped(ctx context.Context, group *alertGroup) error {
	tx := a.db.NewTx()
	return models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		agg, err := group.saveAgg(tx)
		if err != nil {
			return err
		}
		if err := a.checkSupp(ctx, tx, agg); err != nil {
			return err
		}
		notifier := ah.GetNotifier(a.db)
		go notifier.Notify(&ah.AlertEvent{Alert: agg, Type: ah.EventMap[agg.Status.String()]})
		a.statAggsActive.Add(1)
		return nil
	})
}

func (a *Aggregator) checkSupp(ctx context.Context, tx models.Txn, agg *models.Alert) error {
	supp := ah.GetSuppressor(a.db)
	labels := models.Labels{"alert_name": agg.Name}
	rule := supp.Match(labels, models.MatchCond_ANY)
	if rule != nil && rule.Rtype == models.SuppType_ALERT && rule.TimeLeft() > 0 {
		duration := rule.TimeLeft()
		glog.V(2).Infof("Found matching suppression rule for alert %d: %v", agg.Id, rule)
		msg := fmt.Sprintf("Alert suppressed due to matching suppression Rule %s", rule.Name)
		r := models.NewSuppRule(
			models.Labels{"alert_id": agg.Id},
			"alert",
			msg,
			"alert_manager",
			duration,
		)
		if err := supp.SuppressAlert(ctx, tx, agg, r); err != nil {
			return fmt.Errorf("Unable to suppress agg: %v", err)
		}
		tx.NewRecord(agg.Id, fmt.Sprintf("Alert Suppressed by alert_manager for %v : %s", duration, msg))
		go func() {
			time.Sleep(duration)
			tx := a.db.NewTx()
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
				return supp.UnsuppressAlert(ctx, tx, agg)
			})
			if err != nil {
				glog.Errorf("Failed to unsuppress alert %d: %v", agg.Id, err)
			}
			tx.NewRecord(agg.Id, "Alert unsuppressed")
		}()
		return nil
	}
	return nil
}

func (a *Aggregator) checkExpired(ctx context.Context) error {
	tx := a.db.NewTx()
	return models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		allAggregated, err := tx.SelectAlerts(models.QuerySelectAllAggregated)
		if err != nil {
			return fmt.Errorf("Agg: Unable to query aggregated: %v", err)
		}
		// group by agg-id
		aggGroup := make(map[int64]models.Alerts)
		for _, a := range allAggregated {
			aggGroup[a.AggregatorId.Int64] = append(aggGroup[a.AggregatorId.Int64], a)
		}
		// check if every group needs clear/expiry
		for aggId, alerts := range aggGroup {
			aggAlert, err := tx.GetAlert(models.QuerySelectById, aggId)
			if err != nil {
				return fmt.Errorf("Agg: Unable to query agg alert: %v", err)
			}
			rule, ok := ah.Config.GetAggregationRuleConfig(aggAlert.Source)
			if !ok {
				glog.Errorf("Agg: Cant find rule : %s", aggAlert.Source)
				continue
			}
			var status string
			if alerts.AllExpired() && rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire {
				status = "EXPIRED"
				glog.V(2).Infof("Agg : Agg Alert %d has now expired", aggId)
			} else if alerts.AllInactive() {
				glog.V(2).Infof("Agg : Agg Alert %d has now cleared", aggId)
				status = "CLEARED"
			}
			if status != "" {
				aggAlert.Status = models.StatusMap[status]
				if err := tx.UpdateAlert(aggAlert); err != nil {
					return fmt.Errorf("Agg: Unable to update agg status: %v", err)
				}
				a.statAggsActive.Add(-1)
				tx.NewRecord(aggAlert.Id, fmt.Sprintf("Alert %s", status))
				notifier := ah.GetNotifier(a.db)
				go notifier.Notify(&ah.AlertEvent{Alert: aggAlert, Type: ah.EventMap[status]})
			}
		}
		return nil
	})
}

// StartPoll does grouping based on periodic querying the db for matching alerts.
// Only one of this or Start() must be used to fix the grouping method.
func (a *Aggregator) StartPoll(ctx context.Context, db models.Dbase) {
	a.db = db
	for _, alert := range ah.Config.GetConfiguredAlerts() {
		if len(alert.Config.AggregationRules) == 0 {
			continue
		}
		for _, ruleName := range alert.Config.AggregationRules {
			a.grouper.addSubscription(ruleName, alert.Name)
		}
	}
	for _, gpr := range groupers.AllGroupers {
		go func(g groupers.Grouper) {
			rule, _ := ah.Config.GetAggregationRuleConfig(g.Name())
			t := time.NewTicker(rule.PollInterval)
			for {
				select {
				case <-t.C:
					var alerts []*models.Alert
					tx := a.db.NewTx()
					err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
						return tx.InSelect(models.QuerySelectByNames, &alerts, a.grouper.subscribed(g.Name()))
					})
					if err != nil {
						glog.Errorf("Agg: Unable to query: %v", err)
						return
					}
					if len(alerts) == 0 {
						break
					}
					for _, group := range groupers.DoGrouping(g, alerts) {
						groupedChan <- &alertGroup{groupedAlerts: group, grouper: g}
					}
				case <-ctx.Done():
					return
				}
			}
		}(gpr)
	}
	t := time.NewTicker(AGG_CHECK_INTERVAL)
	for {
		select {
		case <-t.C:
			if err := a.checkExpired(ctx); err != nil {
				a.statError.Add(1)
				glog.Errorf("Agg: Unable to Update Agg Alerts: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Start does grouping by subscribing to alerts from the handler and grouping based
// on configured time windows.
func (a *Aggregator) Start(ctx context.Context, db models.Dbase) {
	//a.StartPoll(ctx, h)
	a.db = db
	t := time.NewTicker(AGG_CHECK_INTERVAL)
	go func() {
		for {
			select {
			case <-t.C:
				if err := a.checkExpired(ctx); err != nil {
					a.statError.Add(1)
					glog.Errorf("Agg: Unable to Update Agg Alerts: %v", err)
				}
			case ag := <-groupedChan:
				if err := a.handleGrouped(ctx, ag); err != nil {
					glog.Errorf("Agg: Unable to save Agg alert: %v", err)
					a.statError.Add(1)
				}
			}
		}
	}()
	// subscribe to alerts that have agg rules defined
	for _, alert := range ah.Config.GetConfiguredAlerts() {
		if len(alert.Config.AggregationRules) > 0 {
			ah.RegisterProcessor(alert.Name, a.Notif)
		}
	}
	for {
		select {
		case event := <-a.Notif:
			// check if alert has already been agg'd
			if event.Alert.AggregatorId.Int64 != 0 {
				break
			}
			config, ok := ah.Config.GetAlertConfig(event.Alert.Name)
			if !ok {
				glog.Errorf("Alert config for %s not found", event.Alert.Name)
				break
			}
			for _, ruleName := range config.Config.AggregationRules {
				grouper, ok := groupers.AllGroupers[ruleName]
				if !ok {
					glog.Errorf("No grouper found for rule: %s, skipping", ruleName)
					continue
				}
				switch event.Type {
				case ah.EventType_ACTIVE:
					a.grouper.addAlert(grouper.Name(), event.Alert)
				case ah.EventType_CLEARED:
					a.grouper.removeAlert(grouper.Name(), event.Alert)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	agg := &Aggregator{
		Notif:          make(chan *ah.AlertEvent),
		grouper:        &Grouper{recvBuffers: make(map[string][]*models.Alert), subs: make(map[string][]string)},
		statAggsActive: stats.NewGauge("processors.aggregator.aggs_active"),
		statError:      stats.NewCounter("processors.aggregator.errors"),
	}
	am.AddProcessor(agg)
}
