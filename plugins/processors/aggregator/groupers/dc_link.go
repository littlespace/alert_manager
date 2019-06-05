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
		switch i.Labels["labelType"].(string) {
		case "Circuit":
			switch j.Labels["labelType"].(string) {
			case "Circuit":
				return i.Labels["aSideDeviceName"] == j.Labels["zSideDeviceName"] && j.Labels["aSideDeviceName"] == i.Labels["zSideDeviceName"]

			case "Bgp":
				m := (j.Labels["localInterface"] == i.Labels["aSideInterface"] && j.Labels["localDeviceName"] == i.Labels["aSideDeviceName"]) && (j.Labels["remoteInterface"] == i.Labels["zSideInterface"] && j.Labels["remoteDeviceName"] == i.Labels["zSideDeviceName"])
				n := (j.Labels["localInterface"] == i.Labels["zSideInterface"] && j.Labels["localDeviceName"] == i.Labels["zSideDeviceName"]) && (j.Labels["remoteInterface"] == i.Labels["aSideInterface"] && j.Labels["remoteDeviceName"] == i.Labels["aSideDeviceName"])
				return m || n
			}
		case "Bgp":
			switch j.Labels["labelType"].(string) {
			case "Circuit":
				m := (i.Labels["localInterface"] == j.Labels["aSideInterface"] && i.Labels["localDeviceName"] == j.Labels["aSideDeviceName"]) && (i.Labels["remoteInterface"] == j.Labels["zSideInterface"] && i.Labels["remoteDeviceName"] == j.Labels["zSideDeviceName"])
				n := (i.Labels["localInterface"] == j.Labels["zSideInterface"] && i.Labels["localDeviceName"] == j.Labels["zSideDeviceName"]) && (i.Labels["remoteInterface"] == j.Labels["aSideInterface"] && i.Labels["remoteDeviceName"] == j.Labels["aSideDeviceName"])
				return m || n
			case "Bgp":
				return (i.Labels["localDeviceName"] == j.Labels["remoteDeviceName"] && i.Labels["remoteDeviceName"] == j.Labels["LocalDeviceName"]) || (i.Labels["localDeviceName"] == j.Labels["localDeviceName"] && i.Labels["remoteDeviceName"] == j.Labels["remoteDeviceName"])
			}
		}
		return false
	}
}

func (g dcCktGrouper) Name() string {
	return g.name
}

func (g dcCktGrouper) AggDesc(alerts []*models.Alert) string {
	msg := "Affected entities:\n"
	for _, a := range alerts {
		switch a.Labels["labelType"].(string) {
		case "Circuit":
			msg += fmt.Sprintf("Interface: %s:%s\n", a.Device.String, a.Entity)
		case "Bgp":
			msg += fmt.Sprintf("Bgp Peer: %s\n", a.Entity)
		}
	}
	return msg
}

func (g dcCktGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	allBgp := true
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Labels["labelType"] == nil || alert.Status != models.Status_ACTIVE {
			continue
		}
		// ibgp alerts are possibly not due to dc link events, so let the bgp grouper handle them
		if alert.Labels["labelType"].(string) == "Bgp" && alert.Labels["type"].(string) == "ibgp" {
			continue
		}
		allBgp = allBgp && alert.Labels["labelType"].(string) == "Bgp"
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
