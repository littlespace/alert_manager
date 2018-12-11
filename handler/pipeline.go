package handler

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sort"
)

// Processor is an alert processor thats part of a processor pipeline
type Processor interface {
	Name() string
	Stage() int
	Process(ctx context.Context, db models.Dbase, in chan *AlertEvent) chan *AlertEvent
}

var Processors []Processor

func GetProcessor(name string) Processor {
	for _, p := range Processors {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

func AddProcessor(p Processor) {
	Processors = append(Processors, p)
}

// Pipeline is a pipeline of alert processors
type Pipeline interface {
	Next() Processor
	Run(ctx context.Context, db models.Dbase, in chan *AlertEvent)
}

type ProcessorPipeline struct {
	processors <-chan Processor
}

func NewProcessorPipeline() Pipeline {
	sort.Slice(Processors, func(i, j int) bool { return Processors[i].Stage() < Processors[j].Stage() })
	pChan := make(chan Processor, len(Processors)+1)
	for _, p := range Processors {
		pChan <- p
	}
	pChan <- nil
	return &ProcessorPipeline{processors: pChan}
}

func (p ProcessorPipeline) Next() Processor {
	return <-p.processors
}

// Run starts the processor pipeline
func (p ProcessorPipeline) Run(ctx context.Context, db models.Dbase, in chan *AlertEvent) {
	processor := p.Next()
	if processor == nil {
		go func() {
			for {
				select {
				case <-in:
				case <-ctx.Done():
					return
				}
			}
		}()
		return
	}
	out := processor.Process(ctx, db, in)
	p.Run(ctx, db, out)
}
