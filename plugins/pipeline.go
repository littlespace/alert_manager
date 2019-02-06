package plugins

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sort"
)

// Pipeline is a pipeline of alert processors
type Pipeline interface {
	Next() Processor
	Run(ctx context.Context, db models.Dbase, in chan *models.AlertEvent)
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
func (p ProcessorPipeline) Run(ctx context.Context, db models.Dbase, in chan *models.AlertEvent) {
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
