package models

import (
	"fmt"
)

type Op int

var (
	Op_EQUAL Op = 1
	Op_IN    Op = 2
)

type Param struct {
	Field  string
	Values []string
	Op     Op
}

type Querier interface {
	Run(tx Txn) (Alerts, error)
	toSQL() string
}

type Query struct {
	Limit  int
	Offset int
	Params []Param
}

func (q Query) toSQL() string {
	if len(q.Params) == 0 {
		return querySelect
	}
	query := " WHERE "
	i := 0
	for _, p := range q.Params {
		query = query + p.Field
		switch p.Op {
		case Op_EQUAL:
			query += "=?"
		case Op_IN:
			query += " IN (?)"
		}
		i++
		if i != len(q.Params) {
			query = query + " AND "
		}
	}
	return querySelect + query
}

func (q Query) Run(tx Txn) (Alerts, error) {
	var alerts Alerts
	sql := q.toSQL()
	var values []interface{}
	for _, p := range q.Params {
		for _, v := range p.Values {
			values = append(values, v)
		}
	}
	err := tx.InSelect(sql, &alerts, values...)
	if err != nil {
		return alerts, err
	}
	if len(q.Params) == 0 {
		// set a default limit of 10
		if len(alerts) > 10 {
			q.Limit = 10
		} else {
			q.Limit = len(alerts)
		}
	}
	if q.Offset > 0 {
		alerts = alerts[q.Offset:]
	}
	if q.Limit > 0 {
		return alerts[:q.Limit], nil
	}
	return alerts, nil
}

type Field struct {
	Name, Value string
}

type UpdateQuery struct {
	Set   []Field
	Where []Param
}

func (u UpdateQuery) toSQL() string {
	query := " SET "
	i := 0
	for _, f := range u.Set {
		query += fmt.Sprintf("%s=?", f.Name)
		i++
		if i != len(u.Set) {
			query = query + ", "
		}
	}
	if len(u.Where) == 0 {
		return queryUpdate + query
	}
	query += " WHERE "
	i = 0
	for _, p := range u.Where {
		query = query + p.Field
		switch p.Op {
		case Op_EQUAL:
			query += "=?"
		case Op_IN:
			query += " IN (?)"
		}
		i++
		if i != len(u.Where) {
			query = query + " AND "
		}
	}
	return queryUpdate + query
}

func (u UpdateQuery) Run(tx Txn) (Alerts, error) {
	var alerts Alerts // dummy so that Run can conform to Querier interface
	sql := u.toSQL()
	var values []interface{}
	for _, f := range u.Set {
		values = append(values, f.Value)
	}
	for _, p := range u.Where {
		for _, v := range p.Values {
			values = append(values, v)
		}
	}
	err := tx.InQuery(sql, values...)
	if err != nil {
		return alerts, err
	}
	return alerts, nil
}
