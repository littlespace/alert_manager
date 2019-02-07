package models

import (
	"fmt"
	"regexp"
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
      name, mcond, entities, created_at, duration, reason, creator, team
    ) VALUES (
    :name, :mcond, :entities, :created_at, :duration, :reason, :creator, :team
    ) RETURNING id`

	querySelectRules     = "SELECT * FROM suppression_rules_%s"
	QuerySelectActive    = " WHERE (cast(extract(epoch from now()) as integer) - created_at) < duration"
	queryUpdateRules     = "UPDATE suppression_rules"
	QueryDeleteSuppRules = "DELETE FROM suppression_rules WHERE id IN (?)"
)

func RulesQuery(query string) string {
	return fmt.Sprintf(querySelectRules, TeamName) + query
}

type SuppressionRule struct {
	Id         int64
	Mcond      MatchCondition
	Name       string
	Entities   Labels
	CreatedAt  MyTime `db:"created_at"`
	Duration   int64
	Reason     string
	Creator    string
	Team       string
	DontExpire bool
}

func (s SuppressionRule) Match(labels Labels) bool {
	switch s.Mcond {
	case MatchCond_ALL:
		for ek, ev := range s.Entities {
			lv, ok := labels[ek]
			if !ok {
				return false
			}
			var match bool
			if _, ok = lv.(string); ok {
				match, _ = regexp.MatchString(ev.(string), lv.(string))
			} else {
				match = lv == ev
			}
			if !match {
				return false
			}
		}
		return true
	case MatchCond_ANY:
		for ek, ev := range s.Entities {
			lv, ok := labels[ek]
			if !ok {
				continue
			}
			var match bool
			if _, ok = lv.(string); ok {
				match, _ = regexp.MatchString(ev.(string), lv.(string))
			} else {
				match = lv == ev
			}
			if match {
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

func NewSuppRule(entities Labels, mcond MatchCondition, reason, creator string, duration time.Duration) *SuppressionRule {
	return &SuppressionRule{
		Name:      fmt.Sprintf("Rule - %s - %v", creator, duration),
		Mcond:     mcond,
		CreatedAt: MyTime{time.Now()},
		Duration:  int64(duration.Seconds()),
		Reason:    reason,
		Creator:   creator,
		Entities:  entities,
		Team:      TeamName,
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
