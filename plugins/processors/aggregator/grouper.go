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
	name := grouper.Name()
	rule, _ := ah.Config.GetAggregationRuleConfig(ruleName)
	time.Sleep(rule.Window)
	g.Lock()
	defer g.Unlock()
	for _, group := range groupers.DoGrouping(grouper, g.recvBuffers[name]) {
		groupedChan <- &alertGroup{groupedAlerts: group, grouper: grouper, ruleName: ruleName}
	}
	g.recvBuffers[name] = g.recvBuffers[name][:0]
}

func (g *Grouper) addAlert(grouper groupers.Grouper, ruleName string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	name := grouper.Name()
	if len(g.recvBuffers[name]) == 0 {
		go g.startWindow(grouper, ruleName)
	}
	for _, a := range g.recvBuffers[name] {
		if a.Id == alert.Id {
			return
		}
	}
	g.recvBuffers[name] = append(g.recvBuffers[name], alert)
}

func (g *Grouper) removeAlert(grouperName string, alert *models.Alert) {
	g.Lock()
	defer g.Unlock()
	for i := 0; i < len(g.recvBuffers[grouperName]); i++ {
		if g.recvBuffers[grouperName][i].Id == alert.Id {
			g.recvBuffers[grouperName] = append(g.recvBuffers[grouperName][:i], g.recvBuffers[grouperName][i+1:]...)
			i--
		}
	}
}
