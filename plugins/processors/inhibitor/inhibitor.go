package inhibitor

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
	"sync"
	"time"
)

type Inhibitor struct {
	db       models.Dbase
	alertBuf map[string][]*models.Alert

	statAlertsInhibited stats.Stat
	statError           stats.Stat

	sync.Mutex
}

func (i *Inhibitor) Name() string {
	return "inhibitor"
}

func (i *Inhibitor) Stage() int {
	return 0
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

func (i *Inhibitor) checkRule(ctx context.Context, rule ah.InhibitRuleConfig, out chan *models.AlertEvent) {
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
						break
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
			tx.NewRecord(a.Id, fmt.Sprintf("Alert Inhibited due to matching inhibit rule: %s", rule.Name))
			i.statAlertsInhibited.Add(1)
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Inhibitor: Unable to apply rule %s: %v", rule.Name, err)
		i.statError.Add(1)
	}
	i.Lock()
	// send the events to the next stage of the pipeline
	for _, alert := range i.alertBuf[rule.Name] {
		if alert.Status == models.Status_SUPPRESSED {
			continue
		}
		event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: alert}
		out <- event
	}
	i.alertBuf[rule.Name] = i.alertBuf[rule.Name][:0]
	i.Unlock()
}

func (i *Inhibitor) Process(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) chan *models.AlertEvent {
	i.db = db
	out := make(chan *models.AlertEvent)
	go func() {
		glog.Info("Starting processor - Inhibitor")
		for event := range in {
			if event.Type != models.EventType_ACTIVE {
				out <- event
				continue
			}
			var anyMatched bool
			for _, rule := range ah.Config.GetInhibitRules() {
				// process only interesting alerts
				var toProcess bool
				for _, match := range rule.TargetMatches {
					if match.Alert == event.Alert.Name {
						toProcess = true
						break
					}
				}
				if !toProcess {
					continue
				}
				if rule.Delay == 0 {
					// sequentially check alert against every rule
					i.addAlert(rule.Name, event.Alert)
					i.checkRule(ctx, rule, out)
					continue
				}
				// delay and group alerts before checking rules
				i.Lock()
				l := len(i.alertBuf[rule.Name])
				i.Unlock()
				if l == 0 {
					go i.checkRule(ctx, rule, out)

				}
				i.addAlert(rule.Name, event.Alert)
				anyMatched = true
			}
			if !anyMatched {
				out <- event
			}
		}
		close(out)
	}()
	return out
}

func init() {
	inh := &Inhibitor{
		alertBuf:            make(map[string][]*models.Alert),
		statAlertsInhibited: stats.NewCounter("processors.inhibitor.alerts_inhibited"),
		statError:           stats.NewCounter("processors.inhibitor.errors"),
	}
	plugins.AddProcessor(inh)
}
