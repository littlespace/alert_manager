package handler

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
)

const (
	EXPIRY_CHECK_INTERVAL     = 5 * time.Minute
	ESCALATION_CHECK_INTERVAL = 3 * time.Minute
)

// all listeners send alerts down this channel
var ListenChan = make(chan *models.AlertEvent)

// ClearHandler keeps a track of clearing active alerts
type ClearHandler struct {
	actives map[int64]chan struct{}
	sync.RWMutex
}

func (c *ClearHandler) get(id int64) (chan struct{}, bool) {
	c.RLock()
	defer c.RUnlock()
	resetClear, ok := c.actives[id]
	return resetClear, ok
}

func (c *ClearHandler) add(id int64) chan struct{} {
	c.Lock()
	defer c.Unlock()
	c.actives[id] = make(chan struct{})
	return c.actives[id]
}

func (c *ClearHandler) delete(id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.actives, id)
}

// AlertHandler handles common alert operations such as expiry, suppression etc.
// It also sends alerts to interested receivers
type AlertHandler struct {
	// db handler
	Db            models.Dbase
	Suppressor    *suppressor
	Teams         models.Teams
	procChan      chan *models.AlertEvent
	clearer       *ClearHandler
	clearHolddown time.Duration

	statTransformError stats.Stat
	statDbError        stats.Stat
}

// NewHandler returns a new alert handler which uses the supplied db
func NewHandler(db models.Dbase, clearHolddown time.Duration) *AlertHandler {
	h := &AlertHandler{
		Db:                 db,
		Suppressor:         GetSuppressor(db),
		procChan:           make(chan *models.AlertEvent),
		clearer:            &ClearHandler{actives: make(map[int64]chan struct{})},
		clearHolddown:      clearHolddown,
		statTransformError: stats.NewCounter("handler.transform_errors"),
		statDbError:        stats.NewCounter("handler.db_errors"),
	}
	return h
}

func (h *AlertHandler) loadTeams() error {
	tx := h.Db.NewTx()
	ctx := context.Background()
	var teams models.Teams
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		var er error
		teams, er = tx.SelectTeams(models.QuerySelectTeams)
		return er
	})
	if err != nil {
		return fmt.Errorf("Failed to fetch teams from db: %v", err)
	}
	h.Teams = teams
	return nil
}

func (h *AlertHandler) GetUsersFromConfig() map[string]string {
	return Config.GetTeamConfig().Users
}

// Start needs to be called in a go-routine
func (h *AlertHandler) Start(ctx context.Context) {
	// load teams
	if err := h.loadTeams(); err != nil {
		glog.Exitf("Fatal err: Failed to load teams: %v", err)
	}
	// start the processor pipeline
	procPipeline := plugins.NewProcessorPipeline()
	procPipeline.Run(ctx, h.Db, h.procChan)

	// housekeeping
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
	// start listening for alerts
	for {
		select {
		case alertEvent := <-ListenChan:
			tx := h.Db.NewTx()
			err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
				alert := alertEvent.Alert

				switch alertEvent.Type {
				case models.EventType_ACTIVE:
					return h.handleActive(ctx, tx, alert)
				case models.EventType_CLEARED:
					return h.handleClear(ctx, tx, alert, h.clearHolddown)
				}
				return nil
			})
			if err != nil {
				glog.Errorf("Unable to Handle Alert: %v", err)
			}

		case <-ctx.Done():
			glog.V(4).Infof("Closing handler listen loop")
			close(h.procChan)
			return
		}
	}
}

func (h *AlertHandler) handleActive(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	if h.checkExisting(tx, alert) {
		return nil
	}
	// add transforms
	h.applyTransforms(alert)

	// check if alert matches an existing suppression rule based on alert labels
	alert.ExtendLabels()
	if rule := h.Suppressor.Match(alert.Labels); rule != nil && rule.TimeLeft() > 0 {
		glog.V(2).Infof("Found matching suppression rule for %s:%s:%s: %d:%s", alert.Name, alert.Entity, alert.Device.String, rule.Id, rule.Name)
		return nil
	}
	if h.checkAggregated(tx, alert) {
		return nil
	}
	// new alert
	if !h.Teams.Contains(alert.Team) {
		// create new team
		if err := tx.Exec(models.NewPartition(alert.Team)); err != nil {
			glog.Errorf("Failed to create new team partition: %v", err)
		}
		team := &models.Team{Name: alert.Team}
		id, err := tx.NewInsert(models.QueryInsertTeam, team)
		if err != nil {
			glog.Errorf("Failed to create new team: %v", err)
		}
		team.Id = id
		h.Teams = append(h.Teams, team)
	}
	newId, err := tx.NewInsert(models.QueryInsertAlert, alert)
	if err != nil {
		h.statDbError.Add(1)
		return fmt.Errorf("Unable to insert new alert: %v", err)
	}
	alert.Id = newId
	glog.V(2).Infof("Received alert with ID: %v", alert.Id)
	tx.NewRecord(newId, fmt.Sprintf("Alert created from source %s with severity %s",
		alert.Source, alert.Severity.String()))
	// Send to interested parties
	h.notifyReceivers(alert, models.EventType_ACTIVE)
	return nil
}

func (h *AlertHandler) handleClear(ctx context.Context, tx models.Txn, alert *models.Alert, holddown time.Duration) error {
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
	// wait for a holddown period before clearing the alert to avoid flaps
	if holddown == 0 {
		return h.clearAlert(ctx, tx, existingAlert)
	}
	go func() {
		if _, ok := h.clearer.get(existingAlert.Id); ok {
			return
		}
		t := time.NewTimer(holddown)
		resetClear := h.clearer.add(existingAlert.Id)
		defer h.clearer.delete(existingAlert.Id)
		for {
			select {
			case <-t.C:
				newTx := h.Db.NewTx()
				err := models.WithTx(ctx, newTx, func(ctx context.Context, tx models.Txn) error {
					return h.clearAlert(ctx, tx, existingAlert)
				})
				if err != nil {
					glog.Error(err)
				}
				return
			case <-resetClear:
				return
			}
		}
	}()
	return nil
}

func (h *AlertHandler) clearAlert(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	err := h.Clear(ctx, tx, alert)
	if err != nil {
		h.statDbError.Add(1)
		return fmt.Errorf("Cant clear existing alert %d: %v", alert.Id, err)
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
	if existingAlert.AggregatorId != 0 {
		toUpdate = append(toUpdate, existingAlert.AggregatorId)
	}
	newLastActive := models.MyTime{time.Now()}
	existingAlert.LastActive = newLastActive
	err = tx.InQuery(models.QueryUpdateLastActive, newLastActive, toUpdate)
	if err != nil {
		h.statDbError.Add(1)
		glog.Errorf("Failed update last active: %v", err)
	}
	if resetClear, ok := h.clearer.get(existingAlert.Id); ok {
		resetClear <- struct{}{}
	}
	return err == nil
}

func (h *AlertHandler) checkAggregated(tx models.Txn, alert *models.Alert) bool {
	var dev string
	if alert.Device.Valid {
		dev = alert.Device.String
	}
	existing, err := tx.GetAlert(models.QuerySelectExistingAgg, alert.Name, alert.Entity, dev)
	if err != nil {
		glog.V(4).Infof("No existing alert found")
		return false
	}
	existingAgg, err := tx.GetAlert(models.QuerySelectById, existing.AggregatorId)
	if err != nil {
		glog.V(4).Infof("No existing aggregate alert found")
		return false
	}
	if existingAgg.Status != models.Status_ACTIVE {
		return false
	}
	// ignore the new alert since it was previously aggregated and the agg is still active
	alert.Status = models.Status_SUPPRESSED
	existing.Status = models.Status_ACTIVE
	existing.LastActive = models.MyTime{time.Now()}
	if err = tx.UpdateAlert(existing); err != nil {
		glog.Errorf("Unable to update alert: %v", err)
		return false
	}
	glog.Infof("Found existing aggregated alert for %s: %d", existing.Name, existing.Id)
	return true
}

func (h *AlertHandler) applyTransforms(alert *models.Alert) {
	// apply transforms in order of priority. Lower == first
	var toApply []Transform
	for _, transform := range Transforms {
		if transform.GetRegister() == "" {
			continue
		}
		if match, _ := regexp.MatchString(transform.GetRegister(), alert.Name); match {
			toApply = append(toApply, transform)
		}
	}
	sort.Slice(toApply, func(i, j int) bool {
		return toApply[i].GetPriority() < toApply[j].GetPriority()
	})
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("PANIC while applying transform: %v", r)
			h.statTransformError.Add(1)
		}
	}()
	for _, xform := range toApply {
		glog.V(2).Infof("Applying Transform: %s to alert %s", xform.Name(), alert.Name)
		if err := xform.Apply(alert); err != nil {
			glog.Errorf("Failed to apply transform %s to alert %s: %v", xform.Name(), alert.Name, err)
			h.statTransformError.Add(1)
		}
	}
}

func (h *AlertHandler) notifyReceivers(alert *models.Alert, eventType models.EventType) {
	event := &models.AlertEvent{Alert: alert, Type: eventType}
	// send the alert down the processor pipeline
	if len(plugins.Processors) > 0 {
		h.procChan <- event
	}
	plugins.Send("influx", event)
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
			if err := tx.UpdateAlert(ex); err != nil {
				return err
			}
			tx.NewRecord(ex.Id, "Alert expired")
			toSend := ex // this copy needed to avoid overwriting
			h.notifyReceivers(toSend, models.EventType_EXPIRED)
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
				glog.V(4).Infof("Failed to check escalation for %s : No config found", alert.Name)
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
					alert.SetSeverity(newSev)
					if err := tx.UpdateAlert(alert); err != nil {
						return err
					}
					tx.NewRecord(alert.Id, fmt.Sprintf(
						"Alert severity escalated to %s", newSev.String()))
					break
				}
			}
			if changed {
				toSend := alert // this copy needed to avoid overwriting
				h.notifyReceivers(toSend, models.EventType_ESCALATED)
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
	if alert.IsAggregate {
		var grouped models.Alerts
		if err := tx.InSelect(models.QuerySelectByAggId, &grouped, alert.Id); err != nil {
			return fmt.Errorf("Failed to query alerts: %v", err)
		}
		// suppress individual alerts
		for _, a := range grouped {
			cReason := "Alert suppressed due to aggregated alert suppressed"
			if err := h.Suppress(ctx, tx, a, creator, cReason, duration); err != nil {
				return fmt.Errorf("Unable to suppress alert %d: %v", a.Id, err)
			}
		}
		tx.NewRecord(alert.Id, fmt.Sprintf("Alert Suppressed by %s for %v : %s", creator, duration, reason))
		return h.Suppressor.SuppressAlert(ctx, tx, alert, duration)
	}
	if err := h.Suppressor.SuppressAlert(ctx, tx, alert, duration); err != nil {
		return fmt.Errorf("Unable to suppress alert %d: %v", alert.Id, err)
	}
	// create a new supp rule to suppress any future similar alerts
	ents := models.Labels{"alert_name": alert.Name, "entity": alert.Entity}
	if alert.Device.Valid {
		ents["device"] = alert.Device.String
	}
	r := models.NewSuppRule(ents, models.MatchCond_ALL, reason, "alert_manager", duration)
	if _, err := h.AddSuppRule(ctx, tx, r); err != nil {
		return fmt.Errorf("Failed to suppress alert: %v", err)
	}
	tx.NewRecord(alert.Id, fmt.Sprintf("Alert Suppressed by %s for %v : %s", creator, duration, reason))
	h.notifyReceivers(alert, models.EventType_SUPPRESSED)
	return nil
}

func (h *AlertHandler) Clear(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	alert.Clear()
	if err := tx.Exec(models.QueryUpdateStatus, models.Status_CLEARED, alert.Id); err != nil {
		h.statDbError.Add(1)
		return err
	}
	tx.NewRecord(alert.Id, "Alert cleared")
	h.notifyReceivers(alert, models.EventType_CLEARED)
	return nil
}

// SetOwner sets the owner when an alert is acknowledged
func (h *AlertHandler) SetOwner(ctx context.Context, tx models.Txn, alert *models.Alert, name, teamName string) error {
	if teamName != "" && !h.Teams.Contains(teamName) {
		return fmt.Errorf("Team %s does not exist", teamName)
	}
	alert.SetOwner(name, teamName)
	if err := tx.UpdateAlert(alert); err != nil {
		h.statDbError.Add(1)
		return err
	}
	tx.NewRecord(alert.Id, fmt.Sprintf("Alert owner set to %s, team set to %s", name, teamName))
	// Notify all the receivers
	h.notifyReceivers(alert, models.EventType_ACKD)
	return nil
}

// AddSuppRule adds a new suppression rule into the suppressor
func (h *AlertHandler) AddSuppRule(ctx context.Context, tx models.Txn, rule *models.SuppressionRule) (int64, error) {
	return h.Suppressor.SaveRule(ctx, tx, rule)
}

// DeleteSuppRule deletes an existing suppression rule from the suppressor
func (h *AlertHandler) DeleteSuppRule(ctx context.Context, tx models.Txn, id int64) error {
	return h.Suppressor.DeleteRule(ctx, tx, id)
}
