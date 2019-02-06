package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type bgpGrouper struct {
	name string
}

// grouperFunc defines the condition for two bgp alerts to be considered same to be grouped together
func (g bgpGrouper) GrouperFunc() GroupingFunc {
	return func(i, j *models.Alert) bool {
		return (
		// same host
		i.Device.String == j.Device.String ||
			// two ends of the same session
			i.Labels["LocalDeviceName"] == j.Labels["RemoteDeviceName"] && i.Labels["RemoteDeviceName"] == j.Labels["LocalDeviceName"] ||
			// two sessions from/to same device
			i.Labels["LocalDeviceName"] == j.Labels["LocalDeviceName"] && i.Labels["RemoteDeviceName"] == j.Labels["RemoteDeviceName"])
	}
}

func (g *bgpGrouper) Name() string {
	return g.name
}

func (g *bgpGrouper) AggDesc(alerts []*models.Alert) string {
	msg := "Sessions down:\n"
	for _, a := range alerts {
		msg += a.Description + "\n"
	}
	return msg
}

func (g *bgpGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		if alert.Labels["LabelType"].(string) != "Bgp" {
			glog.V(2).Infof("Bgp Agg: Found non bgp alert, skip grouping")
			return []*models.Alert{}
		}
		valid = append(valid, alert)
	}
	return valid
}

func init() {
	g := &bgpGrouper{name: "bgp_session"}
	AddGrouper(g)
}
