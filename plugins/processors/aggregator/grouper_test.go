package aggregator

import (
	"context"
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
				Circuit{ASide: struct{ Device, Interface, Description string }{Device: "dev1", Interface: "if1", Description: ""}, ZSide: struct{ Device, Interface, Description string }{Device: "dev2", Interface: "if2", Description: ""}, Role: "dc"}),
			tu.MockAlert(
				2, "Neteng DC Link Down", "dev2", "if2", "src", "scp", "2", []string{"dc", "link"},
				Circuit{ASide: struct{ Device, Interface, Description string }{Device: "dev2", Interface: "if2", Description: ""}, ZSide: struct{ Device, Interface, Description string }{Device: "dev1", Interface: "if1", Description: ""}, Role: "dc"}),
			tu.MockAlert(
				3, "Neteng DC Link Down", "dev1", "if3", "src", "scp", "3", []string{"dc", "link"},
				Circuit{ASide: struct{ Device, Interface, Description string }{Device: "dev1", Interface: "if3", Description: ""}, ZSide: struct{ Device, Interface, Description string }{Device: "dev3", Interface: "if1", Description: ""}, Role: "dc"}),
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
}

func TestGrouping(t *testing.T) {
	ctx := context.Background()
	for name, datas := range testDatas {
		grouper := groupers[name]
		for _, alert := range datas.incoming {
			grouper.addToBuf(alert)
		}
		go func() {
			grouper.doGrouping(ctx)
		}()
		var aggregated [][]int64
		for range datas.grouped {
			group := <-groupedChan
			var groupedIds []int64
			for _, a := range group.groupedAlerts {
				groupedIds = append(groupedIds, a.Id)
			}
			aggregated = append(aggregated, groupedIds)
		}
		assert.ElementsMatch(t, aggregated, datas.grouped)
	}
}
