package testutil

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	"time"
)

type MockStat struct{}

func (m *MockStat) Add(value int64) {}
func (m *MockStat) Set(value int64) {}
func (m *MockStat) Reset()          {}

func MockAlert(id int64, name, desc, device, entity, source, scope, extId, sev string, tags []string, meta interface{}) *models.Alert {
	a := &models.Alert{
		Id:          id,
		Name:        name,
		Description: desc,
		Status:      models.Status_ACTIVE,
		Severity:    models.SevMap[sev],
		Entity:      entity,
		Source:      source,
		Scope:       scope,
		ExternalId:  extId,
		StartTime:   models.MyTime{time.Now()},
	}
	a.AddDevice(device)
	a.AddTags(tags...)
	if meta != nil {
		a.AddMeta(meta)
	}
	return a
}
