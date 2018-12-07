package inhibitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
)

type Inhibitor struct {
	Notif    chan *ah.AlertEvent
	db       models.Dbase
	alertBuf map[string][]*models.Alert

	statAlertsInhibited stats.Stat
	statError           stats.Stat

	wg sync.WaitGroup
	sync.Mutex
}

func (i *Inhibitor) Name() string {
	return "inhibitor"
}

func (i *Inhibitor) Stage() int {
	return 0
}

func (i *Inhibitor) SetDb(db models.Dbase) {
	i.db = db
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

func (i *Inhibitor) checkRule(ctx context.Context, rule ah.InhibitRuleConfig, out chan *ah.AlertEvent) {
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
				var inhibit bool
				for _, match := range rule.TargetMatches {
					if tgt.Name == match.Alert && tgt.Labels[match.Label] == srcLabel {
						glog.V(2).Infof("Inhibitor: Found matching inhibit rule for %d:%s", tgt.Id, tgt.Name)
						toInhibit = append(toInhibit, tgt)
						inhibit = true
						break
					}
				}
				if !inhibit {
					out <- &ah.AlertEvent{Type: ah.EventType_ACTIVE, Alert: tgt}
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
	i.alertBuf[rule.Name] = i.alertBuf[rule.Name][:0]
	i.Unlock()
	i.wg.Done()
}

fun (i *Inhibitor) Process(ctx context.Context, in, out chan *ah.AlertEvent, done chan struct{})
	for event := range in {
		if event.Type != ah.EventType_ACTIVE {
			continue
		}
		for _, rule := range ah.Config.GetInhibitRules() {
			// process only interesting alerts
			var toProcess bool
			for _, match := range rule.TargetMatches {
				if match.Alert == in.Alert.Name {
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
				i.wg.Add(1)
				go i.checkRule(ctx, rule)
				
			}
			i.addAlert(rule.Name, event.Alert)
		}
	}
	i.wg.Wait()
	done <- struct{}{}
}

func init() {
	inh := &Inhibitor{
		Notif:               make(chan *ah.AlertEvent),
		alertBuf:            make(map[string][]*models.Alert),
		statAlertsInhibited: stats.NewCounter("processors.inhibitor.alerts_inhibited"),
		statError:           stats.NewCounter("processors.inhibitor.errors"),
	}
	plugins.AddProcessor(inh)
}
