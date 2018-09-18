package netbox

import (
	"bytes"
	"encoding/json"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

var testDatas = map[string][]byte{
	"/api/rblx/device/dm/v1/br1-sjc1?interfaces=lo0.0": []byte(`
  {
    "name": "br1-sjc1",
    "primary_ip": "12.8.1.1/32",
    "interfaces": {
        "lo0.0": {
            "id": 42419,
            "mtu": null,
            "mgmt_only": false,
            "description": "",
            "lag": null,
            "is_lag": false,
            "is_connected": false,
            "peer_name": null,
            "peer_role": null,
            "peer_int": null,
            "peer_is_lag": false,
            "link_is_active": false,
            "speed": null,
            "rblx_description": null
        }
    },
    "site_data": {
        "id": 8,
        "name": "sjc1",
        "region": {
            "id": 8,
            "name": "US_WEST"
        }
    },
    "region": "US_WEST"
  }`),
	"/api/rblx/device/dm/v1/bs1-ash1?interfaces=lo0.0": []byte(`
  {
    "name": "bs1-ash1",
    "primary_ip": "10.1.1.1/32",
    "interfaces": {
        "lo0.0": {
            "id": 42419,
            "mtu": null,
            "mgmt_only": false,
            "description": "",
            "lag": null,
            "is_lag": false,
            "is_connected": false,
            "peer_name": null,
            "peer_role": null,
            "peer_int": null,
            "peer_is_lag": false,
            "link_is_active": false,
            "speed": null,
            "rblx_description": null
        }
    },
    "site_data": {
        "id": 8,
        "name": "ash1",
        "region": {
            "id": 8,
            "name": "US_EAST"
        }
    },
    "region": "US_EAST"
  }`),
	"/api/rblx/device/dm/v1/bs1-ash1?interfaces=et-0/0/47": []byte(`
  {  
    "name": "bs1-ash1",
    "primary_ip": "10.1.1.1/32",
    "interfaces": {
        "et-0/0/47": {
            "id": 106171,
            "description": "",
            "lag": null,
            "is_lag": false,
            "is_connected": true,
            "peer_name": "ps11-c1-ash1",
            "peer_role": "pod-switch",
            "peer_int": "et-0/0/31",
            "peer_is_lag": false,
            "link_is_active": true,
            "speed": 40000,
            "mac_address": null,
            "interface_connection": {
                "interface": {
                    "name": "et-0/0/31",
                    "device": {
                        "id": 4945,
                        "name": "ps11-c1-ash1",
                        "role": "pod-switch"
                    }
                },
                "status": true
            },
            "rblx_description": "et-0/0/31.ps11-c1-ash1"
        }
    },
    "site_data": {
        "id": 8,
        "name": "ash1",
        "region": {
            "id": 8,
            "name": "US_EAST"
        }
    },
    "region": "US_EAST"
  }`),
	"/api/rblx/device/dm/v1/br1-sjc1?interfaces=et-0/0/3:0": []byte(`
  {
    "name": "br1-sjc1",
    "primary_ip": "12.8.1.1/32",
    "interfaces": {
        "et-0/0/3:0": {
            "id": 43649,
            "lag": {
                "name": "ae1"
            },
            "is_lag": false,
            "is_connected": true,
            "peer_name": "br1-ord1",
            "peer_role": "border-router",
            "peer_int": "et-0/0/3:0",
            "peer_is_lag": true,
            "link_is_active": true,
            "speed": 10000,
            "circuit_id": "XXXX-000062-1",
            "circuit": {
                "id": 7,
                "cid": "XXXX-000062-1",
                "install_date": "2018-01-25",
                "description": "10GB Transport",
                "type": "bb",
                "provider": {
                    "slug": "telia",
                    "asn": 1299
                }
            },
            "circuit_termination": {
                "term_side": "Z",
                "port_speed": 10000000,
                "xconnect_id": "20682284-A",
                "pp_info": "SV5:01:053635:ABC PP:0000:1101852: 46"
            },
            "interface_connection": {
                "interface": {
                    "name": "et-0/0/3:0",
                    "device": {
                        "id": 959,
                        "name": "br1-ord1",
                        "role": "border-router"
                    },
                    "circuit_connection": {
                        "term_side": "A",
                        "port_speed": 10000000,
                        "xconnect_id": "20682281",
                        "pp_info": "CH1:05:000730:ABC PP:0000:1101060:48"
                    }
                }
            },
            "peer_lag_name": "ae1",
            "rblx_description": "desc1"
        }
    },
    "site_data": {
        "id": 8,
        "name": "sjc1",
        "region": {
            "id": 8,
            "name": "US_WEST"
        }
    },
    "region": "US_WEST"
  }`),
	"/api/ipam/ip-addresses/?q=10.1.1.121": []byte(`
  {
      "count": 1,
      "next": null,
      "previous": null,
      "results": [
          {
              "id": 8119,
              "family": 4,
              "address": "10.1.1.121/31",
              "vrf": null,
              "tenant": null,
              "status": {
                  "value": 1,
                  "label": "Active"
              },
              "role": null,
              "interface": {
                  "id": 106102,
                  "device": {
                      "id": 4945,
                      "name": "ps11-c1-ash1",
                      "display_name": "ps11-c1-ash1"
                  },
                  "name": "et-0/0/31.0",
                  "enabled": true,
                  "lag": null,
                  "mtu": null,
                  "mac_address": null,
                  "mgmt_only": false,
                  "description": "",
                  "is_connected": false,
                  "interface_connection": null,
                  "circuit_termination": null
              },
              "description": "",
              "nat_inside": null,
              "nat_outside": null,
              "custom_fields": {}
          }
      ]
  }`),
	"/api/ipam/ip-addresses/?q=10.1.1.120": []byte(`
  {
      "count": 1,
      "next": null,
      "previous": null,
      "results": [
          {
              "id": 8117,
              "family": 4,
              "address": "10.1.1.120/31",
              "vrf": null,
              "tenant": null,
              "status": {
                  "value": 1,
                  "label": "Active"
              },
              "role": null,
              "interface": {
                  "id": 106278,
                  "device": {
                      "id": 5140,
                      "name": "bs1-ash1",
                      "display_name": "bs1-ash1"
                  },
                  "name": "et-0/0/47.0",
                  "enabled": true,
                  "lag": null,
                  "mtu": null,
                  "mac_address": null,
                  "mgmt_only": false,
                  "description": "",
                  "is_connected": false,
                  "interface_connection": null,
                  "circuit_termination": null
              },
              "description": "",
              "nat_inside": null,
              "nat_outside": null,
              "custom_fields": {}
          }
      ]
  }`),
	"/api/ipam/ip-addresses/?q=13.8.1.1": []byte(`
  {
      "count": 1,
      "next": null,
      "previous": null,
      "results": [
          {
              "id": 8117,
              "family": 4,
              "address": "13.8.1.1/32",
              "vrf": null,
              "tenant": null,
              "status": {
                  "value": 1,
                  "label": "Active"
              },
              "role": null,
              "interface": {
                  "id": 106278,
                  "device": {
                      "id": 5140,
                      "name": "br1-ord1",
                      "display_name": "br1-ord1"
                  },
                  "name": "lo0.0",
                  "form_factor": {
                      "value": 0,
                      "label": "Virtual"
                  },
                  "enabled": true,
                  "lag": null,
                  "mtu": null,
                  "mac_address": null,
                  "mgmt_only": false,
                  "description": "",
                  "is_connected": false,
                  "interface_connection": null,
                  "circuit_termination": null
              },
              "description": "",
              "nat_inside": null,
              "nat_outside": null,
              "custom_fields": {}
          }
      ]
  }`),
}

type mockClient struct{}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	body := testDatas[req.URL.String()]
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		StatusCode: http.StatusOK,
	}, nil
}

func TestNetboxDevice(t *testing.T) {
	a := tu.MockAlert(1, "Test", "", "br1-sjc1", "ent1", "src1", "device", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "sjc1")
	d := NetboxDevice{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &d); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, d, NetboxDevice{Device: "br1-sjc1", Ip: "12.8.1.1", Site: "sjc1", Region: "US_WEST"})
}

func TestNetboxIntf(t *testing.T) {
	a := tu.MockAlert(1, "Test", "", "bs1-ash1", "et-0/0/47", "src1", "phy_interface", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "ash1")
	d := NetboxInterface{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &d); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, d,
		NetboxInterface{
			Device:      "bs1-ash1",
			Interface:   "et-0/0/47",
			Description: "et-0/0/31.ps11-c1-ash1",
			Role:        "dc",
			Type:        "phy",
			PeerDevice:  "ps11-c1-ash1",
			PeerIntf:    "et-0/0/31",
		})
}

func TestNetboxLink(t *testing.T) {
	a := tu.MockAlert(1, "Test", "", "bs1-ash1", "et-0/0/47", "src1", "link", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "ash1")
	c := NetboxCircuit{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &c); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, c,
		NetboxCircuit{
			ASide: struct{ Device, Interface, Agg string }{Device: "bs1-ash1", Interface: "et-0/0/47"},
			ZSide: struct{ Device, Interface, Agg string }{Device: "ps11-c1-ash1", Interface: "et-0/0/31"},
			Role:  "dc",
		})

	a = tu.MockAlert(1, "Test", "", "br1-sjc1", "et-0/0/3:0", "src1", "link", "1", "WARN", []string{}, nil)
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "sjc1")
	c = NetboxCircuit{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &c); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, c,
		NetboxCircuit{
			ASide:    struct{ Device, Interface, Agg string }{Device: "br1-ord1", Interface: "et-0/0/3:0", Agg: "ae1"},
			ZSide:    struct{ Device, Interface, Agg string }{Device: "br1-sjc1", Interface: "et-0/0/3:0", Agg: "ae1"},
			Role:     "bb",
			CktId:    "XXXX-000062-1",
			Provider: "telia",
		})
}

func TestNetboxBgp(t *testing.T) {
	// ebgp peer
	a := tu.MockAlert(1, "Test", "", "bs1-ash1", "AS65101 10.1.1.121", "src1", "bgp_peer", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "ash1")
	b := BgpPeer{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &b); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, b,
		BgpPeer{
			Type:            "ebgp",
			LocalIp:         "10.1.1.120",
			LocalDevice:     "bs1-ash1",
			LocalInterface:  "et-0/0/47",
			RemoteIp:        "10.1.1.121",
			RemoteDevice:    "ps11-c1-ash1",
			RemoteInterface: "et-0/0/31",
		})

	// ibgp peer
	a = tu.MockAlert(1, "Test", "", "br1-sjc1", "AS22697 13.8.1.1", "src1", "bgp_peer", "1", "WARN", []string{}, nil)
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "sjc1")
	b = BgpPeer{}
	if err := json.Unmarshal([]byte(a.Metadata.String), &b); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, b,
		BgpPeer{
			Type:            "ibgp",
			LocalIp:         "12.8.1.1",
			LocalDevice:     "br1-sjc1",
			LocalInterface:  "lo0",
			RemoteIp:        "13.8.1.1",
			RemoteDevice:    "br1-ord1",
			RemoteInterface: "lo0",
		})
}
