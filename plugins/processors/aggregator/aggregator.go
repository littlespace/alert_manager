package aggregator

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
	"github.com/mayuresh82/alert_manager/plugins/processors/aggregator/groupers"
	"regexp"
	"time"
)

const EXPIRY_CHECK_INTERVAL = 2 * time.Minute

// alertGroup represents a set of grouped alerts for a given grouper and agg rule
type alertGroup struct {
	groupedAlerts []*models.Alert
	grouper       groupers.Grouper
	ruleName      string
}

// aggAlert generates an aggregate alert for a given alert group based on defined config.
func (ag alertGroup) aggAlert() *models.Alert {
	rule, _ := ah.Config.GetAggregationRuleConfig(ag.ruleName)
	desc := ag.grouper.AggDesc(ag.groupedAlerts)
	aggLabels := models.Labels{"device": []string{}, "entity": []string{}, "site": []string{}}
	for _, o := range ag.groupedAlerts {
		aggLabels["entity"] = append(aggLabels["entity"].([]string), o.Entity)
		if o.Device.Valid {
			aggLabels["device"] = append(aggLabels["device"].([]string), o.Device.String)
		}
		if o.Site.Valid {
			aggLabels["site"] = append(aggLabels["site"].([]string), o.Site.String)
		}
	}
	for k, v := range rule.Alert.Config.StaticLabels {
		aggLabels[k] = v
	}
	sev := rule.Alert.Config.Severity
	if sev == "" {
		sev = "INFO"
	}
	agg := models.NewAlert(
		rule.Alert.Name,
		desc,
		"Various",
		rule.Name,
		"aggregated",
		ag.groupedAlerts[0].Team,
		ag.groupedAlerts[0].ExternalId,
		time.Now(),
		sev,
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

func (ag alertGroup) saveAgg(tx models.Txn, agg *models.Alert) (int64, error) {
	var newId int64
	newId, err := tx.NewInsert(models.QueryInsertAlert, agg)
	if err != nil {
		return 0, fmt.Errorf("Unable to insert agg alert: %v", err)
	}
	// update the agg IDs of all the original alerts
	var origIds []int64
	for _, o := range ag.groupedAlerts {
		origIds = append(origIds, o.Id)
		tx.NewRecord(o.Id, fmt.Sprintf("Alert aggregated into alert %d", agg.Id))
	}
	tx.NewRecord(agg.Id, fmt.Sprintf("Aggregated alert created from source alerts %v", origIds))
	err = tx.InQuery(models.QueryUpdateAggId, agg.Id, origIds)
	if err != nil {
		return 0, fmt.Errorf("Unable to update agg Ids: %v", err)
	}
	return newId, nil
}

func (ag alertGroup) SuppAgg(tx models.Txn, agg *models.Alert, ruleId int64) error {
	var origIds []int64
	for _, o := range ag.groupedAlerts {
		origIds = append(origIds, o.Id)
		tx.NewRecord(o.Id, fmt.Sprintf("Alert suppressed due to matching supp rule: %d", ruleId))
	}
	// suppress all the original alerts
	err := tx.InQuery(models.QueryUpdateManyStatus, models.Status_SUPPRESSED, origIds)
	if err != nil {
		return fmt.Errorf("Unable to update many status: %v", err)
	}
	return nil
}

var groupedChan = make(chan *alertGroup)

type Aggregator struct {
	Notif   chan *models.AlertEvent
	grouper *Grouper
	db      models.Dbase

	statAggsActive stats.Stat
	statError      stats.Stat
}

func (a *Aggregator) Name() string {
	return "aggregator"
}

func (a *Aggregator) Stage() int {
	return 1
}

func (a *Aggregator) handleGrouped(ctx context.Context, group *alertGroup, out chan *models.AlertEvent) error {
	tx := a.db.NewTx()
	return models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		agg := group.aggAlert()
		supp := ah.GetSuppressor(a.db)
		labels := models.Labels{"alert_name": agg.Name, "entity": "Various"}
		rule := supp.Match(labels)
		if rule != nil && rule.TimeLeft() > 0 {
			glog.V(2).Infof("Found matching suppression rule for alert %s: %d:%s", agg.Name, rule.Id, rule.Name)
			return group.SuppAgg(tx, agg, rule.Id)
		}
		id, err := group.saveAgg(tx, agg)
		if err != nil {
			return err
		}
		agg.Id = id
		a.statAggsActive.Add(1)
		event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: agg}
		out <- event
		if influxOut, ok := ah.GetOutput("influx"); ok {
			influxOut <- event
		}
		return nil
	})
}

func (a *Aggregator) checkExpired(ctx context.Context, out chan *models.AlertEvent) error {
	tx := a.db.NewTx()
	return models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		allAggregated, err := tx.SelectAlerts(models.QuerySelectAllAggregated)
		if err != nil {
			return fmt.Errorf("Agg: Unable to query aggregated: %v", err)
		}
		// group by agg-id
		aggGroup := make(map[int64]models.Alerts)
		for _, a := range allAggregated {
			aggGroup[a.AggregatorId] = append(aggGroup[a.AggregatorId], a)
		}
		// check if every group needs clear/expiry
		for aggId, alerts := range aggGroup {
			aggAlert, err := tx.GetAlert(models.QuerySelectById, aggId)
			if err != nil {
				return fmt.Errorf("Agg: Unable to query agg alert %d: %v", aggId, err)
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
				out <- &models.AlertEvent{Alert: aggAlert, Type: models.EventMap[status]}
			}
		}
		return nil
	})
}

func (a *Aggregator) grouperForAlert(alert *models.Alert, ruleName string) groupers.Grouper {
	var grouper groupers.Grouper
	rule, ok := ah.Config.GetAggregationRuleConfig(ruleName)
	if ok && len(rule.GroupBy) > 0 {
		var match bool
		for k, v := range rule.Matches {
			lv, ok := alert.Labels[k]
			if !ok {
				break
			}
			if _, ok := v.(string); ok {
				match, _ = regexp.MatchString(v.(string), lv.(string))
			} else {
				match = v == lv
			}
			if !match {
				break
			}
		}
		if match {
			grouper = groupers.AllGroupers["default_label_grouper"]
			g := grouper.(*groupers.LabelGrouper)
			g.SetGroupby(rule.GroupBy)
			return g
		}
	} else {
		grouper = groupers.AllGroupers[ruleName]
	}
	return grouper
}

func (a *Aggregator) startProcess(in, out chan *models.AlertEvent) {
	var labelRules []string
	for _, rule := range ah.Config.GetAggRules() {
		if len(rule.GroupBy) > 0 {
			labelRules = append(labelRules, rule.Name)
		}
	}
	glog.Info("Starting processor - Aggregator")
	for event := range in {
		if event.Alert.AggregatorId != 0 || (event.Type != models.EventType_ACTIVE && event.Type != models.EventType_CLEARED) {
			out <- event
			continue
		}
		config, ok := ah.Config.GetAlertConfig(event.Alert.Name)
		var rules []string
		if ok {
			rules = append(rules, config.Config.AggregationRules...)
		}
		if len(rules) == 0 {
			// use any defined label based rules if no rule is specified or alert not configured
			rules = append(rules, labelRules...)
		}
		var processed bool
		for _, ruleName := range rules {
			grouper := a.grouperForAlert(event.Alert, ruleName)
			if grouper == nil {
				glog.V(2).Infof("No grouper found for rule: %s, skipping", ruleName)
				continue
			}
			processed = true
			switch event.Type {
			case models.EventType_ACTIVE:
				a.grouper.addAlert(grouper, ruleName, event.Alert)
			case models.EventType_CLEARED:
				a.grouper.removeAlert(grouper.Name(), event.Alert)
			}
		}
		if !processed {
			out <- event
		}
	}
	close(out)
}

// Process / group the alerts from the handler and grouping based on configured time windows.
func (a *Aggregator) Process(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) chan *models.AlertEvent {
	a.db = db
	out := make(chan *models.AlertEvent)
	go func() {
		t := time.NewTicker(EXPIRY_CHECK_INTERVAL)
		for {
			select {
			case <-t.C:
				if err := a.checkExpired(ctx, out); err != nil {
					a.statError.Add(1)
					glog.Errorf("Agg: Unable to Update Agg Alerts: %v", err)
				}
			case ag := <-groupedChan:
				if err := a.handleGrouped(ctx, ag, out); err != nil {
					glog.Errorf("Agg: Unable to save Agg alert: %v", err)
					a.statError.Add(1)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	go a.startProcess(in, out)
	return out
}

func init() {
	agg := &Aggregator{
		Notif:          make(chan *models.AlertEvent),
		grouper:        &Grouper{recvBuffers: make(map[string][]*models.Alert)},
		statAggsActive: stats.NewGauge("processors.aggregator.aggs_active"),
		statError:      stats.NewCounter("processors.aggregator.errors"),
	}
	plugins.AddProcessor(agg)
}
