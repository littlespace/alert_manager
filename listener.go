package alert_manager

import (
	"context"
)

// Listener is any agent that listens to alerts. Alerts are sent down a channel that
// processors or outputs can share.
type Listener interface {
	// descriptive name or type
	Name() string
	// The URL to send requests to
	Uri() string
	// Start listening for alerts
	Listen(ctx context.Context)
}
