package stats

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounter(t *testing.T) {
	c := NewCounter("test")
	c.Add(10)
	assert.Equal(t, int(c.value), 10)
	c.Set(5) // should do nothing
	assert.Equal(t, int(c.value), 10)
	c.Add(-1) // should do nothing
	assert.Equal(t, int(c.value), 10)
	dp := c.toDatapoint()
	assert.Nil(t, dp.Tags)
	assert.Equal(t, dp.Fields, map[string]interface{}{"test": int64(10)})
	c.Reset()
	assert.Equal(t, int(c.value), 0)
}

func isPresent(g *Gauge, value int64) bool {
	for _, v := range g.values {
		if v == value {
			return true
		}
	}
	return false
}

func TestGauge(t *testing.T) {
	g := NewGauge("test")
	g.Add(10) // should do nothing
	assert.Equal(t, isPresent(g, 10), false)
	g.Set(10)
	g.Set(20)
	g.Set(-5)
	assert.Equal(t, len(g.values), 3)
	assert.Equal(t, isPresent(g, 10), true)
	assert.Equal(t, isPresent(g, -5), true)
	dps := g.toDatapoint()
	assert.Equal(t, len(dps), 3)
	g.Reset()
	assert.Equal(t, len(g.values), 0)
}
