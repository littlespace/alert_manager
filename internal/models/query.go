package models

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
	"time"
)

const labelQueryTpl = `{{$key := .Field}}{{$length := len .Values }}({{$key}} IN (
{{- range $i, $e := .Values }}
{{- if $i }},{{- end }}'{{.}}'
{{- end }})
{{- range .Values }} OR (labels::jsonb)->'{{$key}}' ? '{{.}}'{{- end }})`

const sqlTpl = `{{.Field}} IN (
{{- range $i, $e := .Values }}
{{- if $i }},{{- end }}'{{.}}'
{{- end }})`

const tagTpl = `{{$f := .Field}}
{{- range $i, $e := .Values }}
{{- if $i }} AND {{ end }}'{{$e}}' = ANY({{$f}})
{{- end }}`

func executeQueryTpl(raw string, data Param) (string, error) {
	tpl, err := template.New("sql").Parse(raw)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

type Param struct {
	Field  string
	Values []string
}

type Querier interface {
	Run(tx Txn) ([]interface{}, error)
	toSQL() string
}

type Query struct {
	Table          string
	Limit          int
	Offset         int
	TimeRange      string
	IncludeHistory bool
	Params         []Param
}

func NewQuery(table string) Query {
	return Query{Table: table, TimeRange: "72h"}
}

func (q Query) toSQL() string {
	baseQ := fmt.Sprintf("SELECT * FROM %s", q.Table)
	start := "start_time"
	if q.Table == "suppression_rules" {
		start = "created_at"
	}
	tr, err := time.ParseDuration(q.TimeRange)
	if err == nil && tr > 0 {
		baseQ += fmt.Sprintf(" WHERE (cast(extract(epoch from now()) as integer) - %s) < %d", start, int64(tr.Seconds()))
	}
	if len(q.Params) == 0 {
		return baseQ
	}
	query := baseQ + " AND "
	for i, p := range q.Params {
		p = sanitizeParam(p)
		if p.Field == "tags" {
			if quer, err := executeQueryTpl(tagTpl, p); err == nil {
				query += quer
			}
			if i != len(q.Params)-1 {
				query = query + " AND "
			}
			continue
		}
		if p.Field == "device" || p.Field == "entity" || p.Field == "site" {
			if quer, err := executeQueryTpl(labelQueryTpl, p); err == nil {
				query += quer
				if i != len(q.Params)-1 {
					query = query + " AND "
				}
				continue
			}
		}
		if p.Field == "id" && q.Table == "alerts" {
			p.Field = "alerts.id"
		}
		if quer, err := executeQueryTpl(sqlTpl, p); err == nil {
			query += quer
		}
		if i != len(q.Params)-1 {
			query = query + " AND "
		}
	}
	return query
}

func (q Query) Run(tx Txn) ([]interface{}, error) {
	var items []interface{}
	sql := q.toSQL()
	sql += fmt.Sprintf(" ORDER BY %s.id", q.Table)
	if q.Limit == 0 {
		q.Limit = 25
	}
	sql += fmt.Sprintf(" LIMIT %d", q.Limit)
	if q.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.Offset)
	}
	var err error
	switch q.Table {
	case "alerts":
		var alerts Alerts
		if q.IncludeHistory {
			alerts, err = tx.SelectAlertsWithHistory(sql)
		} else {
			alerts, err = tx.SelectAlerts(sql)
		}
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
		p = sanitizeParam(p)
		if p.Field == "tags" {
			if quer, err := executeQueryTpl(tagTpl, p); err == nil {
				query += quer
			}
			if i != len(u.Where)-1 {
				query = query + " AND "
			}
			continue
		}
		if quer, err := executeQueryTpl(sqlTpl, p); err == nil {
			query += quer
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
	err := tx.Exec(sql)
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
	return Param{Field: p.Field, Values: newVal}
}
