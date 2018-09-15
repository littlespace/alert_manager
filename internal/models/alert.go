package models

import (
	"database/sql"
	"encoding/json"
	"github.com/golang/glog"
	"strings"
	"time"
)

var (
	QueryInsertNew = `INSERT INTO 
    alerts (
      name, description, entity, external_id, source, device, site, owner, team, tags, start_time, last_active,
      agg_id, auto_expire, auto_clear, expire_after, severity, status, metadata, scope, is_aggregate
    ) VALUES (
      :name, :description, :entity, :external_id, :source, :device, :site, :owner, :team, :tags,
      :start_time, :last_active, :agg_id, :auto_expire, :auto_clear, :expire_after,
      :severity, :status, :metadata, :scope, :is_aggregate
    ) RETURNING id`

	QueryUpdateAlertById = `UPDATE alerts SET
    name=:name, description=:description, entity=:entity, external_id=:external_id, source=:source,
    device=:device, site=:site, owner=:owner, team=:team, tags=:tags, start_time=:start_time,
    last_active=:last_active, agg_id=:agg_id, auto_expire=:auto_expire, auto_clear=:auto_clear,
    expire_after=:expire_after, severity=:severity, status=:status, metadata=:metadata, scope=:scope,
    is_aggregate=:is_aggregate
      WHERE id=:id`

	queryUpdate           = "UPDATE alerts"
	QueryUpdateLastActive = queryUpdate + " SET last_active=? WHERE id IN (?)"
	QueryUpdateAggId      = queryUpdate + " SET agg_id=? WHERE id IN (?)"

	querySelect             = "SELECT * from alerts"
	QuerySelectByNames      = querySelect + " WHERE name IN (?) AND status=1 AND agg_id IS NULL FOR UPDATE"
	QuerySelectById         = querySelect + " WHERE id=$1 FOR UPDATE"
	QuerySelectByStatus     = querySelect + " WHERE status IN (?) FOR UPDATE"
	QuerySelectNoOwner      = querySelect + " WHERE owner is NULL AND status=1 FOR UPDATE"
	QuerySelectByNameEntity = querySelect + " WHERE name=$1 AND entity=$2 AND status=1 FOR UPDATE"
	QuerySelectByDevice     = querySelect + " WHERE name=$1 AND entity=$2 AND device=$3 AND status=1 FOR UPDATE"
	QuerySelectTags         = "SELECT tags from alerts WHERE id=$1"
	QuerySelectExpired      = querySelect + ` WHERE
    status=1 AND auto_expire AND (cast(extract(epoch from now()) as integer) - last_active) > expire_after FOR UPDATE`
	QuerySelectAllAggregated = querySelect + " WHERE agg_id IN (SELECT id from alerts WHERE is_aggregate AND status = 1)"
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
	Tags         sql.NullString // TODO maybe store in separate table
	StartTime    MyTime         `db:"start_time"`
	LastActive   MyTime         `db:"last_active"`
	AutoExpire   bool           `db:"auto_expire"`
	AutoClear    bool           `db:"auto_clear"`
	AggregatorId sql.NullInt64  `db:"agg_id"`
	IsAggregate  bool           `db:"is_aggregate"`
	ExpireAfter  sql.NullInt64  `db:"expire_after"`
	Severity     AlertSeverity
	Status       AlertStatus
	Metadata     sql.NullString // json encoded metadata
}

// custom Marshaler interface for Alert
func (a Alert) MarshalJSON() ([]byte, error) {
	tmp := struct {
		Id                                       int64
		ExternalId                               string `json:"external_id"`
		Name, Description, Entity, Source, Scope string
		Device, Site, Owner, Team, Tags          string
		StartTime                                int64 `json:"start_time"`
		LastActive                               int64 `json:"last_active"`
		AggregatorId                             int64 `json:"agg_id"`
		IsAggregate                              bool  `json:"is_aggregate"`
		Severity                                 string
		Status                                   string
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
		Tags:         a.Tags.String,
		StartTime:    a.StartTime.Unix(),
		LastActive:   a.LastActive.Unix(),
		AggregatorId: a.AggregatorId.Int64,
		IsAggregate:  a.IsAggregate,
		Severity:     a.Severity.String(),
		Status:       a.Status.String(),
	}
	return json.Marshal(&tmp)
}

func NewAlert(name, description, entity, source, scope string, extId string, startTime time.Time, sev string, isAgg bool) *Alert {
	return &Alert{
		Status:      Status_ACTIVE,
		ExternalId:  extId,
		Name:        name,
		Description: description,
		Entity:      entity,
		Source:      source,
		Scope:       scope,
		StartTime:   MyTime{startTime},
		LastActive:  MyTime{time.Now()},
		Severity:    SevMap[sev],
		AutoClear:   true,
		AutoExpire:  false,
		IsAggregate: isAgg,
	}
}

func (a *Alert) AddDevice(device string) {
	a.Device = sql.NullString{device, true}
}

func (a *Alert) AddSite(site string) {
	a.Site = sql.NullString{site, true}
}

func (a *Alert) AddTags(tags ...string) {
	t := []string{}
	t = append(t, tags...)
	tagString := strings.Join(t, ",")
	a.Tags = sql.NullString{tagString, true}
}

func (a *Alert) HasTags(tags ...string) bool {
	hasTags := true
	if !a.Tags.Valid {
		return false
	}
	t := strings.Split(a.Tags.String, ",")
	for _, tag := range tags {
		var found bool
		for _, tt := range t {
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
	glog.V(2).Infof("Suppressing %s for %v", a.Name, suppressFor)
}

func (a *Alert) Unsuppress() {
	glog.V(2).Infof("Un-suppressing %s", a.Name)
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

func (a *Alert) SetAggId(id int64) {
	a.AggregatorId = sql.NullInt64{id, true}
}

func (a *Alert) Clear() {
	glog.V(2).Infof("Clearing out alert %d", a.Id)
	a.Status = Status_CLEARED
}

func (a *Alert) AddMeta(meta interface{}) error {
	// store alert meta as json encoded string
	m, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	a.Metadata = sql.NullString{string(m), true}
	return nil
}

type Alerts []Alert

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

func (tx *Tx) UpdateAlert(alert *Alert) error {
	_, err := tx.NamedExec(QueryUpdateAlertById, alert)
	return err
}

func (tx *Tx) NewAlert(alert *Alert) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertNew)
	err = stmt.Get(&newId, alert)
	return newId, err
}

func (tx *Tx) GetAlert(query string, args ...interface{}) (*Alert, error) {
	alert := &Alert{}
	if err := tx.Get(alert, query, args...); err != nil {
		return nil, err
	}
	return alert, nil
}

func (tx *Tx) SelectAlerts(query string) (Alerts, error) {
	var alerts Alerts
	err := tx.Select(&alerts, query)
	return alerts, err
}
