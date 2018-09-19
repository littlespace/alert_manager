package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/types"
)

type dcCktGrouper struct {
	name string
}

// grouperFunc defines the condition for two circuit related objects  to be considered same to be grouped together
// Currently, it considers a/z endpoints as well as bgp peers
func (g dcCktGrouper) grouperFunc() groupingFunc {
	return func(i, j interface{}) bool {
		switch i := i.(type) {
		case types.Circuit:
			switch j := j.(type) {
			case types.Circuit:
				return i.ASide == j.ZSide && j.ASide == i.ZSide
			case types.BgpPeer:

				m := (j.LocalInterface == i.ASide.Interface && j.LocalDevice.Name == i.ASide.Device.Name) && (j.RemoteInterface == i.ZSide.Interface && j.RemoteDevice.Name == i.ZSide.Device.Name)
				n := (j.LocalInterface == i.ZSide.Interface && j.LocalDevice.Name == i.ZSide.Device.Name) && (j.RemoteInterface == i.ASide.Interface && j.RemoteDevice.Name == i.ASide.Device.Name)
				return m || n
			}
		case types.BgpPeer:
			switch j := j.(type) {
			case types.Circuit:
				m := (i.LocalInterface == j.ASide.Interface && i.LocalDevice.Name == j.ASide.Device.Name) && (i.RemoteInterface == j.ZSide.Interface && i.RemoteDevice.Name == j.ZSide.Device.Name)
				n := (i.LocalInterface == j.ZSide.Interface && i.LocalDevice.Name == j.ZSide.Device.Name) && (i.RemoteInterface == j.ASide.Interface && i.RemoteDevice.Name == j.ASide.Device.Name)
				return m || n
			case types.BgpPeer:
				return (i.LocalDevice.Name == j.RemoteDevice.Name && i.RemoteDevice.Name == j.LocalDevice.Name) || (i.LocalDevice.Name == j.LocalDevice.Name && i.RemoteDevice.Name == j.RemoteDevice.Name)
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
			if c, ok := p.(types.Circuit); ok {
				cond = a.Id == c.AlertId
			} else if b, ok := p.(types.BgpPeer); ok {
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
			p := types.BgpPeer{}
			if err := alert.LoadMeta(&p); err != nil {
				glog.Errorf("Ckt Agg: Unable to unmarshal metadata: %v", err)
				continue
			}
			p.AlertId = alert.Id
			entities = append(entities, p)
		} else {
			c := types.Circuit{}
			if err := alert.LoadMeta(&c); err != nil {
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
	if len(entities) == 0 {
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
	AddGrouper(g)
}
