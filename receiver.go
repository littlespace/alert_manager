package alert_manager

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type Processor interface {
	Name() string
	Start(ctx context.Context, db *models.DB)
}

type Output interface {
	Name() string
	Start(ctx context.Context)
}
