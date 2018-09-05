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
}

var nowTime = models.MyTime{time.Now()}

type MockDb struct{}

func (m *MockDb) NewTx() models.Txn {
	return &MockTx{}
}

func (m *MockDb) Close() error {
	return nil
}

type MockTx struct{}

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

func (t *MockTx) SelectAlerts(query string) (models.Alerts, error) {
	return models.Alerts{}, nil
}

func (t *MockTx) SelectRules(query string) (models.SuppRules, error) {
	return models.SuppRules{}, nil
}

type mockTransform struct {
	name     string
	priority int
	register string
}

func (t *mockTransform) Name() string {
	return t.name
}

func (t *mockTransform) GetPriority() int {
	return t.priority
}

func (t *mockTransform) GetRegister() string {
	return t.register
}

func (t *mockTransform) Apply(a *models.Alert) error {
	return a.AddMeta("test")
}

func TestHandlerAlertActive(t *testing.T) {
	m := &MockDb{}
	tx := m.NewTx()
	h := NewHandler(m)
	h.suppRules = models.SuppRules{}
	ctx := context.Background()

	// test new active alert
	a2 := tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil)
	notif, _ := h.handleActive(ctx, tx, a2)
	assert.Equal(t, notif, true)
	assert.Equal(t, int(a2.Id), 200)
	assert.Equal(t, a2.Metadata.String, "\"test\"")

	// test existing active alert
	a1 := tu.MockAlert(0, "Test Alert 1", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil)
	notif, _ = h.handleActive(ctx, tx, a1)
	assert.Equal(t, notif, false)
	assert.Equal(t, int(a1.Id), 0)
	assert.Equal(t, mockAlerts["existing_a1"].LastActive, nowTime)

	//// test new active alert - suppressed
	h.suppRules = append(h.suppRules, models.NewSuppRuleForAlert("Test Alert 2", "me", "test", "5m"))
	a2 = tu.MockAlert(0, "Test Alert 2", "", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil)
	notif, _ = h.handleActive(ctx, tx, a2)
	assert.Equal(t, notif, false)
	assert.Equal(t, int(a2.Id), 200)
	assert.Equal(t, a2.Status.String(), "SUPPRESSED")

}

func TestMain(m *testing.M) {
	Transforms = []Transform{}
	Processors = make(map[string][]chan *AlertEvent)
	Outputs = make(map[string]chan *AlertEvent)
	AddTransform(&mockTransform{name: "mock", priority: 100, register: "Test Alert"})
	flag.Parse()
	Config = NewConfigHandler("../testutil/testdata/test_config.yaml")
	Config.LoadConfig()
	os.Exit(m.Run())
}
