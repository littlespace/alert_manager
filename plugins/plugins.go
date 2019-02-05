package plugins

import (
	"context"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/listener"
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

type Processor interface {
	Name() string
	Start(ctx context.Context, db models.Dbase)
}

type Output interface {
	Name() string
	Start(ctx context.Context)
}

type ApiPlugins struct {
	Parsers    []string `json:"parsers"`
	Processors []string `json:"processors"`
	Outputs    []string `json:"outputs"`
	Listeners  []string `json:"listeners"`
}

var (
	Listeners  = make(map[string]Listener)
	Processors = make(map[string]Processor)
	Outputs    = make(map[string]Output)
)

func AddListener(l Listener) {
	Listeners[l.Name()] = l
}

func AddProcessor(p Processor) {
	Processors[p.Name()] = p
}

func AddOutput(o Output) {
	Outputs[o.Name()] = o
}

func Init(ctx context.Context, db models.Dbase) error {

	// start all the listeners
	for name, listener := range Listeners {
		glog.Infof("Starting Listener: %s on %s", name, listener.Uri())
		go listener.Listen(ctx)
	}

	// start all the processors/outputs
	for name, processor := range Processors {
		glog.Infof("Starting processor: %s", name)
		go processor.Start(ctx, db)
	}

	for name, output := range Outputs {
		glog.Infof("Starting output: %s", name)
		go output.Start(ctx)
	}

	return nil
}

func GetApiPluginsList() ApiPlugins {

	choices := ApiPlugins{
		Parsers:    listener.GetParsersList(),
		Processors: make([]string, 0, len(Processors)),
		Outputs:    make([]string, 0, len(Outputs)),
		Listeners:  make([]string, 0, len(Listeners)),
	}

	for k := range Processors {
		choices.Processors = append(choices.Processors, k)
	}

	for k := range Outputs {
		choices.Outputs = append(choices.Outputs, k)
	}

	for k := range Listeners {
		choices.Listeners = append(choices.Listeners, k)
	}

	return choices
}
