package models

import (
	"database/sql"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/lib/pq"
	"time"
)

var (
	QueryInsertNew = `INSERT INTO 
    alerts (
      name, description, entity, external_id, source, device, site, owner, team, tags, start_time, last_active,
      agg_id, auto_expire, auto_clear, expire_after, severity, status, labels, scope, is_aggregate
    ) VALUES (
      :name, :description, :entity, :external_id, :source, :device, :site, :owner, :team, :tags,
      :start_time, :last_active, :agg_id, :auto_expire, :auto_clear, :expire_after,
      :severity, :status, :labels, :scope, :is_aggregate
    ) RETURNING id`

	QueryUpdateAlertById = `UPDATE alerts SET
    name=:name, description=:description, entity=:entity, external_id=:external_id, source=:source,
    device=:device, site=:site, owner=:owner, team=:team, tags=:tags, start_time=:start_time,
    last_active=:last_active, agg_id=:agg_id, auto_expire=:auto_expire, auto_clear=:auto_clear,
    expire_after=:expire_after, severity=:severity, status=:status, labels=:labels, scope=:scope,
    is_aggregate=:is_aggregate
      WHERE id=:id`

	queryUpdateAlerts     = "UPDATE alerts"
	QueryUpdateLastActive = queryUpdateAlerts + " SET last_active=? WHERE id IN (?)"
	QueryUpdateAggId      = queryUpdateAlerts + " SET agg_id=? WHERE id IN (?)"
	QueryUpdateStatus     = queryUpdateAlerts + " SET status=$1 WHERE id=$2 OR id IN (SELECT id from alerts WHERE agg_id=$2)"

	querySelectAlerts       = "SELECT * from alerts"
	QuerySelectByNames      = querySelectAlerts + " WHERE name IN (?) AND status=1 AND agg_id IS NULL FOR UPDATE"
	QuerySelectById         = querySelectAlerts + " WHERE id=$1 FOR UPDATE"
	QuerySelectByIds        = querySelectAlerts + " WHERE id IN (?) ORDER BY id FOR UPDATE"
	QuerySelectByStatus     = querySelectAlerts + " WHERE status IN (?) ORDER BY id FOR UPDATE"
	QuerySelectNoOwner      = querySelectAlerts + " WHERE owner is NULL AND status=1 ORDER BY id FOR UPDATE"
	QuerySelectByNameEntity = querySelectAlerts + " WHERE name=$1 AND entity=$2 AND status IN (1,2) FOR UPDATE"
	QuerySelectByDevice     = querySelectAlerts + " WHERE name=$1 AND entity=$2 AND device=$3 AND status IN (1,2) FOR UPDATE"
	QuerySelectExpired      = querySelectAlerts + ` WHERE
    status IN (1,2) AND auto_expire AND (cast(extract(epoch from now()) as integer) - last_active) > expire_after ORDER BY id FOR UPDATE`
	QuerySelectAllAggregated = querySelectAlerts + " WHERE agg_id IN (SELECT id from alerts WHERE is_aggregate AND status = 1)"
	QuerySelectSuppressed    = querySelectAlerts + ` WHERE status=2 AND id IN (
    select (entities->>'alert_id')::int from suppression_rules where rtype = 1 AND
    creator = 'alert_manager' AND
    (cast(extract(epoch from now()) as integer) - created_at) < duration
  )`
)

type AlertSeverity int

func (s AlertSeverity) String() string {
	for name, sev := range SevMap {
		if sev == s {
			return name
		}
	}
	return "UNKNOWN"
}

type AlertStatus int

func (s AlertStatus) String() string {
	for name, status := range StatusMap {
		if status == s {
			return name
		}
	}
	return "UNKNOWN"
}

const (
	Sev_CRITICAL AlertSeverity = 1
	Sev_WARN     AlertSeverity = 2
	Sev_INFO     AlertSeverity = 3

	Status_ACTIVE     AlertStatus = 1
	Status_SUPPRESSED AlertStatus = 2
	Status_EXPIRED    AlertStatus = 3
	Status_CLEARED    AlertStatus = 4
)

var (
	SevMap    = map[string]AlertSeverity{"CRITICAL": Sev_CRITICAL, "WARN": Sev_WARN, "INFO": Sev_INFO}
	StatusMap = map[string]AlertStatus{"ACTIVE": Status_ACTIVE, "SUPPRESSED": Status_SUPPRESSED, "EXPIRED": Status_EXPIRED, "CLEARED": Status_CLEARED}
)

type Alert struct {
	Id           int64
	ExternalId   string `db:"external_id"`
	Name         string
	Description  string
	Entity       string
	Source       string
	Scope        string
	Device       sql.NullString
	Site         sql.NullString
	Owner        sql.NullString
	Team         sql.NullString
	Tags         pq.StringArray
	StartTime    MyTime        `db:"start_time"`
	LastActive   MyTime        `db:"last_active"`
	AutoExpire   bool          `db:"auto_expire"`
	AutoClear    bool          `db:"auto_clear"`
	AggregatorId int64         `db:"agg_id"`
	IsAggregate  bool          `db:"is_aggregate"`
	ExpireAfter  sql.NullInt64 `db:"expire_after"`
	Severity     AlertSeverity
	Status       AlertStatus
	Labels       Labels // json encoded k-v labels
	History      []*Record
}

// custom Marshaler interface for Alert
func (a Alert) MarshalJSON() ([]byte, error) {
	tmp := struct {
		Id                                       int64
		ExternalId                               string `json:"external_id"`
		Name, Description, Entity, Source, Scope string
		Device, Site, Owner, Team                string
		Tags                                     pq.StringArray
		StartTime                                int64 `json:"start_time"`
		LastActive                               int64 `json:"last_active"`
		AggregatorId                             int64 `json:"agg_id"`
		IsAggregate                              bool  `json:"is_aggregate"`
		Severity                                 string
		Status                                   string
		History                                  []struct {
			Timestamp int64
			Event     string
		}
	}{
		Id:           a.Id,
		ExternalId:   a.ExternalId,
		Name:         a.Name,
		Description:  a.Description,
		Entity:       a.Entity,
		Source:       a.Source,
		Scope:        a.Scope,
		Device:       a.Device.String,
		Site:         a.Site.String,
		Owner:        a.Owner.String,
		Team:         a.Team.String,
		Tags:         a.Tags,
		StartTime:    a.StartTime.Unix(),
		LastActive:   a.LastActive.Unix(),
		AggregatorId: a.AggregatorId,
		IsAggregate:  a.IsAggregate,
		Severity:     a.Severity.String(),
		Status:       a.Status.String(),
	}
	for _, h := range a.History {
		tmp.History = append(tmp.History, struct {
			Timestamp int64
			Event     string
		}{h.Timestamp.Unix(), h.Event},
		)
	}
	return json.Marshal(&tmp)
}

func NewAlert(name, description, entity, source, scope string, extId string, startTime time.Time, sev string, isAgg bool) *Alert {
	// sanity checks - if alert > 10 min old, set start time to now
	if time.Now().Sub(startTime) > time.Duration(10*time.Minute) {
		startTime = time.Now()
	}
	return &Alert{
		Status:      Status_ACTIVE,
		ExternalId:  extId,
		Name:        name,
		Description: description,
		Entity:      entity,
		Source:      source,
		Scope:       scope,
		StartTime:   MyTime{startTime},
		LastActive:  MyTime{startTime},
		Severity:    SevMap[sev],
		AutoClear:   true,
		AutoExpire:  false,
		IsAggregate: isAgg,
		Labels:      make(Labels),
	}
}

func (a *Alert) AddDevice(device string) {
	a.Device = sql.NullString{device, true}
}

func (a *Alert) AddSite(site string) {
	a.Site = sql.NullString{site, true}
}

func (a *Alert) AddTags(tags ...string) {
	t := []string(tags)
	a.Tags = pq.StringArray(t)
}

func (a *Alert) HasTags(tags ...string) bool {
	hasTags := true
	for _, tag := range tags {
		var found bool
		for _, tt := range a.Tags {
			if tt == tag {
				found = true
				break
			}
		}
		hasTags = hasTags && found
	}
	return hasTags
}

func (a *Alert) SetAutoExpire(duration time.Duration) {
	a.AutoExpire = true
	a.ExpireAfter = sql.NullInt64{int64(duration.Seconds()), true}
}

func (a *Alert) Suppress(suppressFor time.Duration) {
	a.Status = Status_SUPPRESSED
	glog.V(2).Infof("Suppressing %d:%s for %v", a.Id, a.Name, suppressFor)
}

func (a *Alert) Unsuppress() {
	glog.V(2).Infof("Un-suppressing %d:%s", a.Id, a.Name)
	a.Status = Status_ACTIVE
}

func (a *Alert) SetOwner(name, team string) {
	glog.V(2).Infof("Setting alert %d owner to %s:%s", a.Id, name, team)
	a.Owner = sql.NullString{name, true}
	a.Team = sql.NullString{team, true}
}

func (a *Alert) SetSeverity(sev AlertSeverity) {
	glog.V(2).Infof("Setting alert %d Severity to %v", a.Id, sev)
	a.Severity = sev
}

func (a *Alert) Clear() {
	glog.V(2).Infof("Clearing out alert %d", a.Id)
	a.Status = Status_CLEARED
}

type Alerts []*Alert

func (a Alerts) AnyStatusIn(status AlertStatus) bool {
	for _, alert := range a {
		if alert.Status == status {
			return true
		}
	}
	return false
}

func (a Alerts) AllStatusIn(status AlertStatus) bool {
	for _, alert := range a {
		if alert.Status != status {
			return false
		}
	}
	return true
}

func (a Alerts) AllCleared() bool {
	return a.AllStatusIn(Status_CLEARED)
}

func (a Alerts) AllExpired() bool {
	return a.AllStatusIn(Status_EXPIRED)
}

func (a Alerts) AllInactive() bool {
	for _, al := range a {
		if al.Status == Status_ACTIVE || al.Status == Status_SUPPRESSED {
			return false
		}
	}
	return true
}

func (tx *Tx) UpdateAlert(alert *Alert) error {
	_, err := tx.NamedExec(QueryUpdateAlertById, alert)
	return err
}

func (tx *Tx) NewAlert(alert *Alert) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertNew)
	if err != nil {
		return newId, err
	}
	err = stmt.Get(&newId, alert)
	return newId, err
}

func (tx *Tx) GetAlert(query string, args ...interface{}) (*Alert, error) {
	alert := &Alert{}
	query = tx.Rebind(query)
	if err := tx.Get(alert, query, args...); err != nil {
		return nil, err
	}
	return alert, nil
}

func (tx *Tx) SelectAlerts(query string, args ...interface{}) (Alerts, error) {
	var alerts Alerts
	err := tx.Select(&alerts, query, args...)
	return alerts, err
}

func (tx *Tx) SelectAlertsWithHistory(query string, args ...interface{}) (Alerts, error) {
	alerts, err := tx.SelectAlerts(query, args...)
	if err != nil {
		return Alerts{}, err
	}
	if err := tx.AddAlertHistory(alerts); err != nil {
		return alerts, err
	}
	return alerts, nil
}
