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
	// map of output names: registered outputs
	Outputs = make(map[string]chan *models.AlertEvent)

	gMu sync.Mutex
)

func AddTransform(t Transform) {
	Transforms = append(Transforms, t)
}

func RegisterOutput(outName string, outputChan chan *models.AlertEvent) {
	gMu.Lock()
	defer gMu.Unlock()
	Outputs[outName] = outputChan
}

func GetOutput(output string) (chan *models.AlertEvent, bool) {
	gMu.Lock()
	defer gMu.Unlock()
	o, ok := Outputs[output]
	return o, ok
}
