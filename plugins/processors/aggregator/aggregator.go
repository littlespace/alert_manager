package aggregator

import (
	"context"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type grouper interface {
	start(ctx context.Context, db *models.DB)
	setRule(rule ah.AggregationRuleConfig)
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

func (a *Aggregator) Start(ctx context.Context, db *models.DB) {
	for _, rule := range ah.Config.GetAggRules() {
		grouper := groupers[rule.Name]
		grouper.setRule(rule)
		go grouper.start(ctx, db)
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
				grouper := groupers[ruleName]
				go func() {
					grouper.addAlert(event.Alert)
				}()
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
