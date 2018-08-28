package aggregator

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
)

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

// aggAlert generates new aggregated alerts from originals
func aggAlert(ctx context.Context, tx *models.Tx, g grouper, orig []*models.Alert) error {
	rule := g.getRule()
	desc := ""
	var origIds []int64
	for _, o := range orig {
		origIds = append(origIds, o.Id)
		// TODO send notif for all originals
		desc += o.Description + "\n"
	}
	agg := models.NewAlert(
		rule.Alert.Name,
		desc,
		"Various",
		rule.Alert.Config.Source,
		"aggregated",
		orig[0].ExternalId,
		orig[0].StartTime.Time,
		rule.Alert.Config.Severity,
		true)

	agg.AddTags(rule.Alert.Config.Tags...)
	if rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire {
		agg.SetAutoExpire(rule.Alert.Config.ExpireAfter)
	}
	if rule.Alert.Config.AutoClear != nil {
		agg.AutoClear = *rule.Alert.Config.AutoClear
	}
	var newId int64
	stmt, err := tx.PrepareNamed(models.QueryInsertNew)
	err = stmt.Get(&newId, agg)
	if err != nil {
		return fmt.Errorf("Unable to insert agg alert: %v", err)
	}
	agg.Id = newId
	// update the agg IDs of all the original alerts
	err = tx.InQuery(models.QueryUpdateAggId, agg.Id, origIds)
	if err != nil {
		return fmt.Errorf("Unable to update agg Ids: %v", err)
	}
	// send the aggAlert to the right output
	ah.NotifyOutputs(
		&ah.AlertEvent{Alert: agg, Type: ah.EventType_ACTIVE},
		rule.Alert.Config.Outputs,
	)
	return nil
}

type grouper interface {
	start(ctx context.Context, db *models.DB)
	setRule(rule ah.AggregationRuleConfig)
	getRule() ah.AggregationRuleConfig
	addAlert(alert *models.Alert)
}

var groupers = make(map[string]grouper)

func addGrouper(ruleName string, g grouper) {
	groupers[ruleName] = g
}

type Aggregator struct {
	Notif chan *ah.AlertEvent
}

func (a *Aggregator) Name() string {
	return "aggregator"
}

func (a *Aggregator) handleEvent(ctx context.Context, db *models.DB, g grouper, event *ah.AlertEvent) {
	if event.Type != ah.EventType_CLEARED && event.Type != ah.EventType_EXPIRED {
		return
	}
	tx := models.NewTx(db)
	rule := g.getRule()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
		var grouped models.Alerts
		err := tx.Select(&grouped, models.QuerySelectAggregated, event.Alert.AggregatorId)
		if err != nil {
			return err
		}
		var condition bool
		switch event.Type {
		case ah.EventType_CLEARED:
			condition = grouped.AllCleared()
		case ah.EventType_EXPIRED:
			condition = grouped.AllExpired() && rule.Alert.Config.AutoExpire != nil && *rule.Alert.Config.AutoExpire
		default:
			return nil
		}
		if condition {
			_, err := tx.Exec(models.QueryUpdateStatusById, event.Alert.Status, event.Alert.AggregatorId)
			if err != nil {
				return err
			}
			aggAlert := &models.Alert{}
			err = tx.Get(aggAlert, models.QuerySelectById, event.Alert.AggregatorId)
			if err != nil {
				return err
			}
			ah.NotifyOutputs(
				&ah.AlertEvent{Alert: aggAlert, Type: event.Type},
				rule.Alert.Config.Outputs)
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Failed to update agg alert : %v", err)
	}
}

func (a *Aggregator) Start(ctx context.Context, db *models.DB) {
	for name, grouper := range groupers {
		if rule, ok := ah.Config.GetAggregationRuleConfig(name); ok {
			grouper.setRule(rule)
			go grouper.start(ctx, db)
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
					go func() {
						grouper.addAlert(event.Alert)
					}()
				} else {
					a.handleEvent(ctx, db, grouper, event)
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
