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

const ruleName = "bgp_session"

// grouperFunc defines the condition for two bgp peers to be considered same to be grouped together
var grouperFunc = func(i, j interface{}) bool {
	return (i.(BgpPeer).LocalDevice == j.(BgpPeer).RemoteDevice && i.(BgpPeer).RemoteDevice == j.(BgpPeer).LocalDevice) ||
		(i.(BgpPeer).LocalDevice == j.(BgpPeer).LocalDevice && i.(BgpPeer).RemoteDevice == j.(BgpPeer).RemoteDevice)
}

type bgpGrouper struct {
	ruleConfig ah.AggregationRuleConfig
	recvChan   chan *models.Alert
	recvBuf    []*models.Alert
	db         *models.DB

	sync.Mutex
}

func (g *bgpGrouper) setRule(rule ah.AggregationRuleConfig) {
	g.ruleConfig = rule
}

func (g *bgpGrouper) getRule() ah.AggregationRuleConfig {
	return g.ruleConfig
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

func (g *bgpGrouper) origAlerts(group []BgpPeer) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range g.recvBuf {
			if a.Id == p.AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *bgpGrouper) doGrouping(ctx context.Context) {
	// first group by peer endpoints. Assume the alert metadata contains the peer-device
	g.Lock()
	defer g.Unlock()
	glog.V(4).Infof("Bgp Agg: Now grouping %d alerts", len(g.recvBuf))
	var peers []interface{}
	for _, alert := range g.recvBuf {
		p := BgpPeer{}
		if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
			glog.Errorf("Bgp Agg: Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		peers = append(peers, p)
	}

	groups := group(grouperFunc, peers)
	//TODO : group by device

	// create new aggregated alerts
	tx := models.NewTx(g.db)
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
		for _, group := range groups {
			var b []BgpPeer
			for _, gg := range group {
				b = append(b, gg.(BgpPeer))
			}
			err := aggAlert(ctx, tx, g, g.origAlerts(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		glog.V(2).Infof("Bgp Agg: Unable to create Agg alert: %v", err)
	}
	g.recvBuf = g.recvBuf[:0]
}

func (g *bgpGrouper) start(ctx context.Context, db *models.DB) {
	g.db = db
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
	g := &bgpGrouper{recvChan: make(chan *models.Alert)}
	addGrouper(ruleName, g)
}
