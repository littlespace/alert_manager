package models

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
)

const labelQueryTpl = `{{$key := .Key}}{{$length := len .Values }}({{$key}} IN (
  {{- range $i, $e := .Values }}
  {{- if $i }},{{- end }}'{{.}}'
  {{- end }})
  {{- range .Values }} OR (labels::jsonb)->'{{$key}}' ? '{{.}}'{{- end }})`

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
		if p.Field == "device" || p.Field == "entity" || p.Field == "site" {
			if quer, err := appendLabelQuery(query, p); err == nil {
				query = quer
				if i != len(q.Params)-1 {
					query = query + " AND "
				}
				continue
			}
		}
		query = query + p.Field
		query = buildQuery(query, p)
		if i != len(q.Params)-1 {
			query = query + " AND "
		}
	}
	return baseQuery + query
}

func (q Query) Run(tx Txn) ([]interface{}, error) {
	var items []interface{}
	sql := q.toSQL()
	var err error
	switch q.Table {
	case "alerts":
		var alerts Alerts
		alerts, err = tx.SelectAlerts(sql)
		for _, a := range alerts {
			items = append(items, a)
		}
	case "suppression_rules":
		var rules SuppRules
		rules, err = tx.SelectRules(sql)
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
		query += fmt.Sprintf("%s='%s'", f.Name, f.Value)
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
		query = buildQuery(query, p)
		if i != len(u.Where)-1 {
			query = query + " AND "
		}
	}
	return baseQuery + query
}

func (u UpdateQuery) Run(tx Txn) ([]interface{}, error) {
	var items []interface{} // dummy so that Run can conform to Querier interface
	sql := u.toSQL()
	err := tx.Exec(sql)
	if err != nil {
		return items, err
	}
	return items, nil
}

func buildQuery(query string, p Param) string {
	p = sanitizeParam(p)
	switch p.Op {
	case Op_EQUAL:
		return query + fmt.Sprintf("='%s'", p.Values[0])
	case Op_IN:
		query += " IN ("
		for i, v := range p.Values {
			if i != len(p.Values)-1 {
				query += fmt.Sprintf("'%s', ", v)
			} else {
				query += fmt.Sprintf("'%s')", v)
			}
		}
		return query
	}
	return ""
}

func appendLabelQuery(query string, p Param) (string, error) {
	data := struct {
		Key    string
		Values []string
	}{Key: p.Field, Values: p.Values}

	tpl, err := template.New("query").Parse(labelQueryTpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, data); err != nil {
		return "", err
	}
	query = query + b.String()
	return query, nil
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
	for j, v := range p.Values {
		query = query + fmt.Sprintf("'%s' = ANY(tags)", v)
		if j != len(p.Values)-1 {
			query = query + " AND "
			continue
		}
	}
	return query
}
