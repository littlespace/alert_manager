package handler

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNotify(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 5", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{}, nil)
	event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: mockAlert}
	db := &MockDb{}
	notif = &notifier{notifiedAlerts: make(map[int64]*notification), db: db}
	notifyChan := make(chan *models.AlertEvent, 1)
	RegisterOutput("slack", notifyChan)

	// test notify delay
	notif.Notify(event)
	assert.Equal(t, len(notif.notifiedAlerts), 0)

	// test first notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	recvd := <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_ACTIVE)
	assert.Equal(t, recvd.Alert, mockAlert)
	lastNotified := notif.notifiedAlerts[1].lastNotified

	// test second notification
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	assert.Equal(t, notif.notifiedAlerts[1].lastNotified.Equal(lastNotified), true)

	// test clear notify
	event = &models.AlertEvent{Type: models.EventType_CLEARED, Alert: mockAlert}
	notif.Notify(event)
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_CLEARED)
	assert.Equal(t, recvd.Alert, mockAlert)
}

func TestNotifyReminder(t *testing.T) {
	mockAlert := tu.MockAlert(1, "Test Alert 5", "", "d1", "e1", "src1", "scp1", "1", "WARN", []string{}, nil)
	event := &models.AlertEvent{Type: models.EventType_ACTIVE, Alert: mockAlert}
	db := &MockDb{}
	notif = &notifier{notifiedAlerts: make(map[int64]*notification), db: db}
	notifyChan := make(chan *models.AlertEvent, 1)
	RegisterOutput("slack", notifyChan)

	// first notif
	mockAlert.LastActive.Time = mockAlert.LastActive.Add(10 * time.Minute)
	notif.Notify(event)
	recvd := <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_ACTIVE)
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
	assert.Equal(t, recvd.Type, models.EventType_ACTIVE)
	assert.Equal(t, recvd.Alert, mockAlert)

	// escalate and remind
	mockAlert.Severity = models.Sev_CRITICAL
	event = &models.AlertEvent{Type: models.EventType_ESCALATED, Alert: mockAlert}
	notif.Notify(event)
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_ESCALATED)
	assert.Equal(t, recvd.Alert, mockAlert)
	notif.notifiedAlerts[mockAlert.Id].lastNotified = time.Now().Add(-20 * time.Minute)
	notif.remind()
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_ESCALATED)
	assert.Equal(t, recvd.Alert, mockAlert)
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
	recvd = <-notifyChan
	assert.Equal(t, recvd.Type, models.EventType_EXPIRED)
	assert.Equal(t, recvd.Alert, mockAlert)

	assert.Equal(t, len(notif.notifiedAlerts), 0)
	notif.remind()
}
