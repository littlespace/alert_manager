package testutil

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	"time"
)

type MockStat struct{}

func (m *MockStat) Add(value int64) {}
func (m *MockStat) Set(value int64) {}
func (m *MockStat) Reset()          {}

func MockAlert(id int64, name, desc, device, entity, source, scope, team, extId, sev string, tags []string, labels models.Labels) *models.Alert {
	start := models.MyTime{time.Now()}
	a := &models.Alert{
		Id:          id,
		Name:        name,
		Description: desc,
		Status:      models.Status_ACTIVE,
		Severity:    models.SevMap[sev],
		Entity:      entity,
		Source:      source,
		Scope:       scope,
		Team:        team,
		ExternalId:  extId,
		StartTime:   start,
		LastActive:  start,
		Labels:      make(models.Labels),
	}
	a.AddDevice(device)
	a.AddTags(tags...)
	if labels != nil {
		a.Labels = labels
	}
	return a
}
