package models

import (
	"fmt"
	"time"
)

type MatchCondition int

const (
	MatchCond_ALL MatchCondition = 1
	MatchCond_ANY MatchCondition = 2
)

var CondMap = map[string]MatchCondition{"all": MatchCond_ALL, "any": MatchCond_ANY}

var (
	QueryInsertRule = `INSERT INTO
    suppression_rules (
      name, mcond, entities, created_at, duration, reason, creator
    ) VALUES (
    :name, :mcond, :entities, :created_at, :duration, :reason, :creator
    ) RETURNING id`

	querySelectRules     = "SELECT * FROM suppression_rules"
	QuerySelectActive    = querySelectRules + " WHERE (cast(extract(epoch from now()) as integer) - created_at) < duration"
	queryUpdateRules     = "UPDATE suppression_rules"
	QueryDeleteSuppRules = "DELETE FROM suppression_rules WHERE id IN (?)"
)

type SuppressionRule struct {
	Id         int64
	Mcond      MatchCondition
	Name       string
	Entities   Labels
	CreatedAt  MyTime `db:"created_at"`
	Duration   int64
	Reason     string
	Creator    string
	DontExpire bool
}

func (s SuppressionRule) Match(labels Labels) bool {
	switch s.Mcond {
	case MatchCond_ALL:
		return s.Entities.MatchAll(labels)
	case MatchCond_ANY:
		return s.Entities.MatchAny(labels)
	}
	return false
}

func (s SuppressionRule) TimeLeft() time.Duration {
	if s.DontExpire {
		return time.Duration(s.Duration) * time.Second
	}
	return s.CreatedAt.Add(time.Duration(s.Duration) * time.Second).Sub(time.Now())
}

func NewSuppRule(entities Labels, mcond MatchCondition, reason, creator string, duration time.Duration) *SuppressionRule {
	return &SuppressionRule{
		Name:      fmt.Sprintf("Rule - %s - %v", creator, duration),
		Mcond:     mcond,
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
