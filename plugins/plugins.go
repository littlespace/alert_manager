package plugins

import (
	"context"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"time"
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
	GetParsersList() []string
}

// Processor is an alert processor thats part of a processor pipeline
type Processor interface {
	Name() string
	Stage() int
	Process(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) chan *models.AlertEvent
}

func GetProcessor(name string) Processor {
	for _, p := range Processors {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

type Output interface {
	Name() string
	Start(ctx context.Context, opts *Options)
}

type ApiPlugins struct {
	Parsers    []string `json:"parsers"`
	Processors []string `json:"processors"`
	Outputs    []string `json:"outputs"`
	Listeners  []string `json:"listeners"`
}

var (
	Listeners  = make(map[string]Listener)
	Processors []Processor
	Outputs    = make(map[string]Output)
)

func AddListener(l Listener) {
	Listeners[l.Name()] = l
}

func AddProcessor(p Processor) {
	Processors = append(Processors, p)
}

func AddOutput(o Output) {
	Outputs[o.Name()] = o
}

func Init(ctx context.Context, db models.Dbase, options ...PluginOption) error {

	opts := &Options{
		WebUrl:        "http://localhost",
		ClientTimeout: 5 * time.Second,
	}
	for _, opt := range options {
		opt(opts)
	}
	// start all the listeners
	for name, listener := range Listeners {
		glog.Infof("Starting Listener: %s on %s", name, listener.Uri())
		go listener.Listen(ctx)
	}

	// start all the outputs
	for name, output := range Outputs {
		glog.Infof("Starting output: %s", name)
		go output.Start(ctx, opts)
	}

	return nil
}

func GetApiPluginsList() ApiPlugins {

	choices := ApiPlugins{
		Parsers:    make([]string, 0),
		Processors: make([]string, 0, len(Processors)),
		Outputs:    make([]string, 0, len(Outputs)),
		Listeners:  make([]string, 0, len(Listeners)),
	}

	for _, k := range Processors {
		choices.Processors = append(choices.Processors, k.Name())
	}

	for k := range Outputs {
		choices.Outputs = append(choices.Outputs, k)
	}

	for k := range Listeners {
		choices.Listeners = append(choices.Listeners, k)
		if k == "webhook" {
			choices.Parsers = Listeners[k].GetParsersList()
		}
	}

	return choices
}

type Options struct {
	WebUrl        string
	ClientTimeout time.Duration
}

type PluginOption func(*Options)

func WebUrl(url string) PluginOption {
	return func(o *Options) {
		o.WebUrl = url
	}
}

func ClientTimeout(to time.Duration) PluginOption {
	return func(o *Options) {
		o.ClientTimeout = to
	}
}
