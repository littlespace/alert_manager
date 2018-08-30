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

type dcCktGrouper struct {
	ruleConfig ah.AggregationRuleConfig
	recvChan   chan *models.Alert
	recvBuf    []*models.Alert

	sync.Mutex
}

func (g dcCktGrouper) grouperFunc() func(i, j interface{}) bool {
	return func(i, j interface{}) bool {
		switch i := i.(type) {
		case Circuit:
			switch j := j.(type) {
			case Circuit:
				return i.ASide == j.ZSide && j.ASide == i.ZSide
			case BgpPeer:

				m := (j.LocalInterface == i.ASide.Interface && j.LocalDevice == i.ASide.Device) && (j.RemoteInterface == i.ZSide.Interface && j.RemoteDevice == i.ZSide.Device)
				n := (j.LocalInterface == i.ZSide.Interface && j.LocalDevice == i.ZSide.Device) && (j.RemoteInterface == i.ASide.Interface && j.RemoteDevice == i.ASide.Device)
				return m || n
			}
		case BgpPeer:
			switch j := j.(type) {
			case Circuit:
				m := (i.LocalInterface == j.ASide.Interface && i.LocalDevice == j.ASide.Device) && (i.RemoteInterface == j.ZSide.Interface && i.RemoteDevice == j.ZSide.Device)
				n := (i.LocalInterface == j.ZSide.Interface && i.LocalDevice == j.ZSide.Device) && (i.RemoteInterface == j.ASide.Interface && i.RemoteDevice == j.ASide.Device)
				return m || n
			case BgpPeer:
				return (i.LocalDevice == j.RemoteDevice && i.RemoteDevice == j.LocalDevice) || (i.LocalDevice == j.LocalDevice && i.RemoteDevice == j.RemoteDevice)
			}
		}
		return false
	}
}

func (g *dcCktGrouper) name() string {
	return "dc_circuit_down"
}

func (g *dcCktGrouper) setRule(rule ah.AggregationRuleConfig) {
	g.ruleConfig = rule
}

func (g *dcCktGrouper) getRule() ah.AggregationRuleConfig {
	return g.ruleConfig
}

func (g *dcCktGrouper) addToBuf(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvBuf = append(g.recvBuf, a)
}

func (g *dcCktGrouper) addAlert(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvChan <- a
}

func (g *dcCktGrouper) origAlerts(group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
	innerfor:
		for _, a := range g.recvBuf {
			var cond bool
			if c, ok := p.(Circuit); ok {
				cond = a.Id == c.AlertId
			} else if b, ok := p.(BgpPeer); ok {
				cond = a.Id == b.AlertId
			}
			if cond {
				orig = append(orig, a)
				break innerfor
			}
		}
	}
	return orig
}

func (g *dcCktGrouper) doGrouping(ctx context.Context) {
	g.Lock()
	defer g.Unlock()
	var entities []interface{}
	allBgp := true
	for _, alert := range g.recvBuf {
		allBgp = allBgp && alert.HasTags("bgp")
		if alert.HasTags("bgp") {
			p := BgpPeer{}
			if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
				glog.Errorf("Ckt Agg: Unable to unmarshal metadata: %v", err)
				continue
			}
			p.AlertId = alert.Id
			entities = append(entities, p)
		} else {
			c := Circuit{}
			if err := json.Unmarshal([]byte(alert.Metadata.String), &c); err != nil {
				glog.Errorf("Ckt Agg: Unable to unmarshal metadata: %v", err)
				continue
			}
			c.AlertId = alert.Id
			entities = append(entities, c)
		}
	}
	if allBgp {
		glog.V(2).Infof("Ckt Agg: Did not find a dc link alert, skip grouping")
		g.recvBuf = g.recvBuf[:0]
		return
	}
	glog.V(4).Infof("Ckt Agg: Now grouping %d alerts", len(g.recvBuf))
	groups := group(g.grouperFunc(), entities)
	//TODO : group by device

	// create new aggregated alerts
	for _, group := range groups {
		orig := g.origAlerts(group)
		groupedChan <- &alertGroup{groupedAlerts: orig, grouper: g}
	}
	g.recvBuf = g.recvBuf[:0]
}

func (g *dcCktGrouper) start(ctx context.Context) {
	for {
		alert := <-g.recvChan
		if len(g.recvBuf) == 0 {
			go func() {
				<-time.After(g.ruleConfig.Window)
				g.doGrouping(ctx)
			}()
		}
		g.addToBuf(alert)
	}
}

func init() {
	g := &dcCktGrouper{recvChan: make(chan *models.Alert)}
	addGrouper(g)
}
