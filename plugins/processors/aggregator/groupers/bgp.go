package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/types"
)

type bgpGrouper struct {
	name string
}

// grouperFunc defines the condition for two bgp peers to be considered same to be grouped together
func (g bgpGrouper) grouperFunc() groupingFunc {
	return func(i, j interface{}) bool {
		return (i.(types.BgpPeer).LocalDevice.Name == j.(types.BgpPeer).RemoteDevice.Name && i.(types.BgpPeer).RemoteDevice.Name == j.(types.BgpPeer).LocalDevice.Name)
	}
}

func (g *bgpGrouper) Name() string {
	return g.name
}

func (g *bgpGrouper) origAlerts(alerts []*models.Alert, group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range alerts {
			if a.Id == p.(types.BgpPeer).AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *bgpGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	// first group by peer endpoints. Assume the alert metadata contains the peer-device
	var peers []interface{}
	var groupedAlerts [][]*models.Alert
	for _, alert := range alerts {
		if !alert.Metadata.Valid || alert.Status != models.Status_ACTIVE {
			continue
		}
		if !alert.HasTags("bgp") {
			glog.V(2).Infof("Bgp Agg: Found non bgp alert, skip grouping")
			return groupedAlerts
		}
		p := types.BgpPeer{}
		if err := alert.LoadMeta(&p); err != nil {
			glog.Errorf("Bgp Agg: Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		peers = append(peers, p)
	}
	if len(peers) == 0 {
		return groupedAlerts
	}
	glog.V(4).Infof("Bgp Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), peers)
	//TODO : group by device

	for _, group := range groups {
		orig := g.origAlerts(alerts, group)
		groupedAlerts = append(groupedAlerts, orig)
	}
	return groupedAlerts
}

func init() {
	g := &bgpGrouper{name: "bgp_session"}
	AddGrouper(g)
}
