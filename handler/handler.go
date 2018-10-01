package handler

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"regexp"
	"sort"
	"sync"
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
	SUPPRULE_UPDATE_INTERVAL  = 10 * time.Minute
)

var EventMap = map[string]EventType{
	"ACTIVE":     EventType_ACTIVE,
	"EXPIRED":    EventType_EXPIRED,
	"SUPPRESSED": EventType_SUPPRESSED,
	"CLEARED":    EventType_CLEARED,
	"ACKD":       EventType_ACKD,
	"ESCALATED":  EventType_ESCALATED,
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

func NotifyOutputs(event *AlertEvent, outputs []string) {
	// check if notification has already been sent
	// TODO add initial notification delay check, and subequent notify remiders
	if event.Alert.LastActive != event.Alert.StartTime {
		return
	}
	outputs = append(outputs, DefaultOutput)
	for _, output := range outputs {
		if outChan, ok := Outputs[output]; ok {
			glog.V(2).Infof("Sending alert %s to %s", event.Alert.Name, output)
			outChan <- event
		}
	}
}

// AlertHandler handles common alert operations such as expiry, suppression etc.
// It also sends alerts to interested receivers
type AlertHandler struct {
	// db handler
	Db models.Dbase
	// cache of suppression rules
	suppRules models.SuppRules

	sync.Mutex
}

// NewHandler returns a new alert handler which uses the supplied db
func NewHandler(db models.Dbase) *AlertHandler {
	h := &AlertHandler{
		Db: db,
	}
	h.loadSuppRules(context.Background())
	return h
}

func (h *AlertHandler) loadSuppRules(ctx context.Context) {
	h.Lock()
	defer h.Unlock()
	glog.V(2).Infof("Updating suppression rules")
	tx := h.Db.NewTx()
	var (
		rules models.SuppRules
		er    error
	)
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		if rules, er = tx.SelectRules(models.QuerySelectActive + " LIMIT 50"); er != nil {
			return er
		}
		return nil
	})
	if err != nil {
		glog.Errorf("Unable to select rules from db: %v", err)
	}
	h.suppRules = rules

	// load persistent rules from config
	for _, rule := range Config.GetSuppressionRules() {
		for k, v := range rule.Matches {
			ents := models.Labels{k: v}
			r := models.NewSuppRule(ents, rule.Type, rule.Reason, "alert manager", rule.Duration)
			r.DontExpire = true
			h.suppRules = append(h.suppRules, r)
		}
	}
}

func (h *AlertHandler) handleUnsupressOnStart(ctx context.Context) {
	h.Lock()
	defer h.Unlock()
	tx := h.Db.NewTx()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		suppressedAlerts, err := tx.SelectAlerts(models.QuerySelectSuppressed)
		if err != nil {
			return err
		}
		for _, alert := range suppressedAlerts {
			var secondsLeft time.Duration
			for _, rule := range h.suppRules {
				if rule.Rtype != models.SuppType_ALERT {
					continue
				}
				a, ok := rule.Entities["alert_id"]
				if !ok {
					continue
				}
				if alert.Id == int64(a.(float64)) {
					secondsLeft = rule.CreatedAt.Add(time.Duration(rule.Duration) * time.Second).Sub(time.Now())
					break
				}
			}
			a := alert
			go h.unSuppWait(ctx, &a, secondsLeft)
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
		t1 := time.NewTicker(SUPPRULE_UPDATE_INTERVAL)
		t2 := time.NewTicker(EXPIRY_CHECK_INTERVAL)
		t3 := time.NewTicker(ESCALATION_CHECK_INTERVAL)
		for {
			select {
			case <-t1.C:
				h.loadSuppRules(ctx)
			case <-t2.C:
				h.handleExpiry(ctx)
			case <-t3.C:
				h.handleEscalation(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
	go h.handleUnsupressOnStart(ctx)
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
			return fmt.Errorf("Unable to insert new alert: %v", err)
		}
		alert.Id = newId
		glog.V(2).Infof("Received alert with ID: %v", alert.Id)
	}

	// check if alert matches an existing suppression rule based on alert labels
	if rule, ok := h.suppRules.Find(alert.Labels); ok {
		glog.V(2).Infof("Found matching suppression rule: %v", rule)
		secondsLeft := rule.CreatedAt.Add(time.Duration(rule.Duration) * time.Second).Sub(time.Now())
		if rule.DontExpire {
			secondsLeft = time.Duration(rule.Duration) * time.Second
		}
		if secondsLeft > 0 {
			rule := models.NewSuppRule(
				models.Labels{"alert_id": alert.Id},
				"alert",
				fmt.Sprintf("Alert suppressed due to matching suppression rule %s", rule.Name),
				"alert manager",
				secondsLeft)
			h.suppRules = append(h.suppRules, rule)
			return h.Suppress(ctx, tx, alert, rule)
		} else {
			// rule has expired, remove from cache
			for i := 0; i < len(h.suppRules); i++ {
				if h.suppRules[i].Id == rule.Id {
					h.suppRules = append(h.suppRules[:i], h.suppRules[i+1:]...)
					break
				}
			}
		}
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
	err = tx.InQuery(models.QueryUpdateLastActive, models.MyTime{time.Now()}, toUpdate)
	if err != nil {
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
		}
	}
}

func (h *AlertHandler) notifyReceivers(alert *models.Alert, eventType EventType) {
	var isProcessed bool
	gMu.Lock()
	for alertName, recvChans := range Processors {
		if match, _ := regexp.MatchString(alertName, alert.Name); match {
			isProcessed = true
			for _, recvChan := range recvChans {
				recvChan <- &AlertEvent{Alert: alert, Type: eventType}
			}
		}
	}
	gMu.Unlock()
	// if the alert is not subscribed to by any processor, send it directly to the outputs
	if !isProcessed && eventType != EventType_SUPPRESSED {
		event := &AlertEvent{Alert: alert, Type: eventType}
		if alertConfig, ok := Config.GetAlertConfig(alert.Name); ok {
			NotifyOutputs(event, alertConfig.Config.Outputs)
		} else {
			glog.V(2).Infof("No config defined for alert %s, sending to default: %s",
				alert.Name, DefaultOutput)
			NotifyOutputs(event, []string{})
		}
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
						for name, outChan := range Outputs {
							if name == s {
								outChan <- &AlertEvent{Alert: &alert, Type: EventType_ESCALATED}
							}
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

func (h *AlertHandler) Suppress(ctx context.Context, tx models.Txn, alert *models.Alert, rule models.SuppressionRule) error {
	duration := time.Duration(rule.Duration) * time.Second
	alert.Suppress(duration)
	if err := tx.UpdateAlert(alert); err != nil {
		return err
	}
	_, err := tx.NewSuppRule(&rule)
	if err != nil {
		return fmt.Errorf("Unable to save supp rule: %v", err)
	}
	h.notifyReceivers(alert, EventType_SUPPRESSED)
	go h.unSuppWait(ctx, alert, duration)
	return nil
}

func (h *AlertHandler) unSuppWait(ctx context.Context, alert *models.Alert, duration time.Duration) {
	time.Sleep(duration)
	tx := h.Db.NewTx()
	models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		existing, err := h.GetExisting(tx, alert)
		if err != nil {
			return err
		}
		if existing.Status != models.Status_SUPPRESSED {
			glog.V(4).Infof("Alert %d has cleared or expired, not unsuppressing", existing.Id)
			return nil
		}
		alert.Unsuppress()
		if err := tx.UpdateAlert(alert); err != nil {
			glog.Errorf("Failed up update status: %v", err)
			return err
		}
		ListenChan <- &AlertEvent{Type: EventType_ACTIVE, Alert: alert}
		return nil
	})
}

func (h *AlertHandler) Clear(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	alert.Clear()
	if err := tx.UpdateAlert(alert); err != nil {
		return err
	}
	h.notifyReceivers(alert, EventType_CLEARED)
	return nil
}

// SetOwner sets the owner when an alert is acknowledged
func (h *AlertHandler) SetOwner(ctx context.Context, tx models.Txn, alert *models.Alert, name, teamName string) error {
	alert.SetOwner(name, teamName)
	if err := tx.UpdateAlert(alert); err != nil {
		return err
	}
	// Notify all the receivers
	h.notifyReceivers(alert, EventType_ACKD)
	return nil
}
