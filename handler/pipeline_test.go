package handler

import (
	"context"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

type stage1 struct{}

func (s *stage1) Name() string { return "Stage1" }

func (s *stage1) Stage() int { return 1 }

func (s *stage1) Process(ctx context.Context, db models.Dbase, in chan *AlertEvent) chan *AlertEvent {
	out := make(chan *AlertEvent, 2)
	event1 := <-in
	event1.Type = EventType_SUPPRESSED
	out <- event1
	event2 := <-in
	out <- event2
	close(out)
	return out
}

type stage2 struct{}

func (s *stage2) Name() string { return "Stage2" }

func (s *stage2) Stage() int { return 2 }

func (s *stage2) Process(ctx context.Context, db models.Dbase, in chan *AlertEvent) chan *AlertEvent {
	out := make(chan *AlertEvent, 1)
	for event := range in {
		if event.Type == EventType_SUPPRESSED {
			continue
		}
		out <- event
	}
	close(out)
	return out
}

type collector struct {
	recvd []*AlertEvent
}

func (c *collector) Name() string { return "Collector" }

func (c *collector) Stage() int { return 3 }

func (c *collector) Process(ctx context.Context, db models.Dbase, in chan *AlertEvent) chan *AlertEvent {
	out := make(chan *AlertEvent, 1)
	for event := range in {
		c.recvd = append(c.recvd, event)
	}
	close(out)
	return out
}

func TestPipelineStages(t *testing.T) {
	Processors = []Processor{}
	AddProcessor(&stage2{})
	AddProcessor(&stage1{})
	c := &collector{}
	AddProcessor(c)

	event1 := &AlertEvent{Type: EventType_ACTIVE}
	event2 := &AlertEvent{Type: EventType_ACTIVE}

	in := make(chan *AlertEvent, 2)
	in <- event1
	in <- event2
	close(in)

	// test a 3 stage pipeline
	p := NewProcessorPipeline()
	p.Run(context.Background(), &MockDb{}, in)

	assert.Equal(t, event1.Type, EventType_SUPPRESSED)
	assert.Equal(t, len(c.recvd), 1)
	assert.Equal(t, c.recvd[0], event2)
}
