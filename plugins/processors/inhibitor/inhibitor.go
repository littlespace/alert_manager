package inhibitor

import (
	"context"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sync"
	"time"
)

type Inhibitor struct {
	Notif    chan *ah.AlertEvent
	db       models.Dbase
	alertBuf map[string][]*models.Alert

	sync.Mutex
}

func (i *Inhibitor) Name() string {
	return "inhibitor"
}

func (i *Inhibitor) ruleAlerts(name string) []*models.Alert {
	i.Lock()
	defer i.Unlock()
	return i.alertBuf[name]
}

func (i *Inhibitor) addAlert(name string, alert *models.Alert) {
	i.Lock()
	defer i.Unlock()
	for _, a := range i.alertBuf[name] {
		if a.Id == alert.Id {
			return
		}
	}
	i.alertBuf[name] = append(i.alertBuf[name], alert)
}

func (i *Inhibitor) notify() {
	for _, alerts := range i.alertBuf {
		for _, a := range alerts {
			if a.Status != models.Status_ACTIVE {
				continue
			}
			event := &ah.AlertEvent{Alert: a, Type: ah.EventType_ACTIVE}
			if alertConfig, ok := ah.Config.GetAlertConfig(a.Name); ok {
				if len(alertConfig.Config.Outputs) > 0 {
					ah.NotifyOutputs(event, alertConfig.Config.Outputs)
				}
			} else {
				ah.NotifyOutputs(event, []string{})
			}
		}
	}
}

func (i *Inhibitor) checkRule(ctx context.Context, rule ah.InhibitRuleConfig) {
	time.Sleep(rule.Delay)
	srcNames := []string{rule.SrcMatch.Alert}
	tx := i.db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		var alerts models.Alerts
		var toInhibit []*models.Alert
		err := tx.InSelect(models.QuerySelectByNames, &alerts, srcNames)
		if err != nil {
			return err
		}
		if len(alerts) == 0 {
			return nil
		}
		for _, src := range alerts {
			srcLabel, ok := src.Labels[rule.SrcMatch.Label]
			if !ok {
				continue
			}
			for _, tgt := range i.ruleAlerts(rule.Name) {
				for _, match := range rule.TargetMatches {
					if tgt.Name == match.Alert && tgt.Labels[match.Label] == srcLabel {
						glog.V(2).Infof("Inhibitor: Found matching inhibit rule for %d:%s", tgt.Id, tgt.Name)
						toInhibit = append(toInhibit, tgt)
					}
				}
			}
		}
		for _, a := range toInhibit {
			if a.Status != models.Status_ACTIVE {
				continue
			}
			a.Status = models.Status_SUPPRESSED
			err := tx.UpdateAlert(a)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Inhibitor: Unable to apply rule %s: %v", rule.Name, err)
	}
	i.Lock()
	i.alertBuf[rule.Name] = i.alertBuf[rule.Name][:0]
	i.Unlock()
}

func (i *Inhibitor) Start(ctx context.Context, db models.Dbase) {
	i.db = db
	// subscribe to target alerts
	toSub := make(map[string]struct{})
	for _, rule := range ah.Config.GetInhibitRules() {
		for _, match := range rule.TargetMatches {
			toSub[match.Alert] = struct{}{}
		}
	}
	for a, _ := range toSub {
		ah.RegisterProcessor(a, i.Notif)
	}
	for {
		select {
		case event := <-i.Notif:
			if event.Type != ah.EventType_ACTIVE {
				break
			}
			for _, rule := range ah.Config.GetInhibitRules() {
				if rule.Delay == 0 {
					// sequentially check alert against every rule
					i.addAlert(rule.Name, event.Alert)
					i.checkRule(ctx, rule)
					continue
				}
				// delay and group alerts before checking rules
				if len(i.alertBuf[rule.Name]) == 0 {
					go i.checkRule(ctx, rule)
				}
				i.addAlert(rule.Name, event.Alert)
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	inh := &Inhibitor{
		Notif:    make(chan *ah.AlertEvent),
		alertBuf: make(map[string][]*models.Alert),
	}
	am.AddProcessor(inh)
}
