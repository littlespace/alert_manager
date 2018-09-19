package alert_manager

import (
	"context"
	ah "github.com/mayuresh82/alert_manager/handler"
)

type Processor interface {
	Name() string
	Start(ctx context.Context, h *ah.AlertHandler)
}

type Output interface {
	Name() string
	Start(ctx context.Context)
}
