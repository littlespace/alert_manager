package handler

var (
	// Registered transforms
	Transforms []Transform
	// map of alert names: registered processors
	Processors = make(map[string][]chan *AlertEvent)
	// map of output names: registered outputs
	Outputs = make(map[string]chan *AlertEvent)
)

func AddTransform(t Transform) {
	Transforms = append(Transforms, t)
}

// RegisterProcessor registers a new processor with the handler
func RegisterProcessor(alertName string, recvChan chan *AlertEvent) {
	Processors[alertName] = append(Processors[alertName], recvChan)
}

func RegisterOutput(outName string, outputChan chan *AlertEvent) {
	Outputs[outName] = outputChan
}
