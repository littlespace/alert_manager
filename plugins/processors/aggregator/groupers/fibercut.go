package groupers

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type fibercutGrouper struct {
	name string
}

// grouperFunc defines the condition for circuit endpoints to be considered same to be grouped together
func (g fibercutGrouper) grouperFunc() groupingFunc {
	return func(i, j interface{}) bool {
		var match bool
		if i.(Circuit).Provider != "" && j.(Circuit).Provider != "" {
			match = match || i.(Circuit).Provider == j.(Circuit).Provider
		}
		return (match ||
			// 2 ends of same circuit
			i.(Circuit).ASide == j.(Circuit).ZSide ||
			// phy member of lag
			i.(Circuit).ASide.Device == j.(Circuit).ASide.Device && (i.(Circuit).ASide.Interface == j.(Circuit).ASide.Agg || i.(Circuit).ASide.Agg == j.(Circuit).ASide.Interface))

	}
}

func (g *fibercutGrouper) Name() string {
	return g.name
}

func (g *fibercutGrouper) origAlerts(alerts []*models.Alert, group []interface{}) []*models.Alert {
	var orig []*models.Alert
	for _, p := range group {
		for _, a := range alerts {
			if a.Id == p.(Circuit).AlertId {
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
		p := Circuit{}
		if err := json.Unmarshal([]byte(alert.Metadata.String), &p); err != nil {
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
