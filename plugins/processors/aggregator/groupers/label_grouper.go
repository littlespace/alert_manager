package groupers

import (
	"fmt"

	"github.com/mayuresh82/alert_manager/internal/models"
)

type LabelGrouper struct {
	name    string
	Groupby []string
}

func (g LabelGrouper) GrouperFunc() GroupingFunc {
	return func(i, j *models.Alert) bool {
		for _, k := range g.Groupby {
			if i.Labels[k] != j.Labels[k] {
				return false
			}
		}
		return true
	}
}

func (g LabelGrouper) Name() string {
	return g.name
}

func (g *LabelGrouper) SetGroupby(groupBy []string) {
	g.Groupby = g.Groupby[:0]
	g.Groupby = append(g.Groupby, groupBy...)
}

func (g LabelGrouper) AggLabels(alerts []*models.Alert) models.Labels {
	l := make(models.Labels)
	var entities []string
	for _, a := range alerts {
		if a.Device.Valid {
			entities = append(entities, fmt.Sprintf("%s:%s", a.Device.String, a.Entity))
		} else {
			entities = append(entities, a.Entity)
		}
	}
	l["entities"] = entities
	return l
}

func (g LabelGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	var valid []*models.Alert
	for _, alert := range alerts {
		if alert.Status != models.Status_ACTIVE {
			continue
		}
		var skip bool
		for _, k := range g.Groupby {
			if _, ok := alert.Labels[k]; !ok {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		valid = append(valid, alert)
	}
	return valid
}

func init() {
	g := &LabelGrouper{name: "default_label_grouper"}
	AddGrouper(g)
}
