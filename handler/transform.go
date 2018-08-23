package handler

import (
	"github.com/mayuresh82/alert_manager/internal/models"
)

type Transform interface {
	Name() string
	GetPriority() int
	GetRegister() string
	Apply(alert *models.Alert) error
}
