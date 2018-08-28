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
	ESCALATION_CHECK_INTERVAL = 2 * time.Minute
	SUPPRULE_UPDATE_INTERVAL  = 10 * time.Minute
)

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
	Db *models.DB
	// cache of suppression rules
	suppRules models.SuppRules

	sync.Mutex
}

func NewHandler(db *models.DB) *AlertHandler {
	h := &AlertHandler{
		Db: db,
	}
	h.loadSuppRules()
	return h
}

func (h *AlertHandler) loadSuppRules() {
	glog.V(2).Infof("Updating suppression rules")
	tx := models.NewTx(h.Db)
	ctx := context.Background()
	var rules models.SuppRules
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
		return tx.Select(&rules, models.QuerySelectActive+" LIMIT 50")
	})
	if err != nil {
		glog.Errorf("Unable to select rules from db: %v", err)
	}
	h.suppRules = rules
}

func (h *AlertHandler) updateSuppRules(ctx context.Context) {
	t := time.NewTicker(SUPPRULE_UPDATE_INTERVAL)
	for {
		select {
		case <-t.C:
			h.loadSuppRules()
		case <-ctx.Done():
			return
		}
	}
}

// Start needs to be called in a go-routine
func (h *AlertHandler) Start(ctx context.Context) {
	go h.updateSuppRules(ctx)
	go h.handleExpiry(ctx)
	go h.handleEscalation(ctx)
	// start listening for alerts
	for {
		select {
		case alertEvent := <-ListenChan:
			tx := models.NewTx(h.Db)
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
				alert := alertEvent.Alert

				switch alertEvent.Type {
				case EventType_ACTIVE:
					if h.checkExistingActive(tx, alert) {
						return nil
					}
					// add transforms
					h.applyTransforms(alert)
					// new alert
					var newId int64
					stmt, err := tx.PrepareNamed(models.QueryInsertNew)
					err = stmt.Get(&newId, alert)
					if err != nil {
						return fmt.Errorf("Unable to insert new alert: %v", err)
					}
					alert.Id = newId
					glog.V(2).Infof("Received alert with ID: %v", alert.Id)

					// check if alert matches an existing suppression rule
					filters := map[string]string{"Entity": alert.Entity}
					if alert.Device.Valid {
						filters["Device"] = alert.Device.String
					}
					// TODO Add other filters for site, region etc.
					if rule, ok := h.suppRules.Find(filters); ok {
						glog.V(2).Infof("Found matching suppression rule: %v", rule)
						secondsLeft := rule.CreatedAt.Add(time.Duration(rule.Duration) * time.Second).Sub(time.Now())
						h.Suppress(ctx, tx, alert.Id, secondsLeft)
						return nil
					}

					// Send to interested parties
					h.notifyReceivers(alert, EventType_ACTIVE)

				case EventType_CLEARED:
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
					err = h.Clear(ctx, tx, existingAlert.Id)
					if err != nil {
						return fmt.Errorf("Cant clear existing alert %d: %v", existingAlert.Id, err)
					}
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Unable to Create Alert: %v", err)
			}

		case <-ctx.Done():
			glog.V(4).Infof("Closing handler listen loop")
			return
		}
	}
}

func (h *AlertHandler) checkExistingActive(tx *models.Tx, alert *models.Alert) bool {
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
	if !isProcessed {
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
	t := time.NewTicker(EXPIRY_CHECK_INTERVAL)
	for {
		select {
		case <-t.C:
			tx := models.NewTx(h.Db)
			var expired models.Alerts
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
				err := tx.Select(&expired, models.QuerySelectExpired)
				if err != nil {
					return err
				}
				for _, ex := range expired {
					if ex.IsAggregate {
						continue
					}
					glog.V(2).Infof("Alert ID %d has now expired", ex.Id)
					ex.Status = models.Status_EXPIRED
					_, err := tx.Exec(models.QueryUpdateStatusById, ex.Status, ex.Id)
					if err != nil {
						return err
					}
					h.notifyReceivers(&ex, EventType_EXPIRED)
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Failed to update expired alerts: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *AlertHandler) handleEscalation(ctx context.Context) {
	t := time.NewTicker(ESCALATION_CHECK_INTERVAL)
	for {
		select {
		case <-t.C:
			tx := models.NewTx(h.Db)
			var unAckd models.Alerts
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
				err := tx.Select(&unAckd, models.QuerySelectNoOwner)
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
							_, err = tx.Exec(models.QueryUpdateSevById, alert.Severity, alert.Id)
							if err != nil {
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
						h.notifyReceivers(&alert, EventType_ESCALATED)
					}
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Failed to escalate alerts : %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *AlertHandler) GetExisting(tx *models.Tx, alert *models.Alert) (*models.Alert, error) {
	existing := &models.Alert{}
	var err error
	// an alert is assumed to be uniquely identified by its Id or by its Name:Device:Entity
	if alert.Id > 0 {
		err = tx.Get(existing, models.QuerySelectById, alert.Id)
	} else {
		if alert.Device.Valid {
			err = tx.Get(existing, models.QuerySelectByDevice, alert.Name, alert.Entity, alert.Device)
		} else {
			err = tx.Get(existing, models.QuerySelectByNameEntity, alert.Name, alert.Entity)
		}
	}
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (h *AlertHandler) Suppress(ctx context.Context, tx *models.Tx, alertID int64, duration time.Duration) error {
	alert := &models.Alert{}
	err := tx.Get(alert, models.QuerySelectById, alertID)
	if err != nil {
		return err
	}
	alert.Suppress(duration)
	_, err = tx.Exec(models.QueryUpdateStatusById, alert.Id, alert.Status)
	if err != nil {
		return err
	}
	h.notifyReceivers(alert, EventType_SUPPRESSED)
	go func() {
		<-time.NewTimer(duration).C
		alert.Unsuppress()
		// need a new tx here because the closure wont be active any more
		tx := models.NewTx(h.Db)
		models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
			_, err := tx.Exec(models.QueryUpdateStatusById, alert.Id, alert.Status)
			if err != nil {
				glog.Errorf("Failed up update status: %v", err)
				return err
			}
			h.notifyReceivers(alert, EventType_ACTIVE)
			return nil
		})
	}()
	return err
}

func (h *AlertHandler) Clear(ctx context.Context, tx *models.Tx, alertID int64) error {
	alert := &models.Alert{}
	err := tx.Get(alert, models.QuerySelectById, alertID)
	if err != nil {
		return err
	}
	alert.Clear()
	_, err = tx.Exec(models.QueryUpdateStatusById, alert.Status, alert.Id)
	if err != nil {
		return err
	}
	h.notifyReceivers(alert, EventType_CLEARED)
	return nil
}

// SetOwner sets the owner when an alert is acknowledged
func (h *AlertHandler) SetOwner(ctx context.Context, tx *models.Tx, alertID int64, name, teamName string) error {
	alert := &models.Alert{}
	err := tx.Get(alert, models.QuerySelectById, alertID)
	if err != nil {
		return err
	}
	alert.SetOwner(name, teamName)
	_, err = tx.Exec(models.QueryUpdateOwnerById, alert.Owner, alert.Team, alert.Id)
	if err != nil {
		return err
	}
	// Notify all the receivers
	h.notifyReceivers(alert, EventType_ACKD)
	return nil
}
