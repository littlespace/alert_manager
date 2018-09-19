package models

import (
	"database/sql"
	"fmt"
	"github.com/golang/glog"
	"regexp"
	"time"
)

const (
	SuppType_DEVICE = iota
	SuppType_ALERT
	SuppType_ENTITY
	SuppType_SITE
	SuppType_REGION
)

var (
	QueryInsertRule = `INSERT INTO
    suppression_rules (
      name, type, alert_name, device, site, region, entity, created_at, duration, reason, creator
    ) VALUES (
    :name, :type, :alert_name, :device, :site, :region, :entity, :created_at, :duration,
    :reason, :creator
    )`

	QuerySelectActive = "SELECT * FROM suppression_rules WHERE (cast(extract(epoch from now()) as integer) - created_at) < duration"
)

type SuppressionRule struct {
	Id           int64
	Type         int
	Name         string
	AlertName    sql.NullString `db:"alert_name"`
	Device       sql.NullString
	Entity       sql.NullString
	Site, Region sql.NullString
	CreatedAt    MyTime `db:"created_at"`
	Duration     int64
	Reason       string
	Creator      string
}

func getDuration(duration string) int64 {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		// use default in case of error
		glog.V(4).Infof("Error parsing duration for suppression rule. Using default 4hours")
		dur = time.Duration(240 * time.Second)
	}
	return int64(dur.Seconds())
}

func NewSuppRuleForAlert(alert, creator, reason, duration string) SuppressionRule {
	return SuppressionRule{
		Name:      fmt.Sprintf("%s - %s - %v", alert, creator, duration),
		Type:      SuppType_ALERT,
		AlertName: sql.NullString{alert, true},
		CreatedAt: MyTime{time.Now()},
		Duration:  getDuration(duration),
		Reason:    reason,
		Creator:   creator,
	}
}

func NewSuppRuleForDevice(device, creator, reason, duration string) SuppressionRule {
	return SuppressionRule{
		Name:      fmt.Sprintf("%s - %s - %v", device, creator, duration),
		Type:      SuppType_DEVICE,
		Device:    sql.NullString{device, true},
		CreatedAt: MyTime{time.Now()},
		Duration:  getDuration(duration),
		Reason:    reason,
		Creator:   creator,
	}
}

func NewSuppRuleForEntity(device, entity, creator, reason, duration string) SuppressionRule {
	rule := NewSuppRuleForDevice(device, creator, reason, duration)
	rule.Type = SuppType_ENTITY
	rule.Entity = sql.NullString{entity, true}
	return rule
}

type SuppRules []SuppressionRule

func (s SuppRules) Find(params map[string]string) (SuppressionRule, bool) {
	var match bool
	for _, rule := range s {
		for k, v := range params {
			switch k {
			case "Device":
				if rule.Device.Valid {
					m, _ := regexp.MatchString(rule.Device.String, v)
					match = match || m
				}
			case "Entity":
				if rule.Entity.Valid {
					m, _ := regexp.MatchString(rule.Entity.String, v)
					match = match || m
				}
			case "Alert":
				if rule.AlertName.Valid {
					m, _ := regexp.MatchString(rule.AlertName.String, v)
					match = match || m
				}
			}
		}
		if match {
			return rule, true
		}
	}
	return SuppressionRule{}, false
}

func (tx *Tx) SelectRules(query string) (SuppRules, error) {
	var rules SuppRules
	err := tx.Select(&rules, query)
	return rules, err
}

func (tx *Tx) NewRule(rule *SuppressionRule) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertRule)
	err = stmt.Get(&newId, rule)
	return newId, err
}
