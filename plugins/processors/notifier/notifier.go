package notifier

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	"sync"
	"time"
)

const remindCheckInterval = 2 * time.Minute

type notification struct {
	event        *models.AlertEvent
	lastNotified time.Time
}

type Notifier struct {
	notifiedAlerts map[int64]*notification
	db             models.Dbase
	name           string

	sync.Mutex
}

func (n *Notifier) Name() string {
	return "notifier"
}

func (n *Notifier) Stage() int {
	return 2
}

func (n *Notifier) loadActiveAlerts() {
	n.Lock()
	defer n.Unlock()
	tx := n.db.NewTx()
	ctx := context.Background()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		var active []*models.Alert
		if err := tx.InSelect(models.QuerySelectByStatus, &active, []int64{1}); err != nil {
			return err
		}
		for _, a := range active {
			n.notifiedAlerts[a.Id] = &notification{event: &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: a}}
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Failed to load active alerts: %v", err)
	}
}

func (n *Notifier) remind() {
	n.Lock()
	defer n.Unlock()
	var toNotify []int64
	for alertId, notif := range n.notifiedAlerts {
		if notif.event.Alert.Status == models.Status_SUPPRESSED {
			continue
		}
		if notif.event.Alert.Owner.Valid {
			// dont notify for ackd alerts
			continue
		}
		if alertConfig, ok := ah.Config.GetAlertConfig(notif.event.Alert.Name); ok {
			if alertConfig.Config.NotifyRemind == 0 {
				continue
			}
			if time.Now().Sub(notif.lastNotified) >= alertConfig.Config.NotifyRemind {
				toNotify = append(toNotify, alertId)
			}
		}
	}
	for _, a := range toNotify {
		notif := n.notifiedAlerts[a]
		notif.lastNotified = time.Now()
		glog.V(2).Infof("Sending notification reminder for %d:%s", notif.event.Alert.Id, notif.event.Alert.Name)
		if alertConfig, ok := ah.Config.GetAlertConfig(notif.event.Alert.Name); ok {
			n.send(notif.event, alertConfig.Config.Outputs.Get(notif.event.Alert.Severity.String()))
		} else {
			outputConf := ah.Config.GetOutputConfig()
			n.send(notif.event, outputConf.Defaults.Get(notif.event.Alert.Severity.String()))
		}
	}
}

// Notify notifies about an alert based on the below rules:
// - if the alert config is defined:
//    - Dont notify if alert notifications are disabled for the alert
//    - if the alert is active:
//      - Dont notify if the alert is active for less than the notify_delay if defined
//      - Dont notify if the alert has already been notified once
//      - Notify to the configured outputs or to the default if no ouputs configured
//    - if alert is cleared then notify iff notify_on_clear is set
//    - if alert is expired then notify to configured or default outputs
//    - if alert is suppressed then dont notify
// - else send it to the default output
func (n *Notifier) Notify(event *models.AlertEvent) {
	alert := event.Alert
	alertConfig, ok := ah.Config.GetAlertConfig(alert.Name)
	if ok && alertConfig.Config.DisableNotify {
		return
	}
	n.Lock()
	defer n.Unlock()
	var outputs []string
	if ok {
		outputs = alertConfig.Config.Outputs.Get(event.Alert.Severity.String())
	}
	if len(outputs) == 0 {
		outputConf := ah.Config.GetOutputConfig()
		outputs = outputConf.Defaults.Get(event.Alert.Severity.String())
	}
	notif, alreadyNotified := n.notifiedAlerts[alert.Id]
	if alreadyNotified {
		notif.event = event
	}
	switch event.Type {
	case models.EventType_ACTIVE:
		if alreadyNotified {
			return
		}
		if ok && alert.LastActive.Sub(alert.StartTime.Time) < alertConfig.Config.NotifyDelay {
			return
		}
		n.notifiedAlerts[alert.Id] = &notification{event: event, lastNotified: time.Now()}
	case models.EventType_CLEARED, models.EventType_EXPIRED:
		delete(n.notifiedAlerts, alert.Id)
		if event.Type == models.EventType_CLEARED {
			var notifyOnClear bool
			if ok {
				notifyOnClear = alertConfig.Config.NotifyOnClear
			}
			if !notifyOnClear {
				return
			}
		}
	case models.EventType_SUPPRESSED:
		return
	}
	n.send(event, outputs)
	tx := n.db.NewTx()
	ctx := context.Background()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		msg := fmt.Sprintf("Alert notification sent to %v", outputs)
		_, err := tx.NewRecord(event.Alert.Id, msg)
		return err
	})
	if err != nil {
		glog.V(2).Infof("Failed to create notif record: %v", err)
	}
}

func (n *Notifier) send(event *models.AlertEvent, outputs []string) {
	for _, output := range outputs {
		if outChan, ok := ah.GetOutput(output); ok {
			glog.V(2).Infof("Sending alert %s to %s", event.Alert.Name, output)
			outChan <- event
		}
	}
}

func (n *Notifier) Process(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) chan *models.AlertEvent {
	n.db = db
	n.loadActiveAlerts()
	go func() {
		t := time.NewTicker(remindCheckInterval)
		for range t.C {
			n.remind()
		}
	}()
	out := make(chan *models.AlertEvent)
	go func() {
		glog.Info("Starting processor - Notifier")
		for event := range in {
			go n.Notify(event)
		}
		close(out)
	}()
	return out
}

func init() {
	notif := &Notifier{notifiedAlerts: make(map[int64]*notification)}
	plugins.AddProcessor(notif)
}
