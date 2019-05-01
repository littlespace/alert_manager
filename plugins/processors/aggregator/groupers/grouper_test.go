package groupers

import (
	"testing"

	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

var testDatas = map[string]struct {
	incoming []*models.Alert
	grouped  [][]int64
}{
	"bgp_session": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev1", "localInterface": "if1", "localIp": "1.1.1.1", "remoteDeviceName": "dev2", "remoteInterface": "if2", "remoteIp": "1.1.1.2"}),
			tu.MockAlert(
				2, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "2", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev2", "localInterface": "if2", "localIp": "1.1.1.2", "remoteDeviceName": "dev1", "remoteInterface": "if1", "remoteIp": "1.1.1.1"}),
			tu.MockAlert(
				3, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "3", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev1", "localInterface": "if1", "localIp": "a:b:c::1", "remoteDeviceName": "dev2", "remoteInterface": "if2", "remoteIp": "a:b:c::2"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "4", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev2", "localInterface": "if2", "localIp": "a:b:c::2", "remoteDeviceName": "dev1", "remoteInterface": "if1", "remoteIp": "a:b:c::1"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "", "dev3", "PeerZ", "src", "scp", "t1", "5", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev3", "localInterface": "if3", "localIp": "2.2.2.1", "remoteDeviceName": "dev4", "remoteInterface": "if4", "remoteIp": "2.2.2.2"}),
			tu.MockAlert(
				6, "Neteng BGP Down", "", "dev4", "PeerA", "src", "scp", "t1", "6", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev4", "localInterface": "if4", "localIp": "2.2.2.2", "remoteDeviceName": "dev3", "remoteInterface": "if3", "remoteIp": "2.2.2.1"}),
			tu.MockAlert(
				7, "Neteng BGP Down", "", "dev100", "PeerY", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev100", "localInterface": "if1", "localIp": "1.1.1.100", "remoteDeviceName": "dev200", "remoteInterface": "if2", "remoteIp": "1.1.1.400"}),
			tu.MockAlert(
				8, "Neteng BGP Down", "", "dev100", "PeerZ", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev100", "localInterface": "if1", "localIp": "1.1.1.101", "remoteDeviceName": "dev300", "remoteInterface": "if2", "remoteIp": "1.1.1.300"}),
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
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev1", "aSideInterface": "if1", "aSideAgg": "", "zSideDeviceName": "dev2", "zSideInterface": "if2", "zSideAgg": "", "role": "dc"}),
			tu.MockAlert(
				2, "Neteng DC Link Down", "", "dev2", "if2", "src", "scp", "t1", "2", "WARN", []string{"dc", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev2", "aSideInterface": "if2", "aSideAgg": "", "zSideDeviceName": "dev1", "zSideInterface": "if1", "zSideAgg": "", "role": "dc"}),
			tu.MockAlert(
				3, "Neteng DC Link Down", "", "dev1", "if3", "src", "scp", "t1", "3", "WARN", []string{"dc", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev1", "aSideInterface": "if3", "aSideAgg": "", "zSideDeviceName": "dev3", "zSideInterface": "if1", "zSideAgg": "", "role": "dc"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "", "dev1", "PeerX", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev1", "localInterface": "if1", "remoteDeviceName": "dev2", "remoteInterface": "if2"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "", "dev2", "PeerY", "src", "scp", "t1", "1", "WARN", []string{"bgp", "dc"},
				models.Labels{"labelType": "Bgp", "type": "ebgp", "localDeviceName": "dev2", "localInterface": "if2", "remoteDeviceName": "dev1", "remoteInterface": "if1"}),
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
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev1", "aSideInterface": "if1", "aSideAgg": "ae1", "zSideDeviceName": "dev2", "zSideInterface": "if2", "zSideAgg": "ae2", "role": "bb", "provider": "telstra"}),
			tu.MockAlert(
				2, "Neteng BB Link Down", "", "dev2", "if2", "src", "scp", "t1", "2", "WARN", []string{"bb", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev2", "aSideInterface": "if2", "aSideAgg": "ae2", "zSideDeviceName": "dev1", "zSideInterface": "if1", "zSideAgg": "ae1", "role": "bb", "provider": "telstra"}),
			tu.MockAlert(
				3, "Neteng BB Link Down", "", "dev3", "if3", "src", "scp", "t1", "3", "WARN", []string{"bb", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev3", "aSideInterface": "if3", "aSideAgg": "ae3", "zSideDeviceName": "dev4", "zSideInterface": "if4", "zSideAgg": "ae4", "role": "bb", "provider": "level3"}),
			tu.MockAlert(
				4, "Neteng BB Link Down", "", "dev4", "if4", "src", "scp", "t1", "4", "WARN", []string{"bb", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev4", "aSideInterface": "if4", "aSideAgg": "ae4", "zSideDeviceName": "dev3", "zSideInterface": "if3", "zSideAgg": "ae3", "role": "bb", "provider": "level3"}),
			tu.MockAlert(
				5, "Neteng BB Agg Link Down", "", "dev4", "ae4", "src", "scp", "t1", "5", "WARN", []string{"bb", "link"},
				models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev4", "aSideInterface": "ae4", "aSideAgg": ""}),
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

func TestGroupingDCLinkIbgp(t *testing.T) {
	alerts := []*models.Alert{
		tu.MockAlert(
			1, "Neteng DC Link Down", "", "dev1", "if1", "src", "scp", "t1", "1", "WARN", []string{"dc", "link"},
			models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev1", "aSideInterface": "if1", "aSideAgg": "", "zSideDeviceName": "dev2", "zSideInterface": "if2", "zSideAgg": "", "role": "dc"}),
		tu.MockAlert(
			2, "Neteng DC Link Down", "", "dev2", "if2", "src", "scp", "t1", "2", "WARN", []string{"dc", "link"},
			models.Labels{"labelType": "Circuit", "aSideDeviceName": "dev2", "aSideInterface": "if2", "aSideAgg": "", "zSideDeviceName": "dev1", "zSideInterface": "if1", "zSideAgg": "", "role": "dc"}),
		tu.MockAlert(
			3, "Neteng BGP Down", "", "dev3", "PeerX", "src", "scp", "t1", "1", "WARN", []string{"bgp", "bb"},
			models.Labels{"labelType": "Bgp", "type": "ibgp", "localDeviceName": "dev3", "localInterface": "if1", "remoteDeviceName": "dev4", "remoteInterface": "if2"}),
		tu.MockAlert(
			4, "Neteng BGP Down", "", "dev4", "PeerY", "src", "scp", "t1", "1", "WARN", []string{"bgp", "bb"},
			models.Labels{"labelType": "Bgp", "type": "ibgp", "localDeviceName": "dev4", "localInterface": "if2", "remoteDeviceName": "dev3", "remoteInterface": "if1"}),
	}
	for _, a := range alerts {
		a.ExtendLabels()
	}
	dclink_grouper := AllGroupers["dc_circuit_down"]
	var aggregated [][]int64
	for _, group := range DoGrouping(dclink_grouper, alerts) {
		var groupedIds []int64
		for _, a := range group {
			groupedIds = append(groupedIds, a.Id)
		}
		aggregated = append(aggregated, groupedIds)
	}
	expected := [][]int64{[]int64{1, 2}}
	assert.ElementsMatchf(t, aggregated, expected, "Expected: %v, Got: %v", expected, aggregated)
	bgp_grouper := AllGroupers["bgp_session"]
	aggregated = [][]int64{}
	for _, group := range DoGrouping(bgp_grouper, alerts) {
		var groupedIds []int64
		for _, a := range group {
			groupedIds = append(groupedIds, a.Id)
		}
		aggregated = append(aggregated, groupedIds)
	}
	expected = [][]int64{[]int64{3, 4}}
	assert.ElementsMatchf(t, aggregated, expected, "Expected: %v, Got: %v", expected, aggregated)
}
