package testutil

import (
	"flag"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"time"
)

type MockStat struct{}

func (m *MockStat) Add(value int64) {}
func (m *MockStat) Set(value int64) {}
func (m *MockStat) Reset()          {}

func MockAlertEvent(eType ah.EventType, name, desc, sev, device, entity, source, scope, extId string) *ah.AlertEvent {
	var status models.AlertStatus
	switch eType {
	case ah.EventType_ACTIVE:
		status = models.Status_ACTIVE
	case ah.EventType_CLEARED:
		status = models.Status_CLEARED
	}
	a := &models.Alert{
		Name:        name,
		Description: desc,
		Severity:    models.SevMap[sev],
		Status:      status,
		Entity:      entity,
		Source:      source,
		Scope:       scope,
		ExternalId:  extId,
	}
	a.AddDevice(device)
	return &ah.AlertEvent{Type: eType, Alert: a}
}

func MockAlert(id int64, name, device, entity, source, scope, extId string, tags []string, meta interface{}) *models.Alert {
	a := &models.Alert{
		Id:         id,
		Name:       name,
		Status:     models.Status_ACTIVE,
		Entity:     entity,
		Source:     source,
		Scope:      scope,
		ExternalId: extId,
		StartTime:  models.MyTime{time.Now()},
	}
	a.AddDevice(device)
	a.AddTags(tags...)
	a.AddMeta(meta)
	return a
}

func InitTests() {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
}
