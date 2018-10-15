package models

import (
	"github.com/stretchr/testify/assert"
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

func (tx *MockTx) InQuery(query string, arg ...interface{}) error {
	return nil
}

func (tx *MockTx) InSelect(query string, to interface{}, arg ...interface{}) error {
	if len(arg) == 0 {
		return nil
	}
	for i, _ := range arg[0].([]string) {
		a := Alert{
			Id:          int64(i),
			Name:        "mock",
			Description: "test",
			Entity:      "e1",
			Source:      "src",
			Scope:       "scp",
		}
		if to, ok := to.(*Alerts); ok {
			*to = append(*to, a)
		}
	}
	return nil
}

func (tx *MockTx) SelectAlerts(query string, args ...interface{}) (Alerts, error) {
	var alerts Alerts
	for i := 1; i <= 10; i++ {
		alerts = append(alerts, Alert{
			Id:          int64(i),
			Name:        "mock",
			Description: "test",
			Entity:      "e1",
			Source:      "src",
			Scope:       "scp",
		})
	}
	return alerts, nil
}

var testDatas = map[string]Querier{
	"SELECT * from alerts WHERE id=? AND name=?": Query{
		Params: []Param{
			Param{Field: "id", Values: []string{"1"}, Op: Op_EQUAL},
			Param{Field: "name", Values: []string{"foo"}, Op: Op_EQUAL},
		},
	},
	"SELECT * from alerts WHERE id IN (?) AND name=?": Query{
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2"}, Op: Op_IN},
			Param{Field: "name", Values: []string{"foo"}, Op: Op_EQUAL},
		},
	},
	"SELECT * from alerts WHERE id IN (?) AND name IN (?)": Query{
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2"}, Op: Op_IN},
			Param{Field: "name", Values: []string{"foo", "bar"}, Op: Op_IN},
		},
	},
	"UPDATE alerts SET owner=? WHERE id=? AND name=?": UpdateQuery{
		Set: []Field{
			Field{Name: "owner", Value: "foo"},
		},
		Where: []Param{
			Param{Field: "id", Values: []string{"1"}, Op: Op_EQUAL},
			Param{Field: "name", Values: []string{"foo"}, Op: Op_EQUAL},
		},
	},
	"UPDATE alerts SET owner=?, team=? WHERE id IN (?) AND name=?": UpdateQuery{
		Set: []Field{
			Field{Name: "owner", Value: "foo"},
			Field{Name: "team", Value: "bar"},
		},
		Where: []Param{
			Param{Field: "id", Values: []string{"1", "2"}, Op: Op_IN},
			Param{Field: "name", Values: []string{"foo"}, Op: Op_EQUAL},
		},
	},
}

func TestQuerySQL(t *testing.T) {
	for sql, q := range testDatas {
		assert.Equal(t, sql, q.toSQL())
	}
}

func TestSelectQueryRun(t *testing.T) {
	q := Query{
		Table: "alerts",
		Params: []Param{
			Param{Field: "id", Values: []string{"1"}, Op: Op_EQUAL},
			Param{Field: "name", Values: []string{"foo"}, Op: Op_EQUAL},
		},
	}
	assert.Equal(t, q.toSQL(), "SELECT * from alerts WHERE id=? AND name=?")
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
		ids = append(ids, a.(Alert).Id)
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
		ids = append(ids, a.(Alert).Id)
	}
	assert.ElementsMatch(t, ids, []int64{6, 7, 8, 9, 10})

	q = Query{
		Table: "alerts",
		Params: []Param{
			Param{Field: "id", Values: []string{"1", "2", "3"}, Op: Op_IN},
		},
	}
	assert.Equal(t, q.toSQL(), "SELECT * from alerts WHERE id IN (?)")
	items, err = q.Run(tx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(items), 3)
}
