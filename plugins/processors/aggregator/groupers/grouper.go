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

//  generic grouping func
func group(f groupingFunc, items []interface{}) [][]interface{} {
	groups := [][]interface{}{[]interface{}{items[0]}}
	for i := 1; i < len(items); i++ {
		var found bool
		for j := 0; j < len(groups); j++ {
			if f(items[i], groups[j][0]) {
				found = true
				groups[j] = append(groups[j], items[i])
				break
			}
		}
		if !found {
			groups = append(groups, []interface{}{items[i]})
		}
	}
	return groups
}
