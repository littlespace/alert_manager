package models

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const queryTpl = `
{{- $paramLen := len .Params }}
{{- .BaseQ}} WHERE {{ if .TimeRange}}(cast(extract(epoch from now()) as integer) - {{.TimeRange.TrStart}}) < {{.TimeRange.Seconds}}
{{- if gt $paramLen 0}} AND {{ end }}
{{- end }}
{{- range $index, $param := .Params }}
{{- $f := .Field}}
{{- if eq $f "tags"}}
{{- range $i, $e := .Values }}
{{- if $i }} AND {{ end }}'{{$e}}' = ANY({{$f}}){{- end }}
{{- else if or (eq $f "device") (eq $f "site") (eq $f "entity") }}
{{- $key := .Field}}{{$length := len .Values }}({{$key}} IN (
{{- range $i, $e := .Values }}
{{- if $i }},{{- end }}'{{.}}'
{{- end }})
{{- range .Values }} OR (labels::jsonb)->'{{$key}}' ? '{{.}}'{{- end }})
{{- else }}
{{- .Field }} IN (
{{- range $i, $e := .Values }}
{{- if $i }},{{- end }}'{{.}}'
{{- end }})
{{- end}}
{{- if ne ($index) (decr $paramLen) }} AND {{ end }}
{{- end}}`

const updateTpl = `
{{- $setLen := len .Set }}{{$whereLen := len .Where }}
{{- .BaseQ}} SET {{ range $index, $set := .Set }}{{.Name}}='{{.Value}}'
{{- if ne ($index) (decr $setLen)}}, {{ end }}
{{- end }} WHERE {{ range $index, $param := .Where }}
{{- $f := .Field}}
{{- if eq $f "tags"}}
{{- range $i, $e := .Values }}
{{- if $i }} AND {{ end }}'{{$e}}' = ANY({{$f}}){{- end }}
{{- else }}
{{- .Field }} IN (
{{- range $i, $e := .Values }}
{{- if $i }},{{- end }}'{{.}}'
{{- end }})
{{- end}}
{{- if ne ($index) (decr $whereLen) }} AND {{ end }}
{{- end}}`

func executeQueryTpl(raw string, data map[string]interface{}) (string, error) {
	funcMap := template.FuncMap{
		"decr": func(i int) int {
			return i - 1
		},
	}
	tpl, err := template.New("sql").Funcs(funcMap).Parse(raw)
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
	toSQL() (string, error)
}

type Query struct {
	Table          string
	BaseQuery      string
	Limit          int
	Offset         int
	TimeRange      string
	IncludeHistory bool
	Params         []Param
}

func NewQuery(table string) Query {
	q := Query{Table: table}
	switch table {
	case "alerts", "suppression_rules", "teams":
		q.BaseQuery = fmt.Sprintf("SELECT * FROM %s", q.Table)
	case "users":
		q.BaseQuery = QuerySelectUsers
	}
	return q
}

func (q Query) toSQL() (string, error) {
	baseQ := q.BaseQuery
	start := "last_active"
	if q.Table == "suppression_rules" {
		start = "created_at"
	}
	var timeRange interface{}
	tr, err := time.ParseDuration(q.TimeRange)
	if err == nil && tr > 0 && q.Table != "teams" && q.Table != "users" {
		timeRange = struct {
			TrStart string
			Seconds int64
		}{TrStart: start, Seconds: int64(tr.Seconds())}
	}
	if len(q.Params) == 0 && (err != nil || tr == 0) {
		return baseQ, nil
	}
	var sanitizedParams []Param
	for _, p := range q.Params {
		if p.Field == "id" && q.Table == "alerts" {
			p.Field = "alerts.id"
		}
		// users and teams tables are joined so we need to qualify the field values
		// with table names e.g name -> users.name
		if q.Table == "users" || q.Table == "teams" && !strings.HasPrefix(p.Field, q.Table) {
			p.Field = fmt.Sprintf("%s.%s", q.Table, p.Field)
		}
		sanitizedParams = append(sanitizedParams, sanitizeParam(p))
	}
	data := map[string]interface{}{
		"BaseQ":     baseQ,
		"TimeRange": timeRange,
		"Params":    sanitizedParams,
	}
	sql, err := executeQueryTpl(queryTpl, data)
	if err != nil {
		return "", fmt.Errorf("Failed to execute query template: %v", err)
	}
	return sql, nil
}

func (q Query) Run(tx Txn) ([]interface{}, error) {
	items := []interface{}{}
	sql, err := q.toSQL()
	if err != nil {
		return items, err
	}
	sql += fmt.Sprintf(" ORDER BY %s.id DESC", q.Table)
	if q.Limit == 0 {
		q.Limit = 50
	}
	sql += fmt.Sprintf(" LIMIT %d", q.Limit)
	if q.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.Offset)
	}
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
	case "teams":
		var teams Teams
		teams, err = tx.SelectTeams(sql)
		for _, r := range teams {
			items = append(items, r)
		}
	case "users":
		var users Users
		users, err = tx.SelectUsers(sql)
		for _, r := range users {
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
	Table     string
	BaseQuery string
	Set       []Field
	Where     []Param
}

func NewUpdateQuery(table string) UpdateQuery {
	q := UpdateQuery{Table: table}
	baseQuery := queryUpdateAlerts
	if table == "suppression_rules" {
		baseQuery = queryUpdateRules
	}
	q.BaseQuery = baseQuery
	return q
}

func (u UpdateQuery) toSQL() (string, error) {
	var sanitizedParams []Param
	for _, p := range u.Where {
		sanitizedParams = append(sanitizedParams, sanitizeParam(p))
	}
	data := map[string]interface{}{
		"BaseQ": u.BaseQuery,
		"Set":   u.Set,
		"Where": sanitizedParams,
	}
	sql, err := executeQueryTpl(updateTpl, data)
	if err != nil {
		return "", err
	}
	return sql, nil
}

func (u UpdateQuery) Run(tx Txn) ([]interface{}, error) {
	items := []interface{}{} // dummy so that Run can conform to Querier interface
	sql, err := u.toSQL()
	if err != nil {
		return items, err
	}
	err = tx.Exec(sql)
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
