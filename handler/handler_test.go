package handler

import (
	"context"
	"flag"
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var mockAlerts = map[string]*models.Alert{
	"existing_a1": tu.MockAlert(100, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil),
	"existing_a2": tu.MockAlert(200, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil),
	"existing_a3": tu.MockAlert(300, "Test Alert 3", "", "d3", "e3", "src3", "scp3", "3", "WARN", []string{"e", "f"}, nil),
	"existing_a4": tu.MockAlert(400, "Test Alert 4", "", "d4", "e4", "src4", "scp4", "4", "INFO", []string{"e", "f"}, nil),
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

func (t *MockTx) NewAlert(alert *models.Alert) (int64, error) {
	if alert.Name == "Test Alert 1" {
		return 100, nil
	}
	return 200, nil
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

func (t *MockTx) NewSuppRule(rule *models.SuppressionRule) (int64, error) {
	return 1, nil
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

func (m *mockProcessor) Process(ctx context.Context, db models.Dbase, in chan *AlertEvent) chan *AlertEvent {
	return make(chan *AlertEvent)
}

func TestHandlerAlertActive(t *testing.T) {
	m := &MockDb{}
	tx := m.NewTx()
	h := &AlertHandler{Db: m, statTransformError: &tu.MockStat{}, statDbError: &tu.MockStat{}}
	h.procChan = make(chan *AlertEvent, 3)
	h.clearer = &ClearHandler{actives: make(map[int64]chan struct{})}
	h.Suppressor = &suppressor{db: m}
	h.Notifier = &notifier{notifiedAlerts: make(map[int64]*notification), db: m}
	ctx := context.Background()

	// test new active alert
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil)
	h.handleActive(ctx, tx, a2)
	assert.Equal(t, int(a2.Id), 200)
	assert.Equal(t, a2.Labels, models.Labels{"suppress": "me"})
	event := <-h.procChan
	assert.Equal(t, event.Type, EventType_ACTIVE)
	assert.Equal(t, int(event.Alert.Id), 200)

	// test existing active alert
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil)
	h.handleActive(ctx, tx, a1)
	assert.Equal(t, int(a1.Id), 0)
	assert.Equal(t, mockAlerts["existing_a1"].LastActive, nowTime)
	<-h.procChan

	// test new active alert - suppressed
	rule := models.NewSuppRule(models.Labels{"device": "d2"}, models.MatchCond_ALL, "", "", time.Duration(1*time.Minute))
	h.Suppressor.SaveRule(ctx, tx, rule)
	a2 = tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil)
	h.handleActive(ctx, tx, a2)
}

func TestHandlerAlertClear(t *testing.T) {
	m := &MockDb{}
	tx := m.NewTx()
	h := &AlertHandler{Db: m, statTransformError: &tu.MockStat{}, statDbError: &tu.MockStat{}}
	h.procChan = make(chan *AlertEvent, 1)
	h.clearer = &ClearHandler{actives: make(map[int64]chan struct{})}
	h.Suppressor = &suppressor{db: m}
	h.Notifier = &notifier{notifiedAlerts: make(map[int64]*notification), db: m}
	ctx := context.Background()

	// test alert clear -non existing
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil)
	h.handleClear(ctx, tx, a2, 0)
	assert.Equal(t, a2.Status.String(), "ACTIVE")

	// test alert clear - no autoclear
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil)
	h.handleClear(ctx, tx, a1, 0)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "ACTIVE")

	// test alert clear - autoclear
	mockAlerts["existing_a1"].AutoClear = true
	h.handleClear(ctx, tx, a1, 0)
	assert.Equal(t, mockAlerts["existing_a1"].Status.String(), "CLEARED")
	event := <-h.procChan
	assert.Equal(t, event.Type, EventType_CLEARED)
	assert.Equal(t, int(event.Alert.Id), 100)
}

func TestHandlerAlertExpiry(t *testing.T) {
	m := &MockDb{}
	h := &AlertHandler{Db: m, statTransformError: &tu.MockStat{}, statDbError: &tu.MockStat{}}
	h.procChan = make(chan *AlertEvent, 1)
	h.clearer = &ClearHandler{actives: make(map[int64]chan struct{})}
	h.Suppressor = &suppressor{db: m}
	h.Notifier = &notifier{notifiedAlerts: make(map[int64]*notification), db: m}
	ctx := context.Background()

	h.handleExpiry(ctx)

	event := <-h.procChan
	assert.Equal(t, event.Alert.Status.String(), "EXPIRED")
	assert.Equal(t, event.Type, EventType_EXPIRED)
	assert.Equal(t, int(event.Alert.Id), 300)
}

func TestHandlerAlertEscalate(t *testing.T) {
	m := &MockDb{}
	h := &AlertHandler{Db: m, statTransformError: &tu.MockStat{}, statDbError: &tu.MockStat{}}
	h.procChan = make(chan *AlertEvent, 2)
	h.clearer = &ClearHandler{actives: make(map[int64]chan struct{})}
	h.Suppressor = &suppressor{db: m}
	h.Notifier = &notifier{notifiedAlerts: make(map[int64]*notification), db: m}
	ctx := context.Background()

	// test no escalation needed
	mockAlerts["existing_a4"].StartTime = models.MyTime{time.Now()}
	h.handleEscalation(ctx)

	// test first level escalation
	mockAlerts["existing_a4"].StartTime = models.MyTime{}
	h.handleEscalation(ctx)
	event := <-h.procChan
	assert.Equal(t, event.Alert.Severity.String(), "WARN")
	assert.Equal(t, event.Type, EventType_ESCALATED)
	mockAlerts["existing_a4"] = event.Alert

	// test second level escalation
	h.handleEscalation(ctx)
	event = <-h.procChan
	assert.Equal(t, event.Alert.Severity.String(), "CRITICAL")
	assert.Equal(t, event.Type, EventType_ESCALATED)
}

func TestMain(m *testing.M) {
	AddTransform(&mockTransform{name: "mock", priority: 100, register: "Test Alert 2"})
	AddProcessor(&mockProcessor{})
	flag.Parse()
	Config = NewConfigHandler("../testutil/testdata/test_config.yaml")
	Config.LoadConfig()
	os.Exit(m.Run())
}
