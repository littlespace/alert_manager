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
			i.Labels["localDeviceName"] == j.Labels["remoteDeviceName"] && i.Labels["remoteDeviceName"] == j.Labels["localDeviceName"] ||
			// two sessions from/to same device
			i.Labels["localDeviceName"] == j.Labels["localDeviceName"] && i.Labels["remoteDeviceName"] == j.Labels["remoteDeviceName"])
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
		if len(alert.Labels) == 0 || alert.Labels["labelType"] == nil || alert.Status != models.Status_ACTIVE {
			continue
		}
		if alert.Labels["labelType"].(string) != "Bgp" {
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
