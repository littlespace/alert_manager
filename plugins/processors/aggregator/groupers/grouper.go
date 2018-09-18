package groupers

import (
	"github.com/mayuresh82/alert_manager/internal/models"
)

type Grouper interface {
	Name() string
	DoGrouping(alerts []*models.Alert) [][]*models.Alert
}

var AllGroupers = make(map[string]Grouper)

func AddGrouper(grp Grouper) {
	AllGroupers[grp.Name()] = grp
}

type groupingFunc func(i, j interface{}) bool

// generic grouping func
// Compares items of a slice in pairs, evaluating against the result of the groupingFunc.
// If two items are equal, group them into the same output slice. If not, then they
// are grouped into separate output slices.
func group(gf groupingFunc, items []interface{}) [][]interface{} {
	groups := [][]interface{}{[]interface{}{items[0]}}
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
			groups = append(groups, []interface{}{items[i]})
		}
	}
	return groups
}
