package groupers

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type dcCktGrouper struct {
	name string
}

// grouperFunc defines the condition for two circuit related objects  to be considered same to be grouped together
// Currently, it considers a/z endpoints as well as bgp peers
func (g dcCktGrouper) GrouperFunc() GroupingFunc {
	return func(i, j *models.Alert) bool {
		switch i.Labels["LabelType"].(string) {
		case "Circuit":
			switch j.Labels["LabelType"].(string) {
			case "Circuit":
				return i.Labels["ASideDeviceName"] == j.Labels["ZSideDeviceName"] && j.Labels["ASideDeviceName"] == i.Labels["ZSideDeviceName"]

			case "Bgp":
				m := (j.Labels["LocalInterface"] == i.Labels["ASideInterface"] && j.Labels["LocalDeviceName"] == i.Labels["ASideDeviceName"]) && (j.Labels["RemoteInterface"] == i.Labels["ZSideInterface"] && j.Labels["RemoteDeviceName"] == i.Labels["ZSideDeviceName"])
				n := (j.Labels["LocalInterface"] == i.Labels["ZSideInterface"] && j.Labels["LocalDeviceName"] == i.Labels["ZSideDeviceName"]) && (j.Labels["RemoteInterface"] == i.Labels["ASideInterface"] && j.Labels["RemoteDeviceName"] == i.Labels["ASideDeviceName"])
				return m || n
			}
		case "Bgp":
			switch j.Labels["LabelType"].(string) {
			case "Circuit":
				m := (i.Labels["LocalInterface"] == j.Labels["ASideInterface"] && i.Labels["LocalDeviceName"] == j.Labels["ASideDeviceName"]) && (i.Labels["RemoteInterface"] == j.Labels["ZSideInterface"] && i.Labels["RemoteDeviceName"] == j.Labels["ZSideDeviceName"])
				n := (i.Labels["LocalInterface"] == j.Labels["ZSideInterface"] && i.Labels["LocalDeviceName"] == j.Labels["ZSideDeviceName"]) && (i.Labels["RemoteInterface"] == j.Labels["ASideInterface"] && i.Labels["RemoteDeviceName"] == j.Labels["ASideDeviceName"])
				return m || n
			case "Bgp":
				return (i.Labels["LocalDeviceName"] == j.Labels["RemoteDeviceName"] && i.Labels["RemoteDeviceName"] == j.Labels["LocalDeviceName"]) || (i.Labels["LocalDeviceName"] == j.Labels["LocalDeviceName"] && i.Labels["RemoteDeviceName"] == j.Labels["RemoteDeviceName"])
			}
		}
		return false
	}
}

func (g *dcCktGrouper) Name() string {
	return g.name
}

func (g *dcCktGrouper) AggDesc(alerts []*models.Alert) string {
	msg := "Affected entities:\n"
	for _, a := range alerts {
		switch a.Labels["LabelType"].(string) {
		case "Circuit":
			msg += fmt.Sprintf("Interface: %s:%s\n", a.Device.String, a.Entity)
		case "Bgp":
			msg += fmt.Sprintf("Bgp Peer: %s\n", a.Entity)
		}
	}
	return msg
}

func (g *dcCktGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	allBgp := true
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		allBgp = allBgp && alert.Labels["LabelType"].(string) == "Bgp"
		valid = append(valid, alert)
	}
	if allBgp {
		glog.V(2).Infof("Ckt Agg: Did not find a dc link alert, skip grouping")
		return []*models.Alert{}
	}
	return valid
}

func init() {
	g := &dcCktGrouper{name: "dc_circuit_down"}
	AddGrouper(g)
}
