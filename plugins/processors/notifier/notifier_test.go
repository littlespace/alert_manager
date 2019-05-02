package notifier

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

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

func (t *MockTx) Rollback() error {
	return nil
}

func (t *MockTx) Commit() error {
	return nil
}

func (t *MockTx) NewRecord(alertId int64, event string) (int64, error) {
	return 1, nil
}

type MockOutput struct{}

func (m *MockOutput) Name() string { return "slack" }

func (m *MockOutput) Start(ctx context.Context, opts *plugins.Options) {}

func TestNotify(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 5", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{}, nil)
	mockAlert.ExtendLabels()
	event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: mockAlert}
	db := &MockDb{}
	notif := &Notifier{notifiedAlerts: make(map[int64]*notification), db: db}
	notifyChan := make(chan *plugins.SendRequest, 3)
	plugins.AddOutput(&MockOutput{}, notifyChan)

	// test notify delay
	notif.Notify(event)
	assert.Equal(t, len(notif.notifiedAlerts), 0)

	// test first notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	req := <-notifyChan
	assert.Equal(t, req.Name, "default")
	assert.Equal(t, req.Event.Type, models.EventType_ACTIVE)
	assert.Equal(t, req.Event.Alert, mockAlert)
	lastNotified := notif.notifiedAlerts[1].lastNotified

	// test second notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// test escalated
	mockAlert.SetSeverity(models.Sev_CRITICAL)
	event = &models.AlertEvent{Type: models.EventType_ESCALATED, Alert: mockAlert}
	notif.Notify(event)
	req = <-notifyChan
	assert.Equal(t, req.Name, "test1")
	assert.Equal(t, req.Event.Type, models.EventType_ESCALATED)

	// test clear notify
	event = &models.AlertEvent{Type: models.EventType_CLEARED, Alert: mockAlert}
	notif.Notify(event)
	req = <-notifyChan
	assert.Equal(t, req.Event.Type, models.EventType_CLEARED)
	assert.Equal(t, req.Event.Alert, mockAlert)
}

func TestNotifyReminder(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 5", "", "d1", "e1", "src1", "scp1", "t1", "1", "WARN", []string{}, nil)
	mockAlert.ExtendLabels()
	event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: mockAlert}
	db := &MockDb{}
	notif := &Notifier{notifiedAlerts: make(map[int64]*notification), db: db}
	notifyChan := make(chan *plugins.SendRequest, 5)
	plugins.AddOutput(&MockOutput{}, notifyChan)

	// first notif
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	req := <-notifyChan
	assert.Equal(t, req.Event.Type, models.EventType_ACTIVE)
	assert.Equal(t, req.Event.Alert, mockAlert)
	lastNotified := notif.notifiedAlerts[1].lastNotified

	// not remind
	notif.remind()
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// ackd alert - not remind
	event.Alert.Owner.Valid = true
	notif.remind()
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)
	event.Alert.Owner.Valid = false

	// remind
	notif.notifiedAlerts[mockAlert.Id].lastNotified = time.Now().Add(-20 * time.Minute)
	notif.remind()
	req = <-notifyChan
	assert.Equal(t, req.Name, "default")
	assert.Equal(t, req.Event.Type, models.EventType_ACTIVE)
	assert.Equal(t, req.Event.Alert, mockAlert)

	// escalate and remind
	mockAlert.SetSeverity(models.Sev_CRITICAL)
	event = &models.AlertEvent{Type: models.EventType_ESCALATED, Alert: mockAlert}
	notif.Notify(event)
	req = <-notifyChan
	assert.Equal(t, req.Name, "test1")
	assert.Equal(t, req.Event.Type, models.EventType_ESCALATED)
	assert.Equal(t, req.Event.Alert, mockAlert)
	notif.notifiedAlerts[mockAlert.Id].lastNotified = time.Now().Add(-20 * time.Minute)
	notif.remind()
	req = <-notifyChan
	assert.Equal(t, req.Event.Type, models.EventType_ESCALATED)
	assert.Equal(t, req.Event.Alert, mockAlert)
	lastNotified = notif.notifiedAlerts[1].lastNotified

	// alert suppressed - no remind
	mockAlert.Suppress(30 * time.Minute)
	event = &models.AlertEvent{Type: models.EventType_SUPPRESSED, Alert: mockAlert}
	notif.Notify(event)
	notif.remind()
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// alert expired - no remind
	mockAlert.Status = models.Status_EXPIRED
	event = &models.AlertEvent{Type: models.EventType_EXPIRED, Alert: mockAlert}
	notif.Notify(event)
	req = <-notifyChan
	assert.Equal(t, req.Event.Type, models.EventType_EXPIRED)
	assert.Equal(t, req.Event.Alert, mockAlert)

	assert.Equal(t, len(notif.notifiedAlerts), 0)
	notif.remind()
}

func TestMain(m *testing.M) {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../../../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
