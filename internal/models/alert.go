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
      name, description, entity, external_id, source, device, owner, team, tags, start_time, last_active,
      agg_id, auto_expire, auto_clear, expire_after, severity, status, metadata, scope
    ) VALUES (
      :name, :description, :entity, :external_id, :source, :device, :owner, :team, :tags,
      :start_time, :last_active, :agg_id, :auto_expire, :auto_clear, :expire_after,
      :severity, :status, :metadata, :scope
    ) RETURNING id`

	QueryUpdateAlertById = `UPDATE alerts SET
    name=:name, description=:description, entity=:entity, external_id=:external_id, source=:source,
    device=:device, owner=:owner, team=:team, tags=:tags, start_time=:start_time, last_active=:last_active,
    agg_id=:agg_id, auto_expire=:auto_expire, auto_clear=:auto_clear, expire_after=:expire_after,
    severity=:severity, status=:status, metadata=:metadata, scope=:scope
      WHERE id=:id`

	QueryUpdateOwnerById  = "UPDATE alerts SET owner=$1, team=$2 WHERE id=$3"
	QueryUpdateStatusById = "UPDATE alerts SET status=$1 WHERE id=$2"
	QueryUpdateSevById    = "UPDATE alerts SET severity=$1 WHERE id=$2"
	QueryUpdateLastActive = "UPDATE alerts SET last_active=$1 WHERE id=$2"
	QueryUpdateAggId      = "UPDATE alerts SET agg_id=$1 WHERE id in ($2)"

	querySelect             = "SELECT * from alerts"
	QuerySelectByName       = querySelect + " WHERE name=$1 and status=1 FOR UPDATE"
	QuerySelectById         = querySelect + " WHERE id=$1 FOR UPDATE"
	QuerySelectByStatus     = querySelect + " WHERE status IN ($1) FOR UPDATE"
	QuerySelectNoOwner      = querySelect + " WHERE owner is NULL AND status=1 FOR UPDATE"
	QuerySelectByNameEntity = querySelect + " WHERE name=$1 AND entity=$2 AND status=1 FOR UPDATE"
	QuerySelectByDevice     = querySelect + " WHERE name=$1 AND entity=$2 AND device=$3 AND status=1 FOR UPDATE"
	QuerySelectTags         = "SELECT tags from alerts WHERE id=$1"
	QuerySelectExpired      = querySelect + ` WHERE
    status=1 AND auto_expire AND (cast(extract(epoch from now()) as integer) - last_active) > expire_after`
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
	Owner        sql.NullString
	Team         sql.NullString
	Tags         sql.NullString // TODO maybe store in separate table
	StartTime    MyTime         `db:"start_time"`
	LastActive   MyTime         `db:"last_active"`
	AutoExpire   bool           `db:"auto_expire"`
	AutoClear    bool           `db:"auto_clear"`
	AggregatorId sql.NullInt64  `db:"agg_id"`
	ExpireAfter  sql.NullInt64  `db:"expire_after"`
	Severity     AlertSeverity
	Status       AlertStatus
	Metadata     sql.NullString // json encoded metadata
}

func NewAlert(name, description, entity, source, scope string, extId string, startTime time.Time, sev string) *Alert {
	var severity AlertSeverity
	switch sev {
	case "CRITICAL":
		severity = Sev_CRITICAL
	case "WARN":
		severity = Sev_WARN
	case "INFO":
		severity = Sev_INFO
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
		LastActive:  MyTime{time.Now()},
		Severity:    severity,
		AutoClear:   true,
		AutoExpire:  false,
	}
}

func (a *Alert) AddDevice(device string) {
	a.Device = sql.NullString{device, true}
}

func (a *Alert) AddTags(tags ...string) {
	t := []string{}
	t = append(t, tags...)
	tagString := strings.Join(t, ",")
	a.Tags = sql.NullString{tagString, true}
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
