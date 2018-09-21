package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type bgpGrouper struct {
	name string
}

// grouperFunc defines the condition for two bgp label sets to be considered same to be grouped together
func (g bgpGrouper) grouperFunc() groupingFunc {
	return func(i, j models.Labels) bool {
		return (
		// two ends of the same session
		i["LocalDeviceName"].(string) == j["RemoteDeviceName"].(string) && i["RemoteDeviceName"].(string) == j["LocalDeviceName"].(string) ||
			// two sessions from/to same device
			i["LocalDeviceName"].(string) == j["LocalDeviceName"].(string) && i["RemoteDeviceName"].(string) == j["RemoteDeviceName"].(string))
	}
}

func (g *bgpGrouper) Name() string {
	return g.name
}

func (g *bgpGrouper) origAlerts(alerts []*models.Alert, group []models.Labels) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range alerts {
			if a.Id == p["AlertId"].(int64) {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *bgpGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	// first group by peer endpoints. Assume the alert metadata contains the peer-device
	var labels []models.Labels
	var groupedAlerts [][]*models.Alert
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		if alert.Labels["LabelType"].(string) != "Bgp" {
			glog.V(2).Infof("Bgp Agg: Found non bgp alert, skip grouping")
			return groupedAlerts
		}
		alert.Labels["AlertId"] = alert.Id
		labels = append(labels, alert.Labels)
	}
	if len(labels) == 0 {
		return groupedAlerts
	}
	glog.V(4).Infof("Bgp Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), labels)

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
