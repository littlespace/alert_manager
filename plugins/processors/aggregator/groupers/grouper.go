package groupers

import (
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type Grouper interface {
	Name() string
	AggDesc(alerts []*models.Alert) string
	AggLabels(alerts []*models.Alert) models.Labels
	Valid(alerts []*models.Alert) []*models.Alert
	GrouperFunc() GroupingFunc
}

var AllGroupers = make(map[string]Grouper)

func AddGrouper(grp Grouper) {
	AllGroupers[grp.Name()] = grp
}

type GroupingFunc func(i, j *models.Alert) bool

// generic grouping func
// Compares items of a slice in pairs, evaluating against the result of the groupingFunc.
// If two items are equal, group them into the same output slice. If not, then they
// are grouped into separate output slices.
func group(gf GroupingFunc, items []*models.Alert) [][]*models.Alert {
	groups := [][]*models.Alert{[]*models.Alert{items[0]}}
	for i := 1; i < len(items); i++ {
		var found bool
		for j := 0; j < len(groups); j++ {
			for _, k := range groups[j] {
				if gf(items[i], k) {
					found = true
					groups[j] = append(groups[j], items[i])
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			groups = append(groups, []*models.Alert{items[i]})
		}
	}
	return groups
}

func DoGrouping(g Grouper, alerts []*models.Alert) [][]*models.Alert {
	validAlerts := g.Valid(alerts)
	if len(validAlerts) == 0 {
		return [][]*models.Alert{}
	}
	glog.V(4).Infof("%s: Now grouping %d alerts", g.Name(), len(validAlerts))
	return group(g.GrouperFunc(), validAlerts)
}
