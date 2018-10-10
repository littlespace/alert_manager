package handler

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	"sync"
)

type Transform interface {
	Name() string
	GetPriority() int
	GetRegister() string
	Apply(alert *models.Alert) error
}

var (
	// Registered transforms
	Transforms []Transform
	// map of alert names: registered processors
	Processors = make(map[string][]chan *AlertEvent)
	// map of output names: registered outputs
	Outputs = make(map[string]chan *AlertEvent)

	gMu sync.Mutex
)

func AddTransform(t Transform) {
	Transforms = append(Transforms, t)
}

// RegisterProcessor registers a new processor with the handler
func RegisterProcessor(alertName string, recvChan chan *AlertEvent) {
	gMu.Lock()
	defer gMu.Unlock()
	Processors[alertName] = append(Processors[alertName], recvChan)
}

func RegisterOutput(outName string, outputChan chan *AlertEvent) {
	Outputs[outName] = outputChan
}
