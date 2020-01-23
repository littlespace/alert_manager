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

	sync.Mutex
}

func (g *Grouper) startWindow(grouper groupers.Grouper, ruleName string) {
	rule, _ := ah.Config.GetAggregationRuleConfig(ruleName)
	time.Sleep(rule.Window)
	g.Lock()
	defer g.Unlock()
	for _, group := range groupers.DoGrouping(grouper, g.recvBuffers[ruleName]) {
		groupedChan <- &alertGroup{groupedAlerts: group, grouper: grouper, ruleName: ruleName}
	}
	g.recvBuffers[ruleName] = g.recvBuffers[ruleName][:0]
}

func (g *Grouper) addAlert(grouper groupers.Grouper, ruleName string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	if len(g.recvBuffers[ruleName]) == 0 {
		go g.startWindow(grouper, ruleName)
	}
	for _, a := range g.recvBuffers[ruleName] {
		if a.Id == alert.Id {
			return
		}
	}
	g.recvBuffers[ruleName] = append(g.recvBuffers[ruleName], alert)
}

func (g *Grouper) removeAlert(ruleName string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	for i := 0; i < len(g.recvBuffers[ruleName]); i++ {
		if g.recvBuffers[ruleName][i].Id == alert.Id {
			g.recvBuffers[ruleName] = append(g.recvBuffers[ruleName][:i], g.recvBuffers[ruleName][i+1:]...)
			i--
		}
	}
}
