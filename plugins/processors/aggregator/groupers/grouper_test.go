package groupers

import (
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testDatas = map[string]struct {
	incoming []*models.Alert
	grouped  [][]int64
}{
	"bgp_session": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev1", "LocalInterface": "if1", "LocalIp": "1.1.1.1", "RemoteDeviceName": "dev2", "RemoteInterface": "if2", "RemoteIp": "1.1.1.2"}),
			tu.MockAlert(
				2, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "2", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev2", "LocalInterface": "if2", "LocalIp": "1.1.1.2", "RemoteDeviceName": "dev1", "RemoteInterface": "if1", "RemoteIp": "1.1.1.1"}),
			tu.MockAlert(
				3, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "3", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev1", "LocalInterface": "if1", "LocalIp": "a:b:c::1", "RemoteDeviceName": "dev2", "RemoteInterface": "if2", "RemoteIp": "a:b:c::2"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "4", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev2", "LocalInterface": "if2", "LocalIp": "a:b:c::2", "RemoteDeviceName": "dev1", "RemoteInterface": "if1", "RemoteIp": "a:b:c::1"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "", "dev3", "PeerZ", "src", "scp", "t1", "5", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev3", "LocalInterface": "if3", "LocalIp": "2.2.2.1", "RemoteDeviceName": "dev4", "RemoteInterface": "if4", "RemoteIp": "2.2.2.2"}),
			tu.MockAlert(
				6, "Neteng BGP Down", "", "dev4", "PeerA", "src", "scp", "t1", "6", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev4", "LocalInterface": "if4", "LocalIp": "2.2.2.2", "RemoteDeviceName": "dev3", "RemoteInterface": "if3", "RemoteIp": "2.2.2.1"}),
			tu.MockAlert(
				7, "Neteng BGP Down", "", "dev100", "PeerY", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev100", "LocalInterface": "if1", "LocalIp": "1.1.1.100", "RemoteDeviceName": "dev200", "RemoteInterface": "if2", "RemoteIp": "1.1.1.400"}),
			tu.MockAlert(
				8, "Neteng BGP Down", "", "dev100", "PeerZ", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev100", "LocalInterface": "if1", "LocalIp": "1.1.1.101", "RemoteDeviceName": "dev300", "RemoteInterface": "if2", "RemoteIp": "1.1.1.300"}),
		},
		grouped: [][]int64{
			[]int64{1, 2, 3, 4},
			[]int64{5, 6},
			[]int64{7, 8},
		},
	},
	"dc_circuit_down": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng DC Link Down", "", "dev1", "if1", "src", "scp", "t1", "1", "WARN", []string{"dc", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev1", "ASideInterface": "if1", "ASideAgg": "", "ZSideDeviceName": "dev2", "ZSideInterface": "if2", "ZSideAgg": "", "Role": "dc"}),
			tu.MockAlert(
				2, "Neteng DC Link Down", "", "dev2", "if2", "src", "scp", "t1", "2", "WARN", []string{"dc", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev2", "ASideInterface": "if2", "ASideAgg": "", "ZSideDeviceName": "dev1", "ZSideInterface": "if1", "ZSideAgg": "", "Role": "dc"}),
			tu.MockAlert(
				3, "Neteng DC Link Down", "", "dev1", "if3", "src", "scp", "t1", "3", "WARN", []string{"dc", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev1", "ASideInterface": "if3", "ASideAgg": "", "ZSideDeviceName": "dev3", "ZSideInterface": "if1", "ZSideAgg": "", "Role": "dc"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev1", "LocalInterface": "if1", "RemoteDeviceName": "dev2", "RemoteInterface": "if2"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"LabelType": "Bgp", "Type": "ebgp", "LocalDeviceName": "dev2", "LocalInterface": "if2", "RemoteDeviceName": "dev1", "RemoteInterface": "if1"}),
		},
		grouped: [][]int64{
			[]int64{1, 2, 4, 5},
			[]int64{3},
		},
	},
	"fibercut": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng BB Link Down", "", "dev1", "if1", "src", "scp", "t1", "1", "WARN", []string{"bb", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev1", "ASideInterface": "if1", "ASideAgg": "ae1", "ZSideDeviceName": "dev2", "ZSideInterface": "if2", "ZSideAgg": "ae2", "Role": "bb", "Provider": "telstra"}),
			tu.MockAlert(
				2, "Neteng BB Link Down", "", "dev2", "if2", "src", "scp", "t1", "2", "WARN", []string{"bb", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev2", "ASideInterface": "if2", "ASideAgg": "ae2", "ZSideDeviceName": "dev1", "ZSideInterface": "if1", "ZSideAgg": "ae1", "Role": "bb", "Provider": "telstra"}),
			tu.MockAlert(
				3, "Neteng BB Link Down", "", "dev3", "if3", "src", "scp", "t1", "3", "WARN", []string{"bb", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev3", "ASideInterface": "if3", "ASideAgg": "ae3", "ZSideDeviceName": "dev4", "ZSideInterface": "if4", "ZSideAgg": "ae4", "Role": "bb", "Provider": "level3"}),
			tu.MockAlert(
				4, "Neteng BB Link Down", "", "dev4", "if4", "src", "scp", "t1", "4", "WARN", []string{"bb", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev4", "ASideInterface": "if4", "ASideAgg": "ae4", "ZSideDeviceName": "dev3", "ZSideInterface": "if3", "ZSideAgg": "ae3", "Role": "bb", "Provider": "level3"}),
			tu.MockAlert(
				5, "Neteng BB Agg Link Down", "", "dev4", "ae4", "src", "scp", "t1", "5", "WARN", []string{"bb", "link"},
				models.Labels{"LabelType": "Circuit", "ASideDeviceName": "dev4", "ASideInterface": "ae4", "ASideAgg": ""}),
		},
		grouped: [][]int64{
			[]int64{1, 2},
			[]int64{3, 4, 5},
		},
	},
	"default_label_grouper": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng High Device CPU", "", "dev1", "dev1", "src1", "device", "t1", "1", "WARN", []string{}, nil),
			tu.MockAlert(
				2, "Neteng LC Fail", "", "dev1", "dev1", "src2", "device", "t1", "2", "WARN", []string{}, nil),
			tu.MockAlert(
				3, "Neteng Device Effed", "", "dev2", "dev2", "src1", "device", "t1", "3", "WARN", []string{}, nil),
			tu.MockAlert(
				4, "Neteng Fan Down", "", "dev2", "if2", "src1", "device", "t1", "4", "WARN", []string{}, nil),
		},
		grouped: [][]int64{
			[]int64{1, 2},
			[]int64{3, 4},
		},
	},
}

func TestGrouping(t *testing.T) {
	for name, datas := range testDatas {
		grouper := AllGroupers[name]
		if g, ok := grouper.(*LabelGrouper); ok {
			g.SetGroupby([]string{"device"})
			grouper = g
		}
		var aggregated [][]int64
		for _, a := range datas.incoming {
			a.ExtendLabels()
		}
		for _, group := range DoGrouping(grouper, datas.incoming) {
			var groupedIds []int64
			for _, a := range group {
				groupedIds = append(groupedIds, a.Id)
			}
			aggregated = append(aggregated, groupedIds)
		}
		assert.ElementsMatchf(t, aggregated, datas.grouped,
			"Grouper: %s, Expected: %v, Got: %v", name, datas.grouped, aggregated)
	}
}
