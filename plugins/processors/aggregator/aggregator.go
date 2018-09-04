package aggregator

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"time"
)

const AGG_CHECK_INTERVAL = 2 * time.Minute

type groupingFunc func(i, j interface{}) bool

//  generic grouping func
func group(f groupingFunc, items []interface{}) [][]interface{} {
	groups := [][]interface{}{[]interface{}{items[0]}}
	for i := 1; i < len(items); i++ {
		var found bool
		for j := 0; j < len(groups); j++ {
			if f(items[i], groups[j][0]) {
				found = true
				groups[j] = append(groups[j], items[i])
				break
			}
		}
		if !found {
			groups = append(groups, []interface{}{items[i]})
		}
	}
	return groups
}

type alertGroup struct {
	groupedAlerts []*models.Alert
	grouper       grouper
}

func (ag alertGroup) aggAlert() *models.Alert {
	rule, _ := ah.Config.GetAggregationRuleConfig(ag.grouper.name())
	desc := ""
	for _, o := range ag.groupedAlerts {
		// TODO send notif for all groupedAlerts
		desc += o.Description + "\n"
	}
	agg := models.NewAlert(
		rule.Alert.Name,
		desc,
		"Various",
		rule.Alert.Config.Source,
		"aggregated",
		ag.groupedAlerts[0].ExternalId,
		ag.groupedAlerts[0].StartTime.Time,
		rule.Alert.Config.Severity,
		true)

	agg.AddTags(rule.Alert.Config.Tags...)
	if rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire {
		agg.SetAutoExpire(rule.Alert.Config.ExpireAfter)
	}
	if rule.Alert.Config.AutoClear != nil {
		agg.AutoClear = *rule.Alert.Config.AutoClear
	}
	return agg
}

var groupedChan = make(chan *alertGroup)

type grouper interface {
	name() string
	start(ctx context.Context)
	addSubscription(a string)
	subscribed() []string
	addAlert(alert *models.Alert)
	addToBuf(alert *models.Alert)
	doGrouping()
}

var groupers = make(map[string]grouper)

func addGrouper(g grouper) {
	groupers[g.name()] = g
}

type Aggregator struct {
	Notif chan *ah.AlertEvent
	db    models.Dbase
}

func (a *Aggregator) Name() string {
	return "aggregator"
}

func (a *Aggregator) handleGrouped(ctx context.Context) {
	for {
		select {
		case group := <-groupedChan:
			agg := group.aggAlert()
			tx := models.NewTx(a.db)
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
				var newId int64
				newId, err := tx.NewAlert(agg)
				if err != nil {
					return fmt.Errorf("Unable to insert agg alert: %v", err)
				}
				agg.Id = newId
				// update the agg IDs of all the original alerts
				var origIds []int64
				for _, o := range group.groupedAlerts {
					origIds = append(origIds, o.Id)
				}
				err = tx.InQuery(models.QueryUpdateAggId, agg.Id, origIds)
				if err != nil {
					return fmt.Errorf("Unable to update agg Ids: %v", err)
				}
				// send the aggAlert to the right output
				rule, _ := ah.Config.GetAggregationRuleConfig(group.grouper.name())
				ah.NotifyOutputs(
					&ah.AlertEvent{Alert: agg, Type: ah.EventType_ACTIVE},
					rule.Alert.Config.Outputs,
				)
				return nil
			})
			if err != nil {
				glog.Errorf("Agg: Unable to save Agg alert: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *Aggregator) handleExpiry(ctx context.Context) {
	t := time.NewTicker(AGG_CHECK_INTERVAL)
	for {
		select {
		case <-t.C:
			tx := models.NewTx(a.db)
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
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
					if alerts.AllCleared() {
						glog.V(2).Infof("Agg : Agg Alert %d has now cleared", aggId)
						status = "CLEARED"
					}
					if alerts.AllExpired() && rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire {
						status = "EXPIRED"
						glog.V(2).Infof("Agg : Agg Alert %d has now expired", aggId)
					}
					if status != "" {
						aggAlert.Status = models.StatusMap[status]
						if err := tx.UpdateAlert(aggAlert); err != nil {
							return fmt.Errorf("Agg: Unable to update agg status: %v", err)
						}
						ah.NotifyOutputs(&ah.AlertEvent{Alert: aggAlert, Type: ah.EventMap[status]}, rule.Alert.Config.Outputs)
					}
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Agg: Unable to Update Agg Alerts: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// StartPoll does grouping based on periodic querying the db for matching alerts.
// Only one of this or Start() must be used to fix the grouping method.
func (a *Aggregator) StartPoll(ctx context.Context, db models.Dbase) {
	a.db = db
	go a.handleExpiry(ctx)
	for _, alert := range ah.Config.GetConfiguredAlerts() {
		if len(alert.Config.AggregationRules) == 0 {
			continue
		}
		for _, ruleName := range alert.Config.AggregationRules {
			if grouper, ok := groupers[ruleName]; ok {
				grouper.addSubscription(alert.Name)
			}
		}
	}
	for _, gpr := range groupers {
		go func(g grouper) {
			rule, _ := ah.Config.GetAggregationRuleConfig(g.name())
			t := time.NewTicker(rule.PollInterval)
			for {
				select {
				case <-t.C:
					var alerts []*models.Alert
					tx := models.NewTx(db)
					err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
						return tx.InSelect(models.QuerySelectByNames, &alerts, g.subscribed())
					})
					if err != nil {
						glog.Errorf("Agg: Unable to query: %v", err)
						return
					}
					if len(alerts) == 0 {
						break
					}
					for _, a := range alerts {
						g.addToBuf(a)
					}
					g.doGrouping()
				case <-ctx.Done():
					return
				}
			}
		}(gpr)
	}
	a.handleGrouped(ctx)
}

// Start does grouping by subscribing to alerts from the handler and grouping based
// on configured time windows.
func (a *Aggregator) Start(ctx context.Context, db models.Dbase) {
	//a.StartPoll(ctx, db)
	a.db = db
	go a.handleGrouped(ctx)
	go a.handleExpiry(ctx)
	for name, grouper := range groupers {
		if _, ok := ah.Config.GetAggregationRuleConfig(name); ok {
			go grouper.start(ctx)
		} else {
			glog.Errorf("No agg rule defined for grouper: %s, skipping", name)
		}
	}
	// subscribe to alerts that have agg rules defined
	for _, alert := range ah.Config.GetConfiguredAlerts() {
		if len(alert.Config.AggregationRules) > 0 {
			ah.RegisterProcessor(alert.Name, a.Notif)
		}
	}
	for {
		select {
		case event := <-a.Notif:
			config, ok := ah.Config.GetAlertConfig(event.Alert.Name)
			if !ok {
				glog.Errorf("Alert config for %s not found", event.Alert.Name)
				break
			}
			for _, ruleName := range config.Config.AggregationRules {
				grouper, ok := groupers[ruleName]
				if !ok {
					continue
				}
				if event.Type == ah.EventType_ACTIVE {
					go grouper.addAlert(event.Alert)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	agg := &Aggregator{Notif: make(chan *ah.AlertEvent)}
	am.AddProcessor(agg)
}
