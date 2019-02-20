package handler

import (
	"context"
	"flag"
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var mockAlerts = map[string]*models.Alert{
	"existing_a1": tu.MockAlert(100, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil),
	"existing_a2": tu.MockAlert(200, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil),
	"existing_a3": tu.MockAlert(300, "Test Alert 3", "", "d3", "e3", "src3", "scp3", "t1", "3", "WARN", []string{"e", "f"}, nil),
	"existing_a4": tu.MockAlert(400, "Test Alert 4", "", "d4", "e4", "src4", "scp4", "t1", "4", "INFO", []string{"e", "f"}, nil),
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
}

func (t *MockTx) NewInsert(query string, item interface{}) (int64, error) {
	switch item.(type) {
	case *models.Alert:
		alert := item.(*models.Alert)
		if alert.Name == "Test Alert 1" {
			return 100, nil
		}
		return 200, nil
	case *models.Team:
		return 1, nil
	}
	return 0, nil
}

func (t *MockTx) UpdateAlert(alert *models.Alert) error {
	return nil
}

func (t *MockTx) Rollback() error {
	return nil
}

func (t *MockTx) Commit() error {
	return nil
}

func (t *MockTx) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	if args[0].(string) == "Test Alert 1" {
		return mockAlerts["existing_a1"], nil
	} else {
		return nil, fmt.Errorf("No alert found")
	}
}

func (t *MockTx) InQuery(query string, args ...interface{}) error {
	alertIDs := args[1].([]int64)
	var alert *models.Alert
	switch alertIDs[0] {
	case 100:
		alert = mockAlerts["existing_a1"]
	case 200:
		alert = mockAlerts["existing_a2"]
	}
	alert.LastActive = nowTime
	return nil
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
	h.clearer = &ClearHandler{actives: make(map[int64]chan struct{})}
	h.Suppressor = &suppressor{db: m}
	return h
}

func TestHandlerAlertActive(t *testing.T) {
	h := NewTestHandler(2)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test new active alert
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	h.handleActive(ctx, tx, a2)
	assert.Equal(t, int(a2.Id), 200)
	assert.Equal(t, h.teams[0].Name, "t1")
	assert.Equal(t, h.teams[0].Id, int64(1))
	l := models.Labels{"suppress": "me", "device": "d2", "entity": "e2", "source": "src2", "scope": "scp2", "alert_name": "Test Alert 2"}
	assert.Equal(t, a2.Labels, l)
	event := <-h.procChan
	assert.Equal(t, event.Type, models.EventType_ACTIVE)
	assert.Equal(t, int(event.Alert.Id), 200)

	// test new team
	a3 := tu.MockAlert(0, "Test Alert 6", "", "d6", "e6", "src6", "scp6", "t2", "2", "WARN", []string{"c", "d"}, nil)
	h.handleActive(ctx, tx, a3)
	assert.Equal(t, h.teams.Contains("t2"), true)

	// test existing active alert
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	h.handleActive(ctx, tx, a1)
	assert.Equal(t, int(a1.Id), 0)
	assert.Equal(t, mockAlerts["existing_a1"].LastActive, nowTime)

	// test new active alert - suppressed
	rule := models.NewSuppRule(models.Labels{"device": "d2"}, models.MatchCond_ALL, "", "", time.Duration(1*time.Minute))
	h.Suppressor.SaveRule(ctx, tx, rule)
	a2 = tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	a2.ExtendLabels()
	assert.NotNil(t, h.Suppressor.Match(a2.Labels))
}

func TestHandlerAlertClear(t *testing.T) {
	h := NewTestHandler(1)
	tx := h.Db.NewTx()
	ctx := context.Background()

	// test alert clear -non existing
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "t1", "2", "WARN", []string{"c", "d"}, nil)
	h.handleClear(ctx, tx, a2, 0)
	assert.Equal(t, a2.Status.String(), "ACTIVE")

	// test alert clear - no autoclear
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{"a", "b"}, nil)
	h.handleClear(ctx, tx, a1, 0)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "ACTIVE")

	// test alert clear - autoclear
	mockAlerts["existing_a1"].AutoClear = true
	h.handleClear(ctx, tx, a1, 0)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "CLEARED")
	event := <-h.procChan
	assert.Equal(t, event.Type, models.EventType_CLEARED)
	assert.Equal(t, int(event.Alert.Id), 100)
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
	assert.Nil(t, h.Suppress(ctx, tx, a1, "test", "test", 1*time.Minute))
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
	assert.Nil(t, h.Suppress(ctx, tx, a2, "test", "test", 1*time.Minute))
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
	AddTransform(&mockTransform{name: "mock", priority: 100, register: "Test Alert 2"})
	plugins.AddProcessor(&mockProcessor{})
	flag.Parse()
	Config = NewConfigHandler("../testutil/testdata/test_config.yaml")
	Config.LoadConfig()
	os.Exit(m.Run())
}
