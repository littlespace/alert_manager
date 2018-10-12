package handler

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"regexp"
	"sort"
	"time"
)

type EventType int

const (
	EventType_ACTIVE     EventType = 1
	EventType_EXPIRED    EventType = 2
	EventType_SUPPRESSED EventType = 3
	EventType_CLEARED    EventType = 4
	EventType_ACKD       EventType = 5
	EventType_ESCALATED  EventType = 6

	EXPIRY_CHECK_INTERVAL     = 5 * time.Minute
	ESCALATION_CHECK_INTERVAL = 3 * time.Minute
)

var EventMap = map[string]EventType{
	"ACTIVE":     EventType_ACTIVE,
	"EXPIRED":    EventType_EXPIRED,
	"SUPPRESSED": EventType_SUPPRESSED,
	"CLEARED":    EventType_CLEARED,
	"ACKD":       EventType_ACKD,
	"ESCALATED":  EventType_ESCALATED,
}

func (e EventType) String() string {
	for str, ev := range EventMap {
		if e == ev {
			return str
		}
	}
	return "UNKNOWN"
}

// AlertEvent signifies a type of action on an alert
type AlertEvent struct {
	Alert *models.Alert
	Type  EventType
}

// all listeners send alerts down this channel
var ListenChan = make(chan *AlertEvent)

// default output channel
var DefaultOutput string

// AlertHandler handles common alert operations such as expiry, suppression etc.
// It also sends alerts to interested receivers
type AlertHandler struct {
	// db handler
	Db         models.Dbase
	Notifier   *notifier
	Suppressor *suppressor

	statTransformError stats.Stat
	statDbError        stats.Stat
}

// NewHandler returns a new alert handler which uses the supplied db
func NewHandler(db models.Dbase) *AlertHandler {
	h := &AlertHandler{
		Db:                 db,
		Notifier:           GetNotifier(),
		Suppressor:         GetSuppressor(db),
		statTransformError: stats.NewCounter("handler.transform_errors"),
		statDbError:        stats.NewCounter("handler.db_errors"),
	}
	return h
}

func (h *AlertHandler) handleUnsuppressOnStart(ctx context.Context) {
	tx := h.Db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		var suppressedAlerts []*models.Alert
		if err := tx.InSelect(models.QuerySelectByStatus, &suppressedAlerts, []int64{2}); err != nil {
			h.statDbError.Add(1)
			return err
		}
		if len(suppressedAlerts) == 0 {
			return nil
		}
		var suppIds []int64
		idToAlerts := make(map[int64]*models.Alert)
		for _, alert := range suppressedAlerts {
			idToAlerts[alert.Id] = alert
			suppIds = append(suppIds, alert.Id)
		}
		var alertRules models.SuppRules
		if err := tx.InSelect(models.QuerySelectAlertRules, &alertRules, suppIds); err != nil {
			h.statDbError.Add(1)
			return err
		}
		for _, rule := range alertRules {
			ruleAlertId := int64(rule.Entities["alert_id"].(float64))
			go h.UnsuppWait(ctx, idToAlerts[ruleAlertId], rule.TimeLeft())
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Unable to query suppressed alerts: %v", err)
	}
}

// Start needs to be called in a go-routine
func (h *AlertHandler) Start(ctx context.Context) {
	go func() {
		t1 := time.NewTicker(EXPIRY_CHECK_INTERVAL)
		t2 := time.NewTicker(ESCALATION_CHECK_INTERVAL)
		for {
			select {
			case <-t1.C:
				h.handleExpiry(ctx)
			case <-t2.C:
				h.handleEscalation(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
	go h.handleUnsuppressOnStart(ctx)
	// start listening for alerts
	for {
		select {
		case alertEvent := <-ListenChan:
			tx := h.Db.NewTx()
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
				alert := alertEvent.Alert

				switch alertEvent.Type {
				case EventType_ACTIVE:
					return h.handleActive(ctx, tx, alert)
				case EventType_CLEARED:
					return h.handleClear(ctx, tx, alert)
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Unable to Handle Alert: %v", err)
			}

		case <-ctx.Done():
			glog.V(4).Infof("Closing handler listen loop")
			return
		}
	}
}

func (h *AlertHandler) handleActive(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	if alert.Id == 0 && h.checkExisting(tx, alert) {
		return nil
	}
	// add transforms
	h.applyTransforms(alert)

	// new alert
	if alert.Id == 0 {
		newId, err := tx.NewAlert(alert)
		if err != nil {
			h.statDbError.Add(1)
			return fmt.Errorf("Unable to insert new alert: %v", err)
		}
		alert.Id = newId
		glog.V(2).Infof("Received alert with ID: %v", alert.Id)
	}

	// check if alert matches an existing suppression rule based on alert labels
	// create entities for matching
	labels := make(models.Labels)
	for k, v := range alert.Labels {
		labels[k] = v
	}
	if alert.Device.Valid {
		labels["device"] = alert.Device.String
	}
	labels["entity"] = alert.Entity
	labels["alert_name"] = alert.Name
	if rule, ok := h.Suppressor.Match(labels, models.MatchCond_ANY); ok && rule.TimeLeft() > 0 {
		glog.V(2).Infof("Found matching suppression rule for alert %d: %v", alert.Id, rule)
		return h.Suppress(
			ctx, tx, alert, "alert_manager",
			fmt.Sprintf("Alert suppressed due to matching suppression Rule %s", rule.Name),
			rule.TimeLeft(),
		)
	}

	// Send to interested parties
	h.notifyReceivers(alert, EventType_ACTIVE)
	return nil
}

func (h *AlertHandler) handleClear(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	// clear existing alert if auto clear is true
	existingAlert, err := h.GetExisting(tx, alert)
	if err != nil {
		glog.V(2).Infof("No existing alert found for %s:%s to clear", alert.Name, alert.Entity)
		return nil
	}
	if !existingAlert.AutoClear {
		glog.V(2).Infof("Not auto-clearing alert %d ", existingAlert.Id)
		return nil
	}
	err = h.Clear(ctx, tx, existingAlert)
	if err != nil {
		h.statDbError.Add(1)
		return fmt.Errorf("Cant clear existing alert %d: %v", existingAlert.Id, err)
	}
	return nil
}

func (h *AlertHandler) checkExisting(tx models.Txn, alert *models.Alert) bool {
	existingAlert, err := h.GetExisting(tx, alert)
	if err != nil {
		glog.V(2).Infof("No existing alert found for %s:%s", alert.Name, alert.Entity)
		return false
	}
	// extend the expiry time if alert already exists
	toUpdate := []int64{existingAlert.Id}
	if existingAlert.AggregatorId.Valid {
		toUpdate = append(toUpdate, existingAlert.AggregatorId.Int64)
	}
	newLastActive := models.MyTime{time.Now()}
	existingAlert.LastActive = newLastActive
	err = tx.InQuery(models.QueryUpdateLastActive, newLastActive, toUpdate)
	if err != nil {
		h.statDbError.Add(1)
		glog.Errorf("Failed update last active: %v", err)
	}
	// Send to interested parties
	h.notifyReceivers(existingAlert, EventMap[existingAlert.Status.String()])
	return err == nil
}

func (h *AlertHandler) applyTransforms(alert *models.Alert) {
	// apply transforms in order of priority. Lower == first
	var toApply []Transform
	for _, transform := range Transforms {
		if match, _ := regexp.MatchString(transform.GetRegister(), alert.Name); match {
			toApply = append(toApply, transform)
		}
	}
	sort.Slice(toApply, func(i, j int) bool {
		return toApply[i].GetPriority() < toApply[j].GetPriority()
	})
	for _, xform := range toApply {
		glog.V(2).Infof("Applying Transform: %s to alert %s", xform.Name(), alert.Name)
		if err := xform.Apply(alert); err != nil {
			glog.Errorf("Failed to apply transform %s to alert %s: %v", xform.Name(), alert.Name, err)
			h.statTransformError.Add(1)
		}
	}
}

func (h *AlertHandler) notifyReceivers(alert *models.Alert, eventType EventType) {
	gMu.Lock()
	for alertName, recvChans := range Processors {
		if match, _ := regexp.MatchString(alertName, alert.Name); match {
			for _, recvChan := range recvChans {
				recvChan <- &AlertEvent{Alert: alert, Type: eventType}
			}
		}
	}
	gMu.Unlock()

	// send the alert to the outputs. If the alert config or config outputs is undefined,
	// the notifier will send it to the default output.
	event := &AlertEvent{Alert: alert, Type: eventType}
	h.Notifier.Notify(event)
	// send to influx for reporting
	if influxOut, ok := Outputs["influx"]; ok {
		if eventType == EventType_ACTIVE && alert.StartTime != alert.LastActive {
			return
		}
		influxOut <- event
	}
}

func (h *AlertHandler) handleExpiry(ctx context.Context) {
	tx := h.Db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		expired, err := tx.SelectAlerts(models.QuerySelectExpired)
		if err != nil {
			return err
		}
		for _, ex := range expired {
			if ex.IsAggregate {
				// aggregate expiry handled by aggregators
				continue
			}
			glog.V(2).Infof("Alert ID %d has now expired", ex.Id)
			ex.Status = models.Status_EXPIRED
			if err := tx.UpdateAlert(&ex); err != nil {
				return err
			}
			toSend := ex // this copy needed to avoid overwriting
			h.notifyReceivers(&toSend, EventType_EXPIRED)
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Failed to update expired alerts: %v", err)
		h.statDbError.Add(1)
	}
}

func (h *AlertHandler) handleEscalation(ctx context.Context) {
	tx := h.Db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		unAckd, err := tx.SelectAlerts(models.QuerySelectNoOwner)
		if err != nil {
			return err
		}
		for _, alert := range unAckd {
			config, ok := Config.GetAlertConfig(alert.Name)
			if !ok {
				glog.Errorf("Failed to check escalation for %s : No config found", alert.Name)
				continue
			}
			var changed bool
			for _, rule := range config.Config.EscalationRules {
				newSev := models.SevMap[rule.EscalateTo]
				if newSev >= alert.Severity {
					continue
				}
				timePassed := time.Now().Sub(alert.StartTime.Time)
				if timePassed >= rule.After {
					changed = true
					glog.V(2).Infof("Escalating alert %s:%d to %s", alert.Name, alert.Id, rule.EscalateTo)
					alert.Severity = newSev
					if err := tx.UpdateAlert(&alert); err != nil {
						return err
					}
					for _, s := range rule.SendTo {
						if outChan, ok := Outputs[s]; ok {
							outChan <- &AlertEvent{Alert: &alert, Type: EventType_ESCALATED}
						}
					}
					break
				}
			}
			if changed {
				toSend := alert // this copy needed to avoid overwriting
				h.notifyReceivers(&toSend, EventType_ESCALATED)
			}
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Failed to escalate alerts : %v", err)
		h.statDbError.Add(1)
	}
}

func (h *AlertHandler) GetExisting(tx models.Txn, alert *models.Alert) (*models.Alert, error) {
	var existing *models.Alert
	var err error
	// an alert is assumed to be uniquely identified by its Id or by its Name:Device:Entity
	if alert.Id > 0 {
		existing, err = tx.GetAlert(models.QuerySelectById, alert.Id)
	} else {
		if alert.Device.Valid {
			existing, err = tx.GetAlert(models.QuerySelectByDevice, alert.Name, alert.Entity, alert.Device.String)
		} else {
			existing, err = tx.GetAlert(models.QuerySelectByNameEntity, alert.Name, alert.Entity)
		}
	}
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (h *AlertHandler) Suppress(
	ctx context.Context,
	tx models.Txn,
	alert *models.Alert,
	creator, reason string,
	duration time.Duration,
) error {
	// create an alert specific supp-rule for tracking
	r := models.NewSuppRule(
		models.Labels{"alert_id": alert.Id}, "alert", reason, creator, duration,
	)
	if err := h.Suppressor.SuppressAlert(ctx, tx, alert, r); err != nil {
		return fmt.Errorf("Unable to suppress alert %d: %v", alert.Id, err)
	}
	h.notifyReceivers(alert, EventType_SUPPRESSED)
	go h.UnsuppWait(ctx, alert, duration)
	return nil
}

func (h *AlertHandler) UnsuppWait(ctx context.Context, alert *models.Alert, duration time.Duration) {
	time.Sleep(duration)
	tx := h.Db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		err := h.Suppressor.UnsuppressAlert(ctx, tx, alert)
		if err != nil {
			return err
		}
		return h.handleActive(ctx, tx, alert)
	})
	if err != nil {
		glog.Errorf("Failed to unsuppress alert %d: %v", alert.Id, err)
	}
}

func (h *AlertHandler) Clear(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	alert.Clear()
	if err := tx.UpdateAlert(alert); err != nil {
		h.statDbError.Add(1)
		return err
	}
	h.notifyReceivers(alert, EventType_CLEARED)
	return nil
}

// SetOwner sets the owner when an alert is acknowledged
func (h *AlertHandler) SetOwner(ctx context.Context, tx models.Txn, alert *models.Alert, name, teamName string) error {
	alert.SetOwner(name, teamName)
	if err := tx.UpdateAlert(alert); err != nil {
		h.statDbError.Add(1)
		return err
	}
	// Notify all the receivers
	h.notifyReceivers(alert, EventType_ACKD)
	return nil
}
