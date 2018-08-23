package stats

import (
	"context"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/reporting"
	"sync"
	"time"
)

type statType int

const measurement = "alert_manager_stats"

// counter is always incremented and reset to 0 at the end of the export interval
type Counter struct {
	name  string
	value int64

	sync.Mutex
}

// gauge is any arbitrary value and does not reset
type Gauge struct {
	name   string
	values map[time.Time]int64

	sync.Mutex
}

var (
	allCounters []*Counter
	allGauges   []*Gauge
)

func NewCounter(name string) *Counter {
	s := &Counter{name: name}
	allCounters = append(allCounters, s)
	return s
}

func NewGauge(name string) *Gauge {
	s := &Gauge{name: name, values: make(map[time.Time]int64)}
	allGauges = append(allGauges, s)
	return s
}

func (c *Counter) Add(value int64) {
	c.Lock()
	defer c.Unlock()
	if value < 0 {
		glog.Errorf("Cannot subtract from counter type")
		return
	}
	c.value += value
}

func (c *Counter) Reset() {
	c.Lock()
	defer c.Unlock()
	c.value = 0
}

func (c *Counter) toDatapoint() *reporting.Datapoint {
	return &reporting.Datapoint{
		Measurement: measurement,
		Fields:      map[string]interface{}{c.name: c.value},
		TimeStamp:   time.Now(),
	}
}

func (g *Gauge) Set(value int64) {
	g.Lock()
	defer g.Unlock()
	g.values[time.Now()] = value
}

func (g *Gauge) toDatapoint() []*reporting.Datapoint {
	dp := []*reporting.Datapoint{}
	for ts, v := range g.values {
		dp = append(dp, &reporting.Datapoint{
			Measurement: measurement,
			Fields:      map[string]interface{}{g.name: v},
			TimeStamp:   ts,
		})
	}
	return dp
}

func StartExport(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 60 * time.Second
	}
	t := time.NewTicker(interval)
	for {
		select {
		case <-t.C:
			for _, c := range allCounters {
				reporting.DataChan <- c.toDatapoint()
				c.Reset()
			}
			for _, g := range allGauges {
				for _, dp := range g.toDatapoint() {
					reporting.DataChan <- dp
				}
				g.values = make(map[time.Time]int64)
			}
		case <-ctx.Done():
			return
		}
	}
}
