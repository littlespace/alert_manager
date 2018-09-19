package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/types"
)

type fibercutGrouper struct {
	name string
}

// grouperFunc defines the condition for circuit endpoints to be considered same to be grouped together
func (g fibercutGrouper) grouperFunc() groupingFunc {
	return func(i, j interface{}) bool {
		var match bool
		if i.(types.Circuit).Provider != "" && j.(types.Circuit).Provider != "" {
			match = match || i.(types.Circuit).Provider == j.(types.Circuit).Provider
		}
		return (match ||
			// 2 ends of same circuit
			(i.(types.Circuit).ASide.Device.Name == j.(types.Circuit).ZSide.Device.Name && i.(types.Circuit).ASide.Interface == j.(types.Circuit).ZSide.Interface) ||
			// phy member of lag
			(i.(types.Circuit).ASide.Device.Name == j.(types.Circuit).ASide.Device.Name && (i.(types.Circuit).ASide.Interface == j.(types.Circuit).ASide.Agg || i.(types.Circuit).ASide.Agg == j.(types.Circuit).ASide.Interface)))

	}
}

func (g *fibercutGrouper) Name() string {
	return g.name
}

func (g *fibercutGrouper) origAlerts(alerts []*models.Alert, group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range alerts {
			if a.Id == p.(types.Circuit).AlertId {
				orig = append(orig, a)
				break
			}
		}
	}
	return orig
}

func (g *fibercutGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	var groupedAlerts [][]*models.Alert
	var ckts []interface{}
	for _, alert := range alerts {
		if !alert.Metadata.Valid || alert.Status != models.Status_ACTIVE {
			continue
		}
		p := types.Circuit{}
		if err := alert.LoadMeta(&p); err != nil {
			glog.Errorf("Fibercut Agg: Unable to unmarshal metadata: %v", err)
			continue
		}
		p.AlertId = alert.Id
		ckts = append(ckts, p)
	}
	if len(ckts) == 0 {
		return groupedAlerts
	}
	glog.V(4).Infof("Fibercut Agg: Now grouping %d alerts", len(alerts))
	groups := group(g.grouperFunc(), ckts)

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
