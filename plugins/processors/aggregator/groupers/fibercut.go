package groupers

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type fibercutGrouper struct {
	name string
}

// grouperFunc defines the condition for circuit endpoints to be considered same to be grouped together
func (g fibercutGrouper) GrouperFunc() GroupingFunc {
	return func(i, j *models.Alert) bool {
		var match bool
		if i.Labels["Provider"] != nil && j.Labels["Provider"] != nil {
			match = match || i.Labels["Provider"] == j.Labels["Provider"]
		}
		if i.Labels["CktId"] != nil && j.Labels["CktId"] != nil {
			match = match || i.Labels["CktId"] == j.Labels["CktId"]
		}
		return (match ||
			// 2 ends of same circuit
			(i.Labels["ASideDeviceName"] == j.Labels["ZSideDeviceName"] && i.Labels["ASideInterface"] == j.Labels["ZSideInterface"]) ||
			// phy member of lag
			(i.Labels["ASideDeviceName"] == j.Labels["ASideDeviceName"] && (i.Labels["ASideInterface"] == j.Labels["ASideAgg"] || i.Labels["ASideAgg"] == j.Labels["ASideInterface"])))

	}
}

func (g *fibercutGrouper) Name() string {
	return g.name
}

func (g *fibercutGrouper) AggDesc(alerts []*models.Alert) string {
	msg := "Affected Interfaces:\n"
	for _, a := range alerts {
		msg += fmt.Sprintf(
			"%s:%s, Provider: %s, CktId: %s\n",
			a.Device.String, a.Entity, a.Labels["Provider"].(string), a.Labels["CktId"].(string),
		)
	}
	return msg
}

func (g *fibercutGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Status != models.Status_ACTIVE {
			continue
		}
		valid = append(valid, alert)
	}
	return valid
}

func init() {
	g := &fibercutGrouper{name: "fibercut"}
	AddGrouper(g)
}
