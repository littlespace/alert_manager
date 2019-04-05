package models

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type MockDb struct{}

func (d MockDb) NewTx() Txn {
	return &MockTx{}
}

func (d MockDb) Close() error {
	return nil
}

type MockTx struct {
	*Tx
}

func (tx *MockTx) SelectAlerts(query string, args ...interface{}) (Alerts, error) {
	var alerts Alerts
	for i := 1; i <= 10; i++ {
		alerts = append(alerts, &Alert{
			Id:          int64(i),
			Name:        "mock",
			Description: "test",
			Entity:      "e1",
			Source:      "src",
			Scope:       "scp",
		})
	}
	if strings.Contains(query, "OFFSET") {
		return alerts[5:], nil
	}
	if strings.Contains(query, "LIMIT") && !strings.Contains(query, "50") {
		return alerts[:5], nil
	}
	return alerts, nil
}

func (tx *MockTx) SelectAlertsWithHistory(query string, args ...interface{}) (Alerts, error) {
	alerts, _ := tx.SelectAlerts(query, args...)
	for _, a := range alerts {
		a.History = append(a.History, &Record{AlertId: a.Id, Event: "foobar"})
	}
	return alerts, nil
}

const baseQ = "SELECT * FROM alerts WHERE (cast(extract(epoch from now()) as integer) - last_active) < 5"

var testDatas = map[string]Querier{
	baseQ + " AND alerts.id IN ('1') AND name IN ('foo')": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1"}},
			Param{Field: "name", Values: []string{"foo"}},
		},
	},
	baseQ + " AND alerts.id IN ('1','2') AND status IN ('1')": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2"}},
			Param{Field: "status", Values: []string{"ACTIVE"}},
		},
	},
	baseQ + " AND alerts.id IN ('1','2') AND name IN ('foo','bar')": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2"}},
			Param{Field: "name", Values: []string{"foo", "bar"}},
		},
	},
	baseQ + " AND alerts.id IN ('1') AND 'foo' = ANY(tags)": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1"}},
			Param{Field: "tags", Values: []string{"foo"}},
		},
	},
	baseQ + " AND alerts.id IN ('1','2') AND 'foo' = ANY(tags) AND 'bar' = ANY(tags)": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2"}},
			Param{Field: "tags", Values: []string{"foo", "bar"}},
		},
	},
	baseQ + " AND (device IN ('d1','d2') OR (labels::jsonb)->'device' ? 'd1' OR (labels::jsonb)->'device' ? 'd2')": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "device", Values: []string{"d1", "d2"}},
		},
	},
	baseQ + " AND (device IN ('d1') OR (labels::jsonb)->'device' ? 'd1') AND status IN ('1','2')": Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "device", Values: []string{"d1"}},
			Param{Field: "status", Values: []string{"ACTIVE", "SUPPRESSED"}},
		},
	},
	"UPDATE alerts SET owner='foo' WHERE id IN ('1') AND name IN ('foo')": UpdateQuery{
		Set: []Field{
			Field{Name: "owner", Value: "foo"},
		},
		Where: []Param{
			Param{Field: "id", Values: []string{"1"}},
			Param{Field: "name", Values: []string{"foo"}},
		},
	},
	"UPDATE alerts SET owner='foo', team='bar' WHERE id IN ('1','2') AND name IN ('foo')": UpdateQuery{
		Set: []Field{
			Field{Name: "owner", Value: "foo"},
			Field{Name: "team", Value: "bar"},
		},
		Where: []Param{
			Param{Field: "id", Values: []string{"1", "2"}},
			Param{Field: "name", Values: []string{"foo"}},
		},
	},
	"UPDATE alerts SET owner='foo', team='bar' WHERE 'foo' = ANY(tags) AND 'bar' = ANY(tags) AND name IN ('foo')": UpdateQuery{
		Set: []Field{
			Field{Name: "owner", Value: "foo"},
			Field{Name: "team", Value: "bar"},
		},
		Where: []Param{
			Param{Field: "tags", Values: []string{"foo", "bar"}},
			Param{Field: "name", Values: []string{"foo"}},
		},
	},
}

func TestQuerySQL(t *testing.T) {
	for sql, q := range testDatas {
		renderedSql, err := q.toSQL()
		assert.Nil(t, err)
		assert.Equal(t, sql, renderedSql)
	}
}

func TestSelectQueryRun(t *testing.T) {
	q := Query{
		Table:     "alerts",
		TimeRange: "5s",
		Params: []Param{
			Param{Field: "id", Values: []string{"1"}},
			Param{Field: "name", Values: []string{"foo"}},
		},
	}
	sql, err := q.toSQL()
	assert.Nil(t, err)
	assert.Equal(t, sql, baseQ+" AND alerts.id IN ('1') AND name IN ('foo')")
	tx := &MockTx{}
	items, err := q.Run(tx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(items), 10)

	q.Limit = 5
	items, err = q.Run(tx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(items), 5)
	var ids []int64
	for _, a := range items {
		ids = append(ids, a.(*Alert).Id)
	}
	assert.ElementsMatch(t, ids, []int64{1, 2, 3, 4, 5})

	q.Limit = 0
	q.Offset = 5
	items, err = q.Run(tx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(items), 5)
	ids = ids[:0]
	for _, a := range items {
		ids = append(ids, a.(*Alert).Id)
	}
	assert.ElementsMatch(t, ids, []int64{6, 7, 8, 9, 10})

	q.IncludeHistory = true
	items, err = q.Run(tx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(items[0].(*Alert).History), 1)
}
