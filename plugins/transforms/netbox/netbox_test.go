package netbox

import (
	"bytes"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

var testDatas = map[string][]byte{
	"/api/rblx/device/dm/v1/dev1-bb1?interfaces=lo0.0": []byte(`
  {
    "name": "dev1-bb1",
    "primary_ip": "12.8.1.1/32",
    "status": "Active",
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
        "name": "bb1",
        "region": {
            "id": 8,
            "name": "US_WEST"
        }
    },
    "region": "US_WEST"
  }`),
	"/api/rblx/device/dm/v1/dev2-bb1?interfaces=lo0.0": []byte(`
  {
    "name": "dev2-bb1",
    "primary_ip": "13.8.1.1/32",
    "status": "Active",
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
        "name": "bb1",
        "region": {
            "id": 8,
            "name": "US_WEST"
        }
    },
    "region": "US_WEST"
  }`),
	"/api/rblx/device/dm/v1/dev1-dc1?interfaces=lo0.0": []byte(`
  {
    "name": "dev1-dc1",
    "primary_ip": "10.1.1.1/32",
    "status": "Active",
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
        "name": "dc1",
        "region": {
            "id": 8,
            "name": "US_EAST"
        }
    },
    "region": "US_EAST"
  }`),
	"/api/rblx/device/dm/v1/dev2-dc1?interfaces=lo0.0": []byte(`
  {
    "name": "dev2-dc1",
    "primary_ip": "10.1.1.2/32",
    "status": "Active",
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
        "name": "dc1",
        "region": {
            "id": 8,
            "name": "US_EAST"
        }
    },
    "region": "US_EAST"
  }`),
	"/api/rblx/device/dm/v1/dev1-dc1?interfaces=et-0/0/47": []byte(`
  {  
    "name": "dev1-dc1",
    "primary_ip": "10.1.1.1/32",
    "status": "Active",
    "interfaces": {
        "et-0/0/47": {
            "id": 106171,
            "description": "",
            "lag": null,
            "is_lag": false,
            "is_connected": true,
            "peer_name": "dev2-dc1",
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
                        "name": "dev2-dc1",
                        "role": "pod-switch"
                    }
                },
                "status": true
            },
            "rblx_description": "et-0/0/31.dev2-dc1"
        }
    },
    "site_data": {
        "id": 8,
        "name": "dc1",
        "region": {
            "id": 8,
            "name": "US_EAST"
        }
    },
    "region": "US_EAST"
  }`),
	"/api/rblx/device/dm/v1/dev1-bb1?interfaces=et-0/0/3:0": []byte(`
  {
    "name": "dev1-bb1",
    "primary_ip": "12.8.1.1/32",
    "status": "Active",
    "interfaces": {
        "et-0/0/3:0": {
            "id": 43649,
            "lag": {
                "name": "ae1"
            },
            "is_lag": false,
            "is_connected": true,
            "peer_name": "dev2-bb1",
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
                        "name": "dev2-bb1",
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
        "name": "bb1",
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
                      "name": "dev2-dc1",
                      "display_name": "dev2-dc1"
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
                      "name": "dev1-dc1",
                      "display_name": "dev1-dc1"
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
                      "name": "dev2-bb1",
                      "display_name": "dev2-bb1"
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
	a := tu.MockAlert(1, "Test", "", "dev1-bb1", "ent1", "src1", "device", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "bb1")
	exp := models.Labels{"LabelType": "Device", "Name": "dev1-bb1", "Ip": "12.8.1.1", "Site": "bb1", "Region": "US_WEST", "Status": "Active"}
	assert.Equalf(t, a.Labels.Equal(exp), true, "Expected: %v, Got: %v", exp, a.Labels)
}

func TestNetboxIntf(t *testing.T) {
	a := tu.MockAlert(1, "Test", "", "dev1-dc1", "et-0/0/47", "src1", "phy_interface", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "dc1")
	exp := models.Labels{
		"LabelType":   "Interface",
		"Device":      "dev1-dc1",
		"Status":      "Active",
		"Interface":   "et-0/0/47",
		"Description": "et-0/0/31.dev2-dc1",
		"Role":        "dc",
		"Type":        "phy",
		"PeerDevice":  "dev2-dc1",
		"PeerIntf":    "et-0/0/31",
	}
	assert.Equalf(t, a.Labels.Equal(exp), true, "Expected: %v, Got: %v", exp, a.Labels)
}

func TestNetboxLink(t *testing.T) {
	a := tu.MockAlert(1, "Test", "", "dev1-dc1", "et-0/0/47", "src1", "link", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "dc1")
	assert.Equal(t, a.Labels["ASideDeviceName"], "dev1-dc1")
	assert.Equal(t, a.Labels["ASideInterface"], "et-0/0/47")
	assert.Equal(t, a.Labels["ZSideDeviceName"], "dev2-dc1")
	assert.Equal(t, a.Labels["ZSideInterface"], "et-0/0/31")

	a = tu.MockAlert(1, "Test", "", "dev1-bb1", "et-0/0/3:0", "src1", "link", "1", "WARN", []string{}, nil)
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "bb1")
	assert.Equal(t, a.Labels["ASideDeviceName"], "dev2-bb1")
	assert.Equal(t, a.Labels["ASideInterface"], "et-0/0/3:0")
	assert.Equal(t, a.Labels["ZSideDeviceName"], "dev1-bb1")
	assert.Equal(t, a.Labels["ZSideInterface"], "et-0/0/3:0")
	assert.Equal(t, a.Labels["Role"], "bb")
	assert.Equal(t, a.Labels["CktId"], "XXXX-000062-1")
	assert.Equal(t, a.Labels["Provider"], "telia")
}

func TestNetboxBgp(t *testing.T) {
	// ebgp peer
	a := tu.MockAlert(1, "Test", "", "dev1-dc1", "AS65101 10.1.1.121", "src1", "bgp_peer", "1", "WARN", []string{}, nil)
	n := &Netbox{client: &mockClient{}}
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "dc1")
	exp := models.Labels{
		"LabelType":          "Bgp",
		"Type":               "ebgp",
		"LocalIp":            "10.1.1.120",
		"LocalDeviceName":    "dev1-dc1",
		"LocalDeviceIp":      "10.1.1.1",
		"LocalInterface":     "et-0/0/47",
		"LocalDeviceStatus":  "Active",
		"RemoteIp":           "10.1.1.121",
		"RemoteDeviceName":   "dev2-dc1",
		"RemoteInterface":    "et-0/0/31",
		"RemoteDeviceIp":     "10.1.1.2",
		"RemoteDeviceStatus": "Active",
	}
	assert.Equalf(t, a.Labels.Equal(exp), true, "Expected: %v, Got: %v", exp, a.Labels)

	// ibgp peer
	a = tu.MockAlert(1, "Test", "", "dev1-bb1", "AS22697 13.8.1.1", "src1", "bgp_peer", "1", "WARN", []string{}, nil)
	if err := n.Apply(a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site.String, "bb1")
	exp = models.Labels{
		"LabelType":          "Bgp",
		"Type":               "ibgp",
		"LocalIp":            "12.8.1.1",
		"LocalDeviceIp":      "12.8.1.1",
		"LocalDeviceName":    "dev1-bb1",
		"LocalDeviceStatus":  "Active",
		"LocalInterface":     "lo0",
		"RemoteIp":           "13.8.1.1",
		"RemoteDeviceName":   "dev2-bb1",
		"RemoteDeviceIp":     "13.8.1.1",
		"RemoteInterface":    "lo0",
		"RemoteDeviceStatus": "Active",
	}
	assert.Equalf(t, a.Labels.Equal(exp), true, "Expected: %v, Got: %v", exp, a.Labels)
}
