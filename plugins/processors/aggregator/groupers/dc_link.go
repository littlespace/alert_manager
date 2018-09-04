package groupers

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type dcCktGrouper struct {
	name string
}

// grouperFunc defines the condition for two circuit related objects  to be considered same to be grouped together
// Currently, it considers a/z endpoints as well as bgp peers
func (g dcCktGrouper) grouperFunc() groupingFunc {
	return func(i, j interface{}) bool {
		switch i := i.(type) {
		case Circuit:
			switch j := j.(type) {
			case Circuit:
				return i.ASide == j.ZSide && j.ASide == i.ZSide
			case BgpPeer:

				m := (j.LocalInterface == i.ASide.Interface && j.LocalDevice == i.ASide.Device) && (j.RemoteInterface == i.ZSide.Interface && j.RemoteDevice == i.ZSide.Device)
				n := (j.LocalInterface == i.ZSide.Interface && j.LocalDevice == i.ZSide.Device) && (j.RemoteInterface == i.ASide.Interface && j.RemoteDevice == i.ASide.Device)
				return m || n
			}
		case BgpPeer:
			switch j := j.(type) {
			case Circuit:
				m := (i.LocalInterface == j.ASide.Interface && i.LocalDevice == j.ASide.Device) && (i.RemoteInterface == j.ZSide.Interface && i.RemoteDevice == j.ZSide.Device)
				n := (i.LocalInterface == j.ZSide.Interface && i.LocalDevice == j.ZSide.Device) && (i.RemoteInterface == j.ASide.Interface && i.RemoteDevice == j.ASide.Device)
				return m || n
			case BgpPeer:
				return (i.LocalDevice == j.RemoteDevice && i.RemoteDevice == j.LocalDevice) || (i.LocalDevice == j.LocalDevice && i.RemoteDevice == j.RemoteDevice)
			}
		}
		return false
	}
}

func (g *dcCktGrouper) Name() string {
	return g.name
}

func (g *dcCktGrouper) origAlerts(alerts []*models.Alert, group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
	innerfor:
		for _, a := range alerts {
			var cond bool
			if c, ok := p.(Circuit); ok {
				cond = a.Id == c.AlertId
			} else if b, ok := p.(BgpPeer); ok {
				cond = a.Id == b.AlertId
			}
			if cond {
				orig = append(orig, a)
				break innerfor
			}
		}
	}
	return orig
}

func (g *dcCktGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	var entities []interface{}
	var groupedAlerts [][]*models.Alert
	allBgp := true
	for _, alert := range alerts {
		if !alert.Metadata.Valid || alert.Status != models.Status_ACTIVE {
			continue
		}
		allBgp = allBgp && alert.HasTags("bgp")
		if alert.HasTags("bgp") {
			p := BgpPeer{}
			if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
				glog.Errorf("Ckt Agg: Unable to unmarshal metadata: %v", err)
				continue
			}
			p.AlertId = alert.Id
			entities = append(entities, p)
		} else {
			c := Circuit{}
			if err := json.Unmarshal([]byte(alert.Metadata.String), &c); err != nil {
				glog.Errorf("Ckt Agg: Unable to unmarshal metadata: %v", err)
				continue
			}
			c.AlertId = alert.Id
			entities = append(entities, c)
		}
	}
	if allBgp {
		glog.V(2).Infof("Ckt Agg: Did not find a dc link alert, skip grouping")
		return groupedAlerts
	}
	glog.V(4).Infof("Ckt Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), entities)
	//TODO : group by device

	for _, group := range groups {
		orig := g.origAlerts(alerts, group)
		groupedAlerts = append(groupedAlerts, orig)
	}
	return groupedAlerts
}

func init() {
	g := &dcCktGrouper{name: "dc_circuit_down"}
	addGrouper(g)
}
