package handler

import (
	"context"
	"fmt"
	"sort"
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

// AlertHandler handles common alert operations such as expiry, suppression etc.
// It also sends alerts to interested receivers
type AlertHandler struct {
	// db handler
	Db                 models.Dbase
	Suppressor         *suppressor
	Teams              models.Teams
	procChan           chan *models.AlertEvent
	statTransformError stats.Stat
	statDbError        stats.Stat
}

// NewHandler returns a new alert handler which uses the supplied db
func NewHandler(db models.Dbase) *AlertHandler {
	h := &AlertHandler{
		Db:                 db,
		Suppressor:         GetSuppressor(db),
		procChan:           make(chan *models.AlertEvent),
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
					return h.handleClear(ctx, tx, alert)
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
	var labels models.Labels
	existingAlert, _ := h.GetExisting(tx, alert)
	if existingAlert == nil {
		glog.V(2).Infof("No existing alert found for %s:%s:%s", alert.Name, alert.Device.String, alert.Entity)
		// add transforms
		alert.ExtendLabels()
		h.applyTransforms(alert)
		labels = alert.Labels
	} else {
		existingAlert.ExtendLabels()
		h.applyTransforms(existingAlert)
		labels = existingAlert.Labels
	}
	// check if alert matches an existing suppression rule based on alert labels
	if rule := h.Suppressor.Match(labels); rule != nil && rule.TimeLeft() > 0 {
		glog.V(2).Infof("Found matching suppression rule for %s:%s:%s: %d:%s", alert.Name, alert.Entity, alert.Device.String, rule.Id, rule.Name)
		return nil
	}
	if existingAlert != nil {
		return h.reactivateAlert(tx, existingAlert)
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

func (h *AlertHandler) handleClear(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	// clear existing alert if auto clear is true
	existingAlert, err := h.GetExisting(tx, alert)
	if err != nil {
		glog.V(2).Infof("No existing alert found for %s:%s to clear", alert.Name, alert.Entity)
		return nil
	}
	if existingAlert.Status == models.Status_CLEARED {
		// already cleared
		return nil
	}
	if !existingAlert.AutoClear {
		glog.V(2).Infof("Not auto-clearing alert %d ", existingAlert.Id)
		return nil
	}
	// dont clear acknowledged alerts if the config says so
	if config, ok := Config.GetAlertConfig(existingAlert.Name); ok && config.Config.DontClearAcknowledged && existingAlert.Owner.Valid {
		glog.V(4).Infof("Not clearing ack'd alert: %d", existingAlert.Id)
		return nil
	}
	return h.clearAlert(ctx, tx, existingAlert)
}

func (h *AlertHandler) clearAlert(ctx context.Context, tx models.Txn, alert *models.Alert) error {
	err := h.Clear(ctx, tx, alert, true)
	if err != nil {
		h.statDbError.Add(1)
		return fmt.Errorf("Cant clear existing alert %d: %v", alert.Id, err)
	}
	return nil
}

func (h *AlertHandler) GetExisting(tx models.Txn, alert *models.Alert) (*models.Alert, error) {
	var existing *models.Alert
	var err error
	// an alert is assumed to be uniquely identified by its Id or by its Name:Device:Entity
	if alert.Id > 0 {
		existing, err = tx.GetAlert(models.QuerySelectById, alert.Id)
	} else {
		query := models.QuerySelectByNameEntity
		devQuery := models.QuerySelectByDevice
		config, ok := Config.GetAlertConfig(alert.Name)
		// if disable_dedup is not true in config, only check for currently active alerts.
		if ok && config.Config.DisableDedup {
			query = models.QueryActiveByNameEntity
			devQuery = models.QueryActiveByDevice
		}
		if alert.Device.Valid {
			existing, err = tx.GetAlert(devQuery, alert.Name, alert.Entity, alert.Device.String)
		} else {
			existing, err = tx.GetAlert(query, alert.Name, alert.Entity)
		}
	}
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (h *AlertHandler) reactivateAlert(tx models.Txn, existingAlert *models.Alert) error {
	// reactivate the alert and the agg alert if applicable and extend the expiry time if alert already exists
	toUpdate := models.Alerts{existingAlert}
	toNotify := existingAlert
	if existingAlert.AggregatorId != 0 {
		agg, err := tx.GetAlert(models.QuerySelectById, existingAlert.AggregatorId)
		if err != nil {
			return fmt.Errorf("Failed to get alert %d: %v", existingAlert.AggregatorId, err)
		}
		toNotify = agg
		toUpdate = append(toUpdate, agg)
	}
	newLastActive := models.MyTime{time.Now()}
	alreadyActive := existingAlert.Status == models.Status_ACTIVE || existingAlert.Status == models.Status_SUPPRESSED
	for _, a := range toUpdate {
		a.LastActive = newLastActive
		if !alreadyActive {
			glog.V(4).Infof("Reactivating old alert %d", a.Id)
			a.Status = models.Status_ACTIVE
			a.StartTime = newLastActive
		}
		if err := tx.UpdateAlert(a); err != nil {
			h.statDbError.Add(1)
			return fmt.Errorf("Failed update alert %d: %v", a.Id, err)
		}
	}
	if alreadyActive {
		return nil
	}
	tx.NewRecord(existingAlert.Id, "Alert re-activated")
	if existingAlert.AggregatorId != 0 {
		tx.NewRecord(existingAlert.AggregatorId, fmt.Sprintf("Alert re-activated due to component alert %d", existingAlert.Id))
	}
	toNotify.Status = models.Status_ACTIVE
	h.notifyReceivers(toNotify, models.EventType_ACTIVE)
	return nil
}

func (h *AlertHandler) applyTransforms(alert *models.Alert) {
	// apply transforms in order of priority. Lower == first
	var toApply []Transform
	for _, transform := range Transforms {
		rule, ok := Config.GetTransformRule(transform.Name())
		if !ok {
			// if no rule matches defined, pass all alerts
			toApply = append(toApply, transform)
			continue
		}
		for _, m := range rule.Matches {
			if m.MatchAll(alert.Labels) {
				toApply = append(toApply, transform)
				break
			}
		}
	}
	sort.Slice(toApply, func(i, j int) bool {
		return toApply[i].GetPriority() < toApply[j].GetPriority()
	})
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("PANIC while applying transform to alert %s: %v", alert.Name, r)
			h.statTransformError.Add(1)
		}
	}()
	for _, xform := range toApply {
		glog.V(2).Infof("Applying Transform: %s to alert %s", xform.Name(), alert.Name)
		if err := xform.Apply(alert); err != nil {
			glog.Errorf("Failed to apply transform %s to alert %s: %v, retrying..", xform.Name(), alert.Name, err)
			if err := xform.Apply(alert); err != nil {
				glog.Errorf("Failed to apply transform %s to alert %s: %v", xform.Name(), alert.Name, err)
				h.statTransformError.Add(1)
			}
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

func (h *AlertHandler) Suppress(
	ctx context.Context,
	tx models.Txn,
	alert *models.Alert,
	creator, reason string,
	duration time.Duration,
	notify bool,
) error {
	if alert.IsAggregate {
		var grouped models.Alerts
		if err := tx.InSelect(models.QuerySelectByAggId, &grouped, alert.Id); err != nil {
			return fmt.Errorf("Failed to query alerts: %v", err)
		}
		// suppress individual alerts
		for _, a := range grouped {
			cReason := "Alert suppressed due to aggregated alert suppressed"
			if err := h.Suppress(ctx, tx, a, creator, cReason, duration, notify); err != nil {
				return fmt.Errorf("Unable to suppress alert %d: %v", a.Id, err)
			}
		}
		tx.NewRecord(alert.Id, fmt.Sprintf("Alert Suppressed by %s for %v : %s", creator, duration, reason))
		h.notifyReceivers(alert, models.EventType_SUPPRESSED)
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
	if notify {
		h.notifyReceivers(alert, models.EventType_SUPPRESSED)
	}
	return nil
}

func (h *AlertHandler) Clear(ctx context.Context, tx models.Txn, alert *models.Alert, notify bool) error {
	alert.Clear()
	if err := tx.Exec(models.QueryUpdateStatus, models.Status_CLEARED, alert.Id); err != nil {
		h.statDbError.Add(1)
		return err
	}
	tx.NewRecord(alert.Id, "Alert cleared")
	if notify {
		h.notifyReceivers(alert, models.EventType_CLEARED)
	}
	return nil
}

// SetOwner sets the owner when an alert is acknowledged
func (h *AlertHandler) SetOwner(ctx context.Context, tx models.Txn, alert *models.Alert, name, teamName string, notify bool) error {
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
	if notify {
		h.notifyReceivers(alert, models.EventType_ACKD)
	}
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

// Escalate bumps up alert severity
func (h *AlertHandler) Escalate(ctx context.Context, tx models.Txn, alert *models.Alert, newSev models.AlertSeverity, notify bool) error {
	alert.Severity = newSev
	tx.NewRecord(alert.Id, fmt.Sprintf("Alert severity escalated to %s", newSev.String()))
	if notify {
		h.notifyReceivers(alert, models.EventType_ESCALATED)
	}
	return nil
}
