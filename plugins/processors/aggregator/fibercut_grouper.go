package aggregator

import (
	"context"
	"encoding/json"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sync"
	"time"
)

type fibercutGrouper struct {
	sub      []string
	recvChan chan *models.Alert
	recvBuf  []*models.Alert

	sync.Mutex
}

func (g fibercutGrouper) grouperFunc() func(i, j interface{}) bool {
	return func(i, j interface{}) bool {
		return i.(Circuit).Provider == j.(Circuit).Provider || i.(Circuit).ASide == j.(Circuit).ZSide
	}
}

func (g *fibercutGrouper) name() string {
	return "fibercut"
}

func (g *fibercutGrouper) addSubscription(a string) {
	g.sub = append(g.sub, a)
}

func (g *fibercutGrouper) subscribed() []string {
	return g.sub
}

func (g *fibercutGrouper) addToBuf(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvBuf = append(g.recvBuf, a)
}

func (g *fibercutGrouper) addAlert(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvChan <- a
}

func (g *fibercutGrouper) origAlerts(group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range g.recvBuf {
			if a.Id == p.(Circuit).AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *fibercutGrouper) doGrouping() {
	g.Lock()
	defer g.Unlock()
	var ckts []interface{}
	for _, alert := range g.recvBuf {
		if !alert.Metadata.Valid {
			continue
		}
		p := Circuit{}
		if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
			glog.Errorf("Fibercut Agg: Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		ckts = append(ckts, p)
	}
	glog.V(4).Infof("Fibercut Agg: Now grouping %d alerts", len(g.recvBuf))
	groups := group(g.grouperFunc(), ckts)

	// create new aggregated alerts
	for _, group := range groups {
		orig := g.origAlerts(group)
		groupedChan <- &alertGroup{groupedAlerts: orig, grouper: g}
	}
	g.recvBuf = g.recvBuf[:0]
}

func (g *fibercutGrouper) start(ctx context.Context) {
	rule, _ := ah.Config.GetAggregationRuleConfig(g.name())
	for {
		select {
		case alert := <-g.recvChan:
			if len(g.recvBuf) == 0 {
				go func() {
					<-time.After(rule.Window)
					g.doGrouping()
				}()
			}
			g.addToBuf(alert)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	g := &fibercutGrouper{recvChan: make(chan *models.Alert)}
	addGrouper(g)
}
