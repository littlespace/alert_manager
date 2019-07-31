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
		if i.Labels["provider"] != nil && j.Labels["provider"] != nil {
			match = match || i.Labels["provider"] == j.Labels["provider"]
		}
		if i.Labels["cktId"] != nil && j.Labels["cktId"] != nil {
			match = match || i.Labels["cktId"] == j.Labels["cktId"]
		}
		return (match ||
			// 2 ends of same circuit
			(i.Labels["aSideDeviceName"] == j.Labels["zSideDeviceName"] && i.Labels["aSideInterface"] == j.Labels["zSideInterface"]) ||
			// phy member of lag
			(i.Labels["aSideDeviceName"] == j.Labels["aSideDeviceName"] && (i.Labels["aSideInterface"] == j.Labels["aSideAgg"] || i.Labels["aSideAgg"] == j.Labels["aSideInterface"])))

	}
}

func (g fibercutGrouper) Name() string {
	return g.name
}

func (g fibercutGrouper) AggDesc(alerts []*models.Alert) string {

	type AlertInfo struct {
		provider, aSideDev, aSideInt, zSideDev, zSideInt string
	}
	m := make(map[string]AlertInfo)

	msg := "Affected Interfaces:\n"
	for _, a := range alerts {
		if _, ok := m[a.Labels["cktId"].(string)]; !ok {
			m[a.Labels["cktId"].(string)] = AlertInfo{a.Labels["provider"].(string), a.Labels["aSideDeviceName"].(string), a.Labels["aSideInterface"].(string),
				a.Labels["zSideDeviceName"].(string), a.Labels["zSideInterface"].(string)}
		}
	}

	for k, v := range m {
		msg += fmt.Sprintf("Provider: %s, CktId: %s, A-Side: %s:%s, Z-Side: %s:%s\n", v.provider, k, v.aSideDev, v.aSideInt, v.zSideDev, v.zSideInt)
	}
	return msg
}

func (g fibercutGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	for _, alert := range alerts {
		if len(alert.Labels) == 0 || alert.Labels["labelType"] == nil || alert.Status != models.Status_ACTIVE {
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
