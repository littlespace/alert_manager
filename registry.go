package alert_manager

import (
	"context"
)

// Listener is any agent that listens to alerts.
type Listener interface {
	// descriptive name or type
	Name() string
	// The URL to send requests to
	Uri() string
	// Start listening for alerts
	Listen(ctx context.Context)
}

// Output is any agent that sends alerts to external receiver
type Output interface {
	Name() string
	Start(ctx context.Context)
}

var (
	Listeners = make(map[string]Listener)
	Outputs   = make(map[string]Output)
)

func AddListener(l Listener) {
	Listeners[l.Name()] = l
}

func AddOutput(o Output) {
	Outputs[o.Name()] = o
}
