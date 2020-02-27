package handler

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

var mockAlerts = map[string]*models.Alert{
	"existing_a1":     tu.MockAlert(100, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil),
	"existing_a2":     tu.MockAlert(200, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil),
	"existing_a3":     tu.MockAlert(300, "Test Alert 3", "", "d3", "e3", "src3", "scp3", "t1", "3", "WARN", []string{"e", "f"}, nil),
	"existing_a4":     tu.MockAlert(400, "Test Alert 4", "", "d4", "e4", "src4", "scp4", "t1", "4", "INFO", []string{"e", "f"}, nil),
	"existing_a5":     tu.MockAlert(500, "Test Alert 5", "", "d5", "e5", "src5", "scp5", "t1", "5", "WARN", []string{"g", "h"}, nil),
	"existing_a6_agg": tu.MockAlert(600, "Test Alert 6", "", "d6", "e6", "src6", "scp6", "t1", "6", "WARN", []string{"g", "h"}, nil),
	"existing_a7":     tu.MockAlert(700, "Test Alert 7", "", "d7", "e7", "src7", "scp7", "t1", "7", "WARN", []string{"g", "h"}, nil),
}

var nowTime = models.MyTime{time.Now()}

type MockDb struct{}

func (m *MockDb) NewTx() models.Txn {
	return &MockTx{}
}

func (m *MockDb) Close() error {
	return nil
}

type MockTx struct {
	*models.Tx
	inQuery     func(query string) error
	updateAlert func(alert *models.Alert) error
	newInsert   func(query string, item interface{}) (int64, error)
}

func (t *MockTx) NewInsert(query string, item interface{}) (int64, error) {
	if t.newInsert != nil {
		return t.newInsert(query, item)
	}
	switch item.(type) {
	case *models.Alert:
		alert := item.(*models.Alert)
		for _, a := range mockAlerts {
			if a.Name == alert.Name && a.AggregatorId == 0 {
				return a.Id, nil
			}
		}
		return 999, nil
	case *models.Team:
		return 1, nil
	}
	return 0, nil
}

func (t *MockTx) UpdateAlert(alert *models.Alert) error {
	if t.updateAlert != nil {
		return t.updateAlert(alert)
	}
	return nil
}

func (t *MockTx) Rollback() error {
	return nil
}

func (t *MockTx) Commit() error {
	return nil
}

func (t *MockTx) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	switch args[0].(type) {
	case string:
		for _, a := range mockAlerts {
			if args[0].(string) == a.Name {
				if query == models.QuerySelectByDevice {
					return a, nil
				}
			}
		}
	case int64:
		for _, a := range mockAlerts {
			if args[0].(int64) == a.Id && query == models.QuerySelectById {
				return a, nil
			}
		}
	}
	return nil, fmt.Errorf("No alert found")
}

func (t *MockTx) InQuery(query string, args ...interface{}) error {
	if t.inQuery != nil {
		return t.inQuery(query)
	}
	return fmt.Errorf("InQuery undefined")
}

func (t *MockTx) InSelect(query string, to interface{}, arg ...interface{}) error {
	if query == models.QuerySelectByAggId {
		if to, ok := to.(*models.Alerts); ok {
			*to = append(*to, mockAlerts["existing_a1"])
			*to = append(*to, mockAlerts["existing_a2"])
		}
	}
	return nil
}

func (tx *MockTx) Exec(query string, args ...interface{}) error {
	return nil
}

func (t *MockTx) SelectAlerts(query string, args ...interface{}) (models.Alerts, error) {
	switch query {
	case models.QuerySelectExpired:
		return models.Alerts{mockAlerts["existing_a3"]}, nil
	case models.QuerySelectNoOwner:
		return models.Alerts{mockAlerts["existing_a4"]}, nil
	}
	return models.Alerts{}, nil
}

func (t *MockTx) SelectRules(query string, args ...interface{}) (models.SuppRules, error) {
	return models.SuppRules{}, nil
}

func (t *MockTx) NewRecord(alertId int64, event string) (int64, error) {
	return 1, nil
}

type mockTransform struct {
	name     string
	priority int
	register string
}

func (t *mockTransform) Name() string { return t.name }

func (t *mockTransform) GetPriority() int { return t.priority }

func (t *mockTransform) GetRegister() string { return t.register }

func (t *mockTransform) Apply(a *models.Alert) error {
	a.Labels = models.Labels{"suppress": "me"}
	return nil
}

type mockProcessor struct{}

func (m *mockProcessor) Name() string { return "mockProcessor" }

func (m *mockProcessor) Stage() int { return 1 }

func (m *mockProcessor) Process(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) chan *models.AlertEvent {
	return make(chan *models.AlertEvent)
}

func NewTestHandler(procChanSize int) *AlertHandler {
	m := &MockDb{}
	h := &AlertHandler{Db: m, statTransformError: &tu.MockStat{}, statDbError: &tu.MockStat{}}
	h.procChan = make(chan *models.AlertEvent, procChanSize)
	h.Suppressor = &suppressor{db: m}
	return h
}

func TestHandlerAlertActiveNew(t *testing.T) {
	h := NewTestHandler(2)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test new active alert
	new := tu.MockAlert(0, "New Alert 1", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	err := h.handleActive(ctx, tx, new)
	assert.Nil(t, err)
	assert.Equal(t, int(new.Id), 999)
	assert.Equal(t, h.Teams[0].Name, "t1")
	assert.Equal(t, h.Teams[0].Id, int64(1))
	l := models.Labels{"suppress": "me", "description": "", "device": "d2", "entity": "e2", "source": "src2", "scope": "scp2", "severity": "WARN", "alert_name": "New Alert 1", "team": "t1"}
	assert.Equal(t, new.Labels, l)
	event := <-h.procChan
	assert.Equal(t, event.Type, models.EventType_ACTIVE)
	assert.Equal(t, int(event.Alert.Id), 999)

	// test new team
	a3 := tu.MockAlert(0, "New Alert 6", "", "d6", "e6", "src6", "scp6", "t2", "2", "WARN", []string{"c", "d"}, nil)
	err = h.handleActive(ctx, tx, a3)
	assert.Nil(t, err)
	assert.Equal(t, h.Teams.Contains("t2"), true)
}

func TestHandlerAlertActiveExisting(t *testing.T) {
	h := NewTestHandler(1)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test existing active alert
	tx.(*MockTx).updateAlert = func(alert *models.Alert) error {
		mockAlerts["existing_a1"].LastActive = nowTime
		return nil
	}
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	err := h.handleActive(ctx, tx, a1)
	assert.Nil(t, err)
	assert.Equal(t, int(a1.Id), 0)
	assert.Equal(t, mockAlerts["existing_a1"].LastActive, nowTime)

	// test new active alert - suppressed
	rule := models.NewSuppRule(models.Labels{"device": "d2"}, models.MatchCond_ALL, "", "", time.Duration(1*time.Minute))
	h.Suppressor.SaveRule(ctx, tx, rule)
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	a2.ExtendLabels()
	assert.NotNil(t, h.Suppressor.Match(a2.Labels))

	// test new active alert - suppressed
	rule2 := models.NewSuppRule(models.Labels{"description": "PE Error code: 0x2104be"}, models.MatchCond_ALL, "", "", time.Duration(1*time.Minute))
	h.Suppressor.SaveRule(ctx, tx, rule2)
	a4 := tu.MockAlert(100, "Test Alert 2", "QFX PE Error code: 0x2104be test", "dev3", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	a4.ExtendLabels()
	assert.NotNil(t, h.Suppressor.Match(a4.Labels))
}

func TestHandlerAlertActiveDedup(t *testing.T) {
	h := NewTestHandler(2)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test a de-dedup of a cleared alert
	tx.(*MockTx).updateAlert = func(alert *models.Alert) error {
		for _, a := range mockAlerts {
			if a.Id == alert.Id {
				a.LastActive = nowTime
				a.StartTime = nowTime
				break
			}
		}
		return nil
	}
	new := tu.MockAlert(0, "Test Alert 5", "", "d5", "e5", "src5", "scp5", "t1", "5", "WARN", []string{"g", "h"}, nil)
	mockAlerts["existing_a5"].Status = models.Status_CLEARED
	mockAlerts["existing_a6_agg"].Status = models.Status_CLEARED
	mockAlerts["existing_a5"].AggregatorId = 600
	err := h.handleActive(ctx, tx, new)
	assert.Nil(t, err)
	assert.Equal(t, mockAlerts["existing_a5"].Status, models.Status_ACTIVE)
	assert.Equal(t, mockAlerts["existing_a5"].LastActive, nowTime)
	assert.Equal(t, mockAlerts["existing_a6_agg"].Status, models.Status_ACTIVE)
	assert.Equal(t, mockAlerts["existing_a6_agg"].LastActive, nowTime)
	assert.Equal(t, mockAlerts["existing_a6_agg"].StartTime, nowTime)
	event := <-h.procChan
	assert.Equal(t, event.Alert.Id, int64(600))

	// test dedup of cleared alert - active supprule
	rule3 := models.NewSuppRule(models.Labels{"device": "d5"}, models.MatchCond_ALL, "", "", time.Duration(1*time.Minute))
	h.Suppressor.SaveRule(ctx, tx, rule3)
	mockAlerts["existing_a5"].ExtendLabels()
	mockAlerts["existing_a5"].Status = models.Status_CLEARED
	mockAlerts["existing_a6_agg"].Status = models.Status_CLEARED
	err = h.handleActive(ctx, tx, new)
	assert.Nil(t, err)
	assert.Equal(t, mockAlerts["existing_a5"].Status, models.Status_CLEARED)
	assert.Equal(t, mockAlerts["existing_a6_agg"].Status, models.Status_CLEARED)

	// test dedup of cleared alert - disabled by config
	tx.(*MockTx).newInsert = func(query string, item interface{}) (int64, error) {
		return 999, nil
	}
	new = tu.MockAlert(0, "Test Alert 7", "", "d7", "e7", "src7", "scp7", "t1", "7", "WARN", []string{"g", "h"}, nil)
	mockAlerts["existing_a7"].Status = models.Status_CLEARED
	err = h.handleActive(ctx, tx, new)
	assert.Nil(t, err)
	assert.Equal(t, int(new.Id), 999)
	event = <-h.procChan
	assert.Equal(t, event.Type, models.EventType_ACTIVE)
}

func TestHandlerAlertClear(t *testing.T) {
	h := NewTestHandler(1)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test alert clear -non existing
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	h.handleClear(ctx, tx, a2)
	assert.Equal(t, a2.Status.String(), "ACTIVE")

	// test alert clear - no autoclear
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	h.handleClear(ctx, tx, a1)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "ACTIVE")

	// test alert clear - autoclear
	mockAlerts["existing_a1"].AutoClear = true
	h.handleClear(ctx, tx, a1)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "CLEARED")
	event := <-h.procChan
	assert.Equal(t, event.Type, models.EventType_CLEARED)
	assert.Equal(t, int(event.Alert.Id), 100)

	// test alert clear - ack'd
	mockAlerts["existing_a3"].AutoClear = true
	mockAlerts["existing_a3"].Owner.Valid = true
	h.handleClear(ctx, tx, mockAlerts["existing_a3"])
	assert.Equal(t, mockAlerts["existing_a3"].Status.String(), "ACTIVE")
}

func TestHandlerAlertExpiry(t *testing.T) {
	h := NewTestHandler(1)
	ctx := context.Background()

	h.handleExpiry(ctx)

	event := <-h.procChan
	assert.Equal(t, event.Alert.Status.String(), "EXPIRED")
	assert.Equal(t, event.Type, models.EventType_EXPIRED)
	assert.Equal(t, int(event.Alert.Id), 300)
}

func TestHandlerAlertEscalate(t *testing.T) {
	h := NewTestHandler(2)
	ctx := context.Background()

	// test no escalation needed
	mockAlerts["existing_a4"].StartTime = models.MyTime{time.Now()}
	h.handleEscalation(ctx)

	// test first level escalation
	mockAlerts["existing_a4"].StartTime = models.MyTime{}
	h.handleEscalation(ctx)
	event := <-h.procChan
	assert.Equal(t, event.Alert.Severity.String(), "WARN")
	assert.Equal(t, event.Type, models.EventType_ESCALATED)
	mockAlerts["existing_a4"] = event.Alert

	// test second level escalation
	h.handleEscalation(ctx)
	event = <-h.procChan
	assert.Equal(t, event.Alert.Severity.String(), "CRITICAL")
	assert.Equal(t, event.Type, models.EventType_ESCALATED)
}

func TestHandlerAlertSuppress(t *testing.T) {
	h := NewTestHandler(3)
	tx := h.Db.NewTx()
	ctx := context.Background()
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	assert.Nil(t, h.Suppress(ctx, tx, a1, "test", "test", 1*time.Minute, true))
	event := <-h.procChan
	assert.Equal(t, event.Type, models.EventType_SUPPRESSED)
	assert.Equal(t, a1.Status, models.Status_SUPPRESSED)
	a1 = tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	a1.ExtendLabels()
	assert.NotNil(t, h.Suppressor.Match(a1.Labels))

	// test agg suppress
	a2 := tu.MockAlert(100, "Test Agg Alert 2", "", "d2", "e2", "src2", "aggregated", "t1", "2", "WARN", []string{"c", "d"}, nil)
	a2.IsAggregate = true
	mockAlerts["existing_a1"].AggregatorId = a2.Id
	mockAlerts["existing_a2"].AggregatorId = a2.Id
	assert.Nil(t, h.Suppress(ctx, tx, a2, "test", "test", 1*time.Minute, true))
	assert.Equal(t, a2.Status, models.Status_SUPPRESSED)
	event = <-h.procChan
	assert.Equal(t, event.Type, models.EventType_SUPPRESSED)
	assert.Equal(t, event.Alert.Status, models.Status_SUPPRESSED)
	event = <-h.procChan
	assert.Equal(t, event.Type, models.EventType_SUPPRESSED)
	assert.Equal(t, event.Alert.Status, models.Status_SUPPRESSED)

	a1 = tu.MockAlert(100, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	a1.ExtendLabels()
	assert.NotNil(t, h.Suppressor.Match(a1.Labels))
}

func TestMain(m *testing.M) {
	AddTransform(&mockTransform{name: "mock", priority: 100, register: "New*"})
	plugins.AddProcessor(&mockProcessor{})
	flag.Parse()
	Config = NewConfigHandler("../testutil/testdata/test_config.yaml")
	Config.LoadConfig()
	os.Exit(m.Run())
}
