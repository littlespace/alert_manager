package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"sync"
	"time"
)

const ruleName = "bgp_session"

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

func (g *bgpGrouper) addToBuf(a *models.Alert) {
	g.Lock()
	defer g.Unlock()
	g.recvBuf = append(g.recvBuf, a)
}

func (g *bgpGrouper) addAlert(a *models.Alert) {
	g.recvChan <- a
}

func (g *bgpGrouper) aggAlert(ctx context.Context, tx *models.Tx, group []BgpPeer) (*models.Alert, error) {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range g.recvBuf {
			if a.Id == p.AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	peer := group[0] // alert only on the first peer in the group
	desc := fmt.Sprintf("BGP session down: %s (%s) <-> %s (%s)",
		peer.LocalDevice, peer.LocalIp, peer.RemoteDevice, peer.RemoteIp)
	agg := &models.Alert{
		Name:        g.ruleConfig.Alert.Name,
		Description: desc,
		Source:      g.ruleConfig.Alert.Config.Source,
		Severity:    models.SevMap[g.ruleConfig.Alert.Config.Severity],
		StartTime:   orig[0].StartTime,
		LastActive:  orig[0].LastActive,
		ExternalId:  orig[0].ExternalId,
	}
	agg.AddDevice(peer.LocalDevice)
	agg.AddTags(g.ruleConfig.Alert.Config.Tags...)
	if g.ruleConfig.Alert.Config.AutoExpire != nil && *g.ruleConfig.Alert.Config.AutoExpire {
		agg.SetAutoExpire(g.ruleConfig.Alert.Config.ExpireAfter)
	}
	if g.ruleConfig.Alert.Config.AutoClear != nil {
		agg.AutoClear = *g.ruleConfig.Alert.Config.AutoClear
	}
	var newId int64
	stmt, err := tx.PrepareNamed(models.QueryInsertNew)
	err = stmt.Get(&newId, agg)
	if err != nil {
		return nil, fmt.Errorf("Unable to insert new alert: %v", err)
	}
	agg.Id = newId
	// update the agg IDs of all the original alerts
	var origIds []int64
	for _, o := range orig {
		origIds = append(origIds, o.Id)
	}
	err = tx.InQuery(models.QueryUpdateAggId, agg.Id, origIds)
	if err != nil {
		return nil, fmt.Errorf("Unable to update agg Ids: %v", err)
	}
	return agg, nil
}

func (g *bgpGrouper) doGrouping(ctx context.Context) {
	// first group by peer endpoints. Assume the alert metadata contains the peer-device
	g.Lock()
	defer g.Unlock()
	var peers []BgpPeer
	for _, alert := range g.recvBuf {
		p := BgpPeer{}
		if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
			glog.Errorf("Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		peers = append(peers, p)
	}
	groups := groupBySession(peers)
	//TODO : group by device

	// create new aggregated alerts
	tx := models.NewTx(g.db)
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx *models.Tx) error {
		for _, group := range groups {
			aggAlert, err := g.aggAlert(ctx, tx, group)
			if err != nil {
				return err
			}
			// send the aggAlert to the right output
			ah.NotifyOutputs(
				&ah.AlertEvent{Alert: aggAlert, Type: ah.EventType_ACTIVE},
				g.ruleConfig.Alert.Config.Outputs,
			)
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
	g.recvChan = make(chan *models.Alert)
	var start time.Time
	for {
		alert := <-g.recvChan
		if len(g.recvBuf) == 0 {
			start = time.Now()
		}
		g.addToBuf(alert)
		if time.Now().Sub(start) >= g.ruleConfig.Window {
			glog.V(4).Infof("Bgp Agg: Now grouping %d alerts", len(g.recvBuf))
			g.doGrouping(ctx)
			start = time.Time{}
		}
	}
}

func init() {
	g := &bgpGrouper{recvChan: make(chan *models.Alert)}
	addGrouper(ruleName, g)
}
