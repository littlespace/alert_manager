package handler

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sync"
	"time"
)

const remindCheckInterval = 2 * time.Minute

type notification struct {
	event        *AlertEvent
	lastNotified time.Time
}

type notifier struct {
	notifiedAlerts map[int64]*notification
	db             models.Dbase
	sync.Mutex
}

// Global Notifier Singleton
var notif *notifier
var once sync.Once

func GetNotifier(db models.Dbase) *notifier {
	once.Do(func() {
		notif = &notifier{notifiedAlerts: make(map[int64]*notification), db: db}
		go func() {
			t := time.NewTicker(remindCheckInterval)
			for range t.C {
				notif.remind()
			}
		}()
	})
	return notif
}

func (n *notifier) remind() {
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
		if alertConfig, ok := Config.GetAlertConfig(notif.event.Alert.Name); ok {
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
		if alertConfig, ok := Config.GetAlertConfig(notif.event.Alert.Name); ok {
			n.send(notif.event, alertConfig.Config.Outputs)
		} else {
			n.send(notif.event, []string{})
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
func (n *notifier) Notify(event *AlertEvent) {
	alert := event.Alert
	alertConfig, ok := Config.GetAlertConfig(alert.Name)
	if (ok && alertConfig.Config.DisableNotify) || alert.Owner.Valid {
		n.reportToInflux(event)
		return
	}
	n.Lock()
	defer n.Unlock()
	switch event.Type {
	case EventType_ACTIVE:
		if notif, alreadyNotified := n.notifiedAlerts[alert.Id]; alreadyNotified {
			notif.event = event
			return
		}
		if ok && alert.LastActive.Sub(alert.StartTime.Time) < alertConfig.Config.NotifyDelay {
			return
		}
		n.notifiedAlerts[alert.Id] = &notification{event: event, lastNotified: time.Now()}
	case EventType_CLEARED, EventType_EXPIRED:
		delete(n.notifiedAlerts, alert.Id)
		if event.Type == EventType_CLEARED {
			var notifyOnClear bool
			if ok {
				notifyOnClear = alertConfig.Config.NotifyOnClear
			}
			if !notifyOnClear {
				n.reportToInflux(event)
				return
			}
		}
	case EventType_SUPPRESSED, EventType_ESCALATED, EventType_ACKD:
		if notif, ok := n.notifiedAlerts[alert.Id]; ok {
			if notif.event.Type == EventType_ACTIVE && event.Type == EventType_SUPPRESSED {
				n.reportToInflux(event)
			}
			notif.event = event
			return
		}
		if event.Type == EventType_SUPPRESSED || event.Type == EventType_ACKD {
			n.reportToInflux(event)
			return
		}
	}
	tx := n.db.NewTx()
	ctx := context.Background()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		var err error
		if ok {
			n.send(event, alertConfig.Config.Outputs)
			_, err = tx.NewRecord(event.Alert.Id, fmt.Sprintf("Alert notification sent to %v", alertConfig.Config.Outputs))
		} else {
			n.send(event, []string{})
			_, err = tx.NewRecord(event.Alert.Id, fmt.Sprintf("Alert notification sent to %s", DefaultOutput))
		}
		return err
	})
	if err != nil {
		glog.V(2).Infof("Failed to create notif record: %v", err)
	}
}

func (n *notifier) send(event *AlertEvent, outputs []string) {
	if len(outputs) == 0 {
		outputs = append(outputs, DefaultOutput)
	}
	gMu.Lock()
	for _, output := range outputs {
		if outChan, ok := Outputs[output]; ok {
			glog.V(2).Infof("Sending alert %s to %s", event.Alert.Name, output)
			outChan <- event
		}
	}
	gMu.Unlock()
	n.reportToInflux(event)
}

func (n *notifier) reportToInflux(event *AlertEvent) {
	gMu.Lock()
	defer gMu.Unlock()
	if influxOut, ok := Outputs["influx"]; ok {
		influxOut <- event
	}
}
