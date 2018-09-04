package aggregator

import (
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins/processors/aggregator/groupers"
	"sync"
	"time"
)

// Grouper manages the alert buffers for the different groupers and their grouping for time-window based grouping methods.
type Grouper struct {
	recvBuffers map[string][]*models.Alert
	subs        map[string][]string

	sync.Mutex
}

func (g *Grouper) startWindow(name string) {
	rule, _ := ah.Config.GetAggregationRuleConfig(name)
	<-time.After(rule.Window)
	g.Lock()
	defer g.Unlock()
	grouper := groupers.AllGroupers[name]
	for _, group := range grouper.DoGrouping(g.recvBuffers[name]) {
		groupedChan <- &alertGroup{groupedAlerts: group, grouper: grouper}
	}
	g.recvBuffers[name] = g.recvBuffers[name][:0]
}

func (g *Grouper) addAlert(name string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	if len(g.recvBuffers[name]) == 0 {
		go g.startWindow(name)
	}
	g.recvBuffers[name] = append(g.recvBuffers[name], alert)
}

func (g *Grouper) removeAlert(name string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	for i := 0; i < len(g.recvBuffers[name]); i++ {
		if g.recvBuffers[name][i].Id == alert.Id {
			g.recvBuffers[name] = append(g.recvBuffers[name][:i], g.recvBuffers[name][i+1:]...)
			i--
		}
	}
}

func (g *Grouper) addSubscription(name string, alertName string) {
	g.Lock()
	defer g.Unlock()
	g.subs[name] = append(g.subs[name], alertName)
}

func (g *Grouper) subscribed(name string) []string {
	g.Lock()
	defer g.Unlock()
	return g.subs[name]
}
