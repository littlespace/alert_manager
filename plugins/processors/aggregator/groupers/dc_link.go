package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type dcCktGrouper struct {
	name string
}

// grouperFunc defines the condition for two circuit related objects  to be considered same to be grouped together
// Currently, it considers a/z endpoints as well as bgp peers
func (g dcCktGrouper) grouperFunc() groupingFunc {
	return func(i, j models.Labels) bool {
		switch i["LabelType"].(string) {
		case "Circuit":
			switch j["LabelType"].(string) {
			case "Circuit":
				return i["ASideDeviceName"] == j["ZSideDeviceName"] && j["ASideDeviceName"] == i["ZSideDeviceName"]

			case "Bgp":
				m := (j["LocalInterface"] == i["ASideInterface"] && j["LocalDeviceName"] == i["ASideDeviceName"]) && (j["RemoteInterface"] == i["ZSideInterface"] && j["RemoteDeviceName"] == i["ZSideDeviceName"])
				n := (j["LocalInterface"] == i["ZSideInterface"] && j["LocalDeviceName"] == i["ZSideDeviceName"]) && (j["RemoteInterface"] == i["ASideInterface"] && j["RemoteDeviceName"] == i["ASideDeviceName"])
				return m || n
			}
		case "Bgp":
			switch j["LabelType"].(string) {
			case "Circuit":
				m := (i["LocalInterface"] == j["ASideInterface"] && i["LocalDeviceName"] == j["ASideDeviceName"]) && (i["RemoteInterface"] == j["ZSideInterface"] && i["RemoteDeviceName"] == j["ZSideDeviceName"])
				n := (i["LocalInterface"] == j["ZSideInterface"] && i["LocalDeviceName"] == j["ZSideDeviceName"]) && (i["RemoteInterface"] == j["ASideInterface"] && i["RemoteDeviceName"] == j["ASideDeviceName"])
				return m || n
			case "Bgp":
				return (i["LocalDeviceName"] == j["RemoteDeviceName"] && i["RemoteDeviceName"] == j["LocalDeviceName"]) || (i["LocalDeviceName"] == j["LocalDeviceName"] && i["RemoteDeviceName"] == j["RemoteDeviceName"])
			}
		}
		return false
	}
}

func (g *dcCktGrouper) Name() string {
	return g.name
}

func (g *dcCktGrouper) origAlerts(alerts []*models.Alert, group []models.Labels) []*models.Alert {
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

func (g *dcCktGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	var labels []models.Labels
	var groupedAlerts [][]*models.Alert
	allBgp := true
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		allBgp = allBgp && alert.Labels["LabelType"].(string) == "Bgp"
		alert.Labels["AlertId"] = alert.Id
		labels = append(labels, alert.Labels)
	}
	if allBgp {
		glog.V(2).Infof("Ckt Agg: Did not find a dc link alert, skip grouping")
		return groupedAlerts
	}
	if len(labels) == 0 {
		return groupedAlerts
	}
	glog.V(4).Infof("Ckt Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), labels)
	//TODO : group by device

	for _, group := range groups {
		orig := g.origAlerts(alerts, group)
		groupedAlerts = append(groupedAlerts, orig)
	}
	return groupedAlerts
}

func init() {
	g := &dcCktGrouper{name: "dc_circuit_down"}
	AddGrouper(g)
}
