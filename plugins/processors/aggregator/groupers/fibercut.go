package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type fibercutGrouper struct {
	name string
}

// grouperFunc defines the condition for circuit endpoints to be considered same to be grouped together
func (g fibercutGrouper) grouperFunc() groupingFunc {
	return func(i, j models.Labels) bool {
		var match bool
		if i["Provider"] != "" && j["Provider"] != "" {
			match = match || i["Provider"] == j["Provider"]
		}
		return (match ||
			// 2 ends of same circuit
			(i["ASideDeviceName"] == j["ZSideDeviceName"] && i["ASideInterface"] == j["ZSideInterface"]) ||
			// phy member of lag
			(i["ASideDeviceName"] == j["ASideDeviceName"] && (i["ASideInterface"] == j["ASideAgg"] || i["ASideAgg"] == j["ASideInterface"])))

	}
}

func (g *fibercutGrouper) Name() string {
	return g.name
}

func (g *fibercutGrouper) origAlerts(alerts []*models.Alert, group []models.Labels) []*models.Alert {
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

func (g *fibercutGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	var groupedAlerts [][]*models.Alert
	var labels []models.Labels
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		alert.Labels["AlertId"] = alert.Id
		labels = append(labels, alert.Labels)
	}
	if len(labels) == 0 {
		return groupedAlerts
	}
	glog.V(4).Infof("Fibercut Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), labels)

	for _, group := range groups {
		orig := g.origAlerts(alerts, group)
		groupedAlerts = append(groupedAlerts, orig)
	}
	return groupedAlerts
}

func init() {
	g := &fibercutGrouper{name: "fibercut"}
	AddGrouper(g)
}
