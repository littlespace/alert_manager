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
				1, "Neteng BGP Down", "dev1", "PeerX", "src", "scp", "1", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev1", LocalInterface: "if1", LocalIp: "1.1.1.1", RemoteDevice: "dev2", RemoteInterface: "if2", RemoteIp: "1.1.1.2"}),
			tu.MockAlert(
				2, "Neteng BGP Down", "dev2", "PeerY", "src", "scp", "2", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev2", LocalInterface: "if2", LocalIp: "1.1.1.2", RemoteDevice: "dev1", RemoteInterface: "if1", RemoteIp: "1.1.1.1"}),
			tu.MockAlert(
				3, "Neteng BGP Down", "dev1", "PeerX", "src", "scp", "3", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev1", LocalInterface: "if1", LocalIp: "a:b:c::1", RemoteDevice: "dev2", RemoteInterface: "if2", RemoteIp: "a:b:c::2"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "dev2", "PeerY", "src", "scp", "4", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev2", LocalInterface: "if2", LocalIp: "a:b:c::2", RemoteDevice: "dev1", RemoteInterface: "if1", RemoteIp: "a:b:c::1"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "dev1", "PeerZ", "src", "scp", "5", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev1", LocalInterface: "if3", LocalIp: "2.2.2.1", RemoteDevice: "dev3", RemoteInterface: "if4", RemoteIp: "2.2.2.2"}),
			tu.MockAlert(
				6, "Neteng BGP Down", "dev3", "PeerA", "src", "scp", "6", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev3", LocalInterface: "if4", LocalIp: "2.2.2.2", RemoteDevice: "dev1", RemoteInterface: "if3", RemoteIp: "2.2.2.1"}),
		},
		grouped: [][]int64{
			[]int64{1, 2, 3, 4},
			[]int64{5, 6},
		},
	},
	"dc_circuit_down": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng DC Link Down", "dev1", "if1", "src", "scp", "1", []string{"dc", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev1", Interface: "if1", Agg: ""}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev2", Interface: "if2", Agg: ""}, Role: "dc"}),
			tu.MockAlert(
				2, "Neteng DC Link Down", "dev2", "if2", "src", "scp", "2", []string{"dc", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev2", Interface: "if2", Agg: ""}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev1", Interface: "if1", Agg: ""}, Role: "dc"}),
			tu.MockAlert(
				3, "Neteng DC Link Down", "dev1", "if3", "src", "scp", "3", []string{"dc", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev1", Interface: "if3", Agg: ""}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev3", Interface: "if1", Agg: ""}, Role: "dc"}),
			tu.MockAlert(
				4, "Neteng BGP Down", "dev1", "PeerX", "src", "scp", "1", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev1", LocalInterface: "if1", RemoteDevice: "dev2", RemoteInterface: "if2"}),
			tu.MockAlert(
				5, "Neteng BGP Down", "dev2", "PeerY", "src", "scp", "1", []string{"bgp", "dc"},
				BgpPeer{Type: "ebgp", LocalDevice: "dev2", LocalInterface: "if2", RemoteDevice: "dev1", RemoteInterface: "if1"}),
		},
		grouped: [][]int64{
			[]int64{1, 2, 4, 5},
			[]int64{3},
		},
	},
	"fibercut": {
		incoming: []*models.Alert{
			tu.MockAlert(
				1, "Neteng BB Link Down", "dev1", "if1", "src", "scp", "1", []string{"bb", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev1", Interface: "if1", Agg: "ae1"}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev2", Interface: "if2", Agg: "ae2"}, Role: "bb", Provider: "telstra"}),
			tu.MockAlert(
				2, "Neteng BB Link Down", "dev2", "if2", "src", "scp", "2", []string{"bb", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev2", Interface: "if2", Agg: "ae2"}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev1", Interface: "if1", Agg: "ae1"}, Role: "bb", Provider: "telstra"}),
			tu.MockAlert(
				3, "Neteng BB Link Down", "dev3", "if3", "src", "scp", "3", []string{"bb", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev3", Interface: "if3", Agg: "ae3"}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev4", Interface: "if4", Agg: "ae4"}, Role: "bb", Provider: "level3"}),
			tu.MockAlert(
				4, "Neteng BB Link Down", "dev4", "if4", "src", "scp", "4", []string{"bb", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev4", Interface: "if4", Agg: "ae4"}, ZSide: struct{ Device, Interface, Agg string }{Device: "dev3", Interface: "if3", Agg: "ae3"}, Role: "bb", Provider: "level3"}),
			tu.MockAlert(
				5, "Neteng BB Link Down", "dev4", "ae4", "src", "scp", "5", []string{"bb", "link"},
				Circuit{ASide: struct{ Device, Interface, Agg string }{Device: "dev4", Interface: "ae4", Agg: ""}}),
		},
		grouped: [][]int64{
			[]int64{1, 2},
			[]int64{3, 4, 5},
		},
	},
}

func TestGrouping(t *testing.T) {
	for name, datas := range testDatas {
		grouper := AllGroupers[name]
		var aggregated [][]int64
		for _, group := range grouper.DoGrouping(datas.incoming) {
			var groupedIds []int64
			for _, a := range group {
				groupedIds = append(groupedIds, a.Id)
			}
			aggregated = append(aggregated, groupedIds)
		}
		assert.ElementsMatch(t, aggregated, datas.grouped)
	}
}
