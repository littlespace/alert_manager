package handler

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var mockRules = map[string]models.SuppressionRule{
	"rule1": models.NewSuppRule(models.Labels{"alert_name": "Test Alert 1"}, "alert", "test", "test", 1*time.Minute),
	"rule2": models.NewSuppRule(models.Labels{"device": "dev2"}, "device", "test", "test", 1*time.Minute),
	"rule3": models.NewSuppRule(models.Labels{"device": "dev3", "entity": "ent1"}, "entity", "test", "test", 1*time.Minute),
}

type MockDb2 struct{}

func (m *MockDb2) NewTx() models.Txn {
	return &MockTx2{}
}

func (m *MockDb2) Close() error {
	return nil
}

type MockTx2 struct {
	*models.Tx
}

func (tx *MockTx2) SelectRules(query string, args ...interface{}) (models.SuppRules, error) {
	m := models.SuppRules{}
	for n, r := range mockRules {
		if n == "rule2" {
			r.CreatedAt.Time = r.CreatedAt.Add(-5 * time.Minute)
		}
		m = append(m, r)
	}
	return m, nil
}

func (tx *MockTx2) NewSuppRule(rule *models.SuppressionRule) (int64, error) {
	return 1, nil
}

func (tx *MockTx2) UpdateAlert(alert *models.Alert) error {
	return nil
}

func (tx *MockTx2) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	a := tu.MockAlert(args[0].(int64), "Test Alert 1", "", "dev1", "ent1", "src1", "scp1", "1", "WARN", []string{}, nil)
	a.Status = models.Status_SUPPRESSED
	if a.Id == 2 {
		a.Status = models.Status_CLEARED
	}
	return a, nil
}

func (tx *MockTx2) Rollback() error {
	return nil
}

func (tx *MockTx2) Commit() error {
	return nil
}

func TestRuleMatch(t *testing.T) {
	s := &suppressor{db: &MockDb2{}}
	s.loadSuppRules(context.Background())

	// test active match - any
	labels := models.Labels{"alert_name": "Test Alert 1", "device": "dev1"}
	rule, ok := s.Match(labels, models.MatchCond_ANY)
	assert.Equal(t, rule, mockRules["rule1"])

	// test active match - all
	labels = models.Labels{"device": "dev3", "entity": "ent1"}
	rule, ok = s.Match(labels, models.MatchCond_ALL)
	assert.Equal(t, rule, mockRules["rule3"])

	// test no match
	labels = models.Labels{"foo": "bar"}
	rule, ok = s.Match(labels, models.MatchCond_ANY)
	assert.Equal(t, ok, false)

	// test expired match - rule removal
	labels = models.Labels{"device": "dev2"}
	rule, ok = s.Match(labels, models.MatchCond_ALL)
	assert.Equal(t, ok, true)
	rule, ok = s.Match(labels, models.MatchCond_ALL)
	assert.Equal(t, ok, false)
}

func TestSaveRule(t *testing.T) {
	e := models.Labels{"alert_id": 1}
	r := models.NewSuppRule(e, "alert", "test", "test", 5*time.Minute)
	s := &suppressor{db: &MockDb2{}}
	if _, err := s.SaveRule(context.Background(), &MockTx2{}, r); err != nil {
		t.Fatal(err)
	}
	rule, _ := s.Match(e, models.MatchCond_ANY)
	assert.Equal(t, int(rule.Id), 1)
}

func TestSuppAlert(t *testing.T) {
	a1 := tu.MockAlert(1, "Test Alert 1", "", "dev1", "ent1", "src1", "scp1", "1", "WARN", []string{}, nil)
	s := &suppressor{db: &MockDb{}}
	ctx := context.Background()
	tx := &MockTx2{}

	// suppress and find a match
	labels := models.Labels{"alert_id": a1.Id}
	r := models.NewSuppRule(labels, "alert", "test", "test", 1*time.Minute)
	if err := s.SuppressAlert(ctx, tx, a1, r); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a1.Status, models.Status_SUPPRESSED)
	rule, ok := s.Match(labels, models.MatchCond_ANY)
	assert.Equal(t, ok, true)
	assert.Equal(t, int(rule.Id), 1)

	// unsuppress
	if err := s.UnsuppressAlert(ctx, tx, a1); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a1.Status, models.Status_ACTIVE)

	a2 := tu.MockAlert(2, "Test Alert 1", "", "dev1", "ent1", "src1", "scp1", "1", "WARN", []string{}, nil)
	r = models.NewSuppRule(models.Labels{"alert_id": a2.Id}, "alert", "test", "test", 1*time.Minute)
	if err := s.SuppressAlert(ctx, tx, a2, r); err != nil {
		t.Fatal(err)
	}
	// try to unsuppress a cleared alert, no go
	err := s.UnsuppressAlert(ctx, tx, a2)
	assert.Error(t, err)
	assert.Equal(t, a2.Status, models.Status_SUPPRESSED)
}
