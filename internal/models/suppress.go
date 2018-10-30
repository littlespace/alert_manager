package models

import (
	"fmt"
	"time"
)

const (
	SuppType_DEVICE = iota
	SuppType_ALERT
	SuppType_ENTITY
)

var typeMap = map[string]int{
	"device": SuppType_DEVICE,
	"alert":  SuppType_ALERT,
	"entity": SuppType_ENTITY}

var (
	QueryInsertRule = `INSERT INTO
    suppression_rules (
      name, rtype, entities, created_at, duration, reason, creator
    ) VALUES (
    :name, :rtype, :entities, :created_at, :duration, :reason, :creator
    ) RETURNING id`

	querySelectRules      = "SELECT * FROM suppression_rules"
	QuerySelectActive     = querySelectRules + " WHERE (cast(extract(epoch from now()) as integer) - created_at) < duration"
	QuerySelectAlertRules = `SELECT DISTINCT ON ((entities->>'alert_id')::int) *
		FROM suppression_rules
		WHERE  (entities->>'alert_id')::int IN (?) AND
		rtype = 1 AND creator = 'alert_manager' 
		ORDER BY ((entities->>'alert_id')::int), created_at DESC;`

	queryUpdateRules = "UPDATE suppression_rules"

	QueryDeleteSuppRules = "DELETE FROM suppression_rules WHERE id IN (?)"
)

type SuppressionRule struct {
	Id         int64
	Rtype      int
	Name       string
	Entities   Labels
	CreatedAt  MyTime `db:"created_at"`
	Duration   int64
	Reason     string
	Creator    string
	DontExpire bool
}

type MatchCondition int

const (
	MatchCond_ALL MatchCondition = 1
	MatchCond_ANY MatchCondition = 2
)

func (s SuppressionRule) Match(labels Labels, cond MatchCondition) bool {
	switch cond {
	case MatchCond_ALL:
		for lk, lv := range labels {
			ev, ok := s.Entities[lk]
			if !ok || ev != lv {
				return false
			}
		}
		return true
	case MatchCond_ANY:
		for lk, lv := range labels {
			ev, ok := s.Entities[lk]
			if ok && ev == lv {
				return true
			}
		}
	}
	return false
}

func (s SuppressionRule) TimeLeft() time.Duration {
	if s.DontExpire {
		return time.Duration(s.Duration) * time.Second
	}
	return s.CreatedAt.Add(time.Duration(s.Duration) * time.Second).Sub(time.Now())
}

func NewSuppRule(entities Labels, rtype, reason, creator string, duration time.Duration) *SuppressionRule {
	return &SuppressionRule{
		Name:      fmt.Sprintf("Rule - %s - %v", creator, duration),
		Rtype:     typeMap[rtype],
		CreatedAt: MyTime{time.Now()},
		Duration:  int64(duration.Seconds()),
		Reason:    reason,
		Creator:   creator,
		Entities:  entities,
	}
}

type SuppRules []*SuppressionRule

func (tx *Tx) SelectRules(query string, args ...interface{}) (SuppRules, error) {
	var rules SuppRules
	err := tx.Select(&rules, query, args...)
	return rules, err
}

func (tx *Tx) NewSuppRule(rule *SuppressionRule) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertRule)
	if err != nil {
		return newId, err
	}
	err = stmt.Get(&newId, rule)
	return newId, err
}
