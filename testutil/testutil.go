package testutil

import (
	"database/sql"
	"flag"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type MockStat struct{}

func (m *MockStat) Add(value int64) {}
func (m *MockStat) Set(value int64) {}
func (m *MockStat) Reset()          {}

func MockAlertEvent(eType ah.EventType, name, device, entity, source, scope, extId string) *ah.AlertEvent {
	var status models.AlertStatus
	switch eType {
	case ah.EventType_ACTIVE:
		status = models.Status_ACTIVE
	case ah.EventType_CLEARED:
		status = models.Status_CLEARED
	}
	return &ah.AlertEvent{
		Type: eType,
		Alert: &models.Alert{
			Name:       name,
			Status:     status,
			Device:     sql.NullString{device, true},
			Entity:     entity,
			Source:     source,
			Scope:      scope,
			ExternalId: extId,
		},
	}
}

func InitTests() {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
}
