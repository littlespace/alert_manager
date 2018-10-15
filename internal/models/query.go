package models

import (
	"fmt"
	"strconv"
	"strings"
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
	Run(tx Txn) ([]interface{}, error)
	toSQL() string
}

type Query struct {
	Table  string
	Limit  int
	Offset int
	Params []Param
}

func NewQuery(table string) Query {
	return Query{Table: table}
}

func (q Query) toSQL() string {
	baseQuery := querySelectAlerts
	if q.Table == "suppression_rules" {
		baseQuery = querySelectRules
	}
	if len(q.Params) == 0 {
		return baseQuery
	}
	query := " WHERE "
	for i, p := range q.Params {
		if p.Field == "tags" {
			query = handleTags(query, p)
			if i != len(q.Params)-1 {
				query = query + " AND "
			}
			continue
		}
		query = query + p.Field
		switch p.Op {
		case Op_EQUAL:
			query += "=?"
		case Op_IN:
			query += " IN (?)"
		}
		if i != len(q.Params)-1 {
			query = query + " AND "
		}
	}
	return baseQuery + query
}

func (q Query) Run(tx Txn) ([]interface{}, error) {
	var items []interface{}
	sql := q.toSQL()
	var values []interface{}
	for _, p := range q.Params {
		p = sanitizeParam(p)
		if p.Op == Op_IN && p.Field != "tags" {
			values = append(values, p.Values)
			continue
		}
		for _, v := range p.Values {
			values = append(values, v)
		}
	}
	var err error
	switch q.Table {
	case "alerts":
		var alerts Alerts
		if strings.Contains(sql, "IN") {
			err = tx.InSelect(sql, &alerts, values...)
		} else {
			alerts, err = tx.SelectAlerts(sql, values...)
		}
		for _, a := range alerts {
			items = append(items, a)
		}
	case "suppression_rules":
		var rules SuppRules
		if strings.Contains(sql, "IN") {
			err = tx.InSelect(sql, &rules, values...)
		} else {
			rules, err = tx.SelectRules(sql, values...)
		}
		for _, r := range rules {
			items = append(items, r)
		}
	}
	if err != nil {
		return items, err
	}
	if len(q.Params) == 0 {
		// set a default limit of 25
		if len(items) > 25 {
			q.Limit = 25
		} else {
			q.Limit = len(items)
		}
	}
	if q.Offset > 0 {
		if q.Offset > len(items) {
			return items, nil
		}
		items = items[q.Offset:]
	}
	if q.Limit > 0 {
		if q.Limit > len(items) {
			return items, nil
		}
		return items[:q.Limit], nil
	}
	return items, nil
}

type Field struct {
	Name, Value string
}

type UpdateQuery struct {
	Table string
	Set   []Field
	Where []Param
}

func NewUpdateQuery(table string) UpdateQuery {
	return UpdateQuery{Table: table}
}

func (u UpdateQuery) toSQL() string {
	baseQuery := queryUpdateAlerts
	if u.Table == "suppression_rules" {
		baseQuery = queryUpdateRules
	}
	query := " SET "
	for i, f := range u.Set {
		query += fmt.Sprintf("%s=?", f.Name)
		if i != len(u.Set)-1 {
			query = query + ", "
		}
	}
	if len(u.Where) == 0 {
		return baseQuery + query
	}
	query += " WHERE "
	for i, p := range u.Where {
		if p.Field == "tags" {
			query = handleTags(query, p)
			if i != len(u.Where)-1 {
				query = query + " AND "
			}
			continue
		}
		query = query + p.Field
		switch p.Op {
		case Op_EQUAL:
			query += "=?"
		case Op_IN:
			query += " IN (?)"
		}
		if i != len(u.Where)-1 {
			query = query + " AND "
		}
	}
	return baseQuery + query
}

func (u UpdateQuery) Run(tx Txn) ([]interface{}, error) {
	var items []interface{} // dummy so that Run can conform to Querier interface
	sql := u.toSQL()
	var values []interface{}
	for _, f := range u.Set {
		values = append(values, f.Value)
	}
	for _, p := range u.Where {
		p = sanitizeParam(p)
		if p.Op == Op_IN && p.Field != "tags" {
			values = append(values, p.Values)
			continue
		}
		for _, v := range p.Values {
			values = append(values, v)
		}
	}
	err := tx.InQuery(sql, values...)
	if err != nil {
		return items, err
	}
	return items, nil
}

func sanitizeParam(p Param) Param {
	var newVal []string
	for _, v := range p.Values {
		if p.Field == "status" {
			if _, err := strconv.Atoi(v); err != nil {
				newVal = append(newVal, strconv.Itoa(int(StatusMap[v])))
				continue
			}
		} else if p.Field == "severity" {
			if _, err := strconv.Atoi(v); err != nil {
				newVal = append(newVal, strconv.Itoa(int(SevMap[v])))
				continue
			}
		}
		newVal = append(newVal, v)
	}
	return Param{Field: p.Field, Op: p.Op, Values: newVal}
}

func handleTags(query string, p Param) string {
	// special handling for array data type
	for j, _ := range p.Values {
		query = query + "? = ANY(tags)"
		if j != len(p.Values)-1 {
			query = query + " AND "
			continue
		}
	}
	return query
}
