package handler

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sort"
)

type Processor interface {
	Name() string
	Stage() int
	SetDb(db models.Dbase)
	Process(ctx context.Context, in, out chan *AlertEvent, done chan struct{})
}

var Processors []Processor

func AddProcessor(p Processor) {
	Processors = append(Processors, p)
}

type Pipeline interface {
	Next() Processor
	Run(ctx context.Context, in chan *AlertEvent)
}

type ProcessorPipeline struct {
	processors <-chan Processor
}

func NewProcessorPipeline(db models.Dbase) Pipeline {
	sort.Slice(Processors, func(i, j int) bool { return Processors[i].Stage() < Processors[j].Stage() })
	pChan := make(chan Processor, len(Processors)+1)
	for _, p := range Processors {
		p.SetDb(db)
		pChan <- p
	}
	pChan <- nil
	return &ProcessorPipeline{processors: pChan}
}

func (p ProcessorPipeline) Next() Processor {
	return <-p.processors
}

func (p ProcessorPipeline) Run(ctx context.Context, in chan *AlertEvent) {
	processor := p.Next()
	if processor == nil {
		return
	}
	out := make(chan *AlertEvent)
	done := make(chan struct{})
	go processor.Process(ctx, in, out, done)
	<-done
	p.Run(ctx, out)
}
