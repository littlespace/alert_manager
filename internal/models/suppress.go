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

	QuerySelectActive = "SELECT * FROM suppression_rules WHERE (cast(extract(epoch from now()) as integer) - created_at) < duration"
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

func NewSuppRule(entities Labels, rtype, reason, creator string, duration time.Duration) SuppressionRule {
	return SuppressionRule{
		Name:      fmt.Sprintf("Rule - %s - %v", creator, duration),
		Rtype:     typeMap[rtype],
		CreatedAt: MyTime{time.Now()},
		Duration:  int64(duration.Seconds()),
		Reason:    reason,
		Creator:   creator,
		Entities:  entities,
	}
}

type SuppRules []SuppressionRule

func (s SuppRules) Find(labels Labels) (SuppressionRule, bool) {
	// a match is successful when all entities in rule are matched in the alert labels
	for _, rule := range s {
		match := true
		for ek, ev := range rule.Entities {
			match = match && ev == labels[ek]
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

func (tx *Tx) NewSuppRule(rule *SuppressionRule) (int64, error) {
	var newId int64
	stmt, err := tx.PrepareNamed(QueryInsertRule)
	if err != nil {
		return newId, err
	}
	err = stmt.Get(&newId, rule)
	return newId, err
}
