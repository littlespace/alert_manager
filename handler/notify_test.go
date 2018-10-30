package handler

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNotify(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 4", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{}, nil)
	event := &AlertEvent{Type: EventType_ACTIVE, Alert: mockAlert}
	notif := &notifier{notifiedAlerts: make(map[int64]*notification)}
	notifyChan := make(chan *AlertEvent, 1)
	RegisterOutput("slack", notifyChan)
	tx := &MockTx{}

	// test notify delay
	notif.Notify(event, tx)
	assert.Equal(t, len(notif.notifiedAlerts), 0)

	// test ackd alert
	event.Alert.Owner.Valid = true
	notif.Notify(event, tx)
	assert.Equal(t, len(notif.notifiedAlerts), 0)
	event.Alert.Owner.Valid = false

	// test first notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event, tx)
	recvd := <-notifyChan
	assert.Equal(t, recvd.Type, EventType_ACTIVE)
	assert.Equal(t, recvd.Alert, mockAlert)
	lastNotified := notif.notifiedAlerts[1].lastNotified

	// test second notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event, tx)
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// test clear notify
	event = &AlertEvent{Type: EventType_CLEARED, Alert: mockAlert}
	notif.Notify(event, tx)
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, EventType_CLEARED)
	assert.Equal(t, recvd.Alert, mockAlert)
}

func TestNotifyReminder(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 4", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{}, nil)
	event := &AlertEvent{Type: EventType_ACTIVE, Alert: mockAlert}
	notif := &notifier{notifiedAlerts: make(map[int64]*notification)}
	notifyChan := make(chan *AlertEvent, 1)
	RegisterOutput("slack", notifyChan)
	tx := &MockTx{}

	// first notif
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event, tx)
	recvd := <-notifyChan
	assert.Equal(t, recvd.Type, EventType_ACTIVE)
	assert.Equal(t, recvd.Alert, mockAlert)
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
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, EventType_ACTIVE)
	assert.Equal(t, recvd.Alert, mockAlert)

	// escalate and remind
	mockAlert.Severity = models.Sev_CRITICAL
	event = &AlertEvent{Type: EventType_ESCALATED, Alert: mockAlert}
	notif.Notify(event, tx)
	notif.notifiedAlerts[mockAlert.Id].lastNotified = time.Now().Add(-20 * time.Minute)
	notif.remind()
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, EventType_ESCALATED)
	assert.Equal(t, recvd.Alert, mockAlert)
	lastNotified = notif.notifiedAlerts[1].lastNotified

	// alert suppressed - no remind
	mockAlert.Suppress(30 * time.Minute)
	notif.Notify(event, tx)
	notif.remind()
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// alert expired - no remind
	mockAlert.Status = models.Status_EXPIRED
	event = &AlertEvent{Type: EventType_EXPIRED, Alert: mockAlert}
	notif.Notify(event, tx)
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, EventType_EXPIRED)
	assert.Equal(t, recvd.Alert, mockAlert)

	assert.Equal(t, len(notif.notifiedAlerts), 0)
	notif.remind()
}
