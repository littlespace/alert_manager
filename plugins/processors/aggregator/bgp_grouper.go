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

type bgpGrouper struct {
	sub      []string
	recvChan chan *models.Alert
	recvBuf  []*models.Alert

	sync.Mutex
}

// grouperFunc defines the condition for two bgp peers to be considered same to be grouped together
func (g bgpGrouper) grouperFunc() func(i, j interface{}) bool {
	return func(i, j interface{}) bool {
		return (i.(BgpPeer).LocalDevice == j.(BgpPeer).RemoteDevice && i.(BgpPeer).RemoteDevice == j.(BgpPeer).LocalDevice) ||
			(i.(BgpPeer).LocalDevice == j.(BgpPeer).LocalDevice && i.(BgpPeer).RemoteDevice == j.(BgpPeer).RemoteDevice)
	}
}

func (g *bgpGrouper) name() string {
	return "bgp_session"
}

func (g *bgpGrouper) addSubscription(a string) {
	g.sub = append(g.sub, a)
}

func (g *bgpGrouper) subscribed() []string {
	return g.sub
}

func (g *bgpGrouper) addToBuf(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvBuf = append(g.recvBuf, a)
}

func (g *bgpGrouper) addAlert(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvChan <- a
}

func (g *bgpGrouper) origAlerts(group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range g.recvBuf {
			if a.Id == p.(BgpPeer).AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *bgpGrouper) doGrouping() {
	g.Lock()
	defer g.Unlock()
	// first group by peer endpoints. Assume the alert metadata contains the peer-device
	var peers []interface{}
	for _, alert := range g.recvBuf {
		if !alert.Metadata.Valid {
			continue
		}
		if !alert.HasTags("bgp") {
			glog.V(2).Infof("Bgp Agg: Found non bgp alert, skip grouping")
			g.recvBuf = g.recvBuf[:0]
			return
		}
		p := BgpPeer{}
		if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
			glog.Errorf("Bgp Agg: Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		peers = append(peers, p)
	}
	glog.V(4).Infof("Bgp Agg: Now grouping %d alerts", len(g.recvBuf))
	groups := group(g.grouperFunc(), peers)
	//TODO : group by device

	// create new aggregated alerts
	for _, group := range groups {
		orig := g.origAlerts(group)
		groupedChan <- &alertGroup{groupedAlerts: orig, grouper: g}
	}
	g.recvBuf = g.recvBuf[:0]
}

func (g *bgpGrouper) start(ctx context.Context) {
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
	g := &bgpGrouper{recvChan: make(chan *models.Alert)}
	addGrouper(g)
}
