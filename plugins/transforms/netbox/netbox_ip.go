package netbox

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
)

const queryUrlIp = "/api/ipam/ip-addresses"

type netboxIpamIpAddresses struct {
	Count   int
	Results []struct {
		Address     string
		Description string
		Status      struct {
			Label string
		}
		Role struct {
			Label string
		}
		Interface struct {
			Name        string
			Enabled     bool
			Description string
			Device      struct {
				Name string
			}
		}
	}
}

func IptoDevice(n *Netbox, ip string) (string, error) {

	url := n.Addr + queryUrlIp + fmt.Sprintf("?q=%s", ip)
	glog.Infof("query %s", url)
	body, err := n.query(url)
	if err != nil {
		return "", err
	}

	d := netboxIpamIpAddresses{}

	if err := json.Unmarshal(body, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return "", err
	}

	if d.Count == 0 {
		return "", fmt.Errorf("no match for the Ip %s, skipping", ip)
	} else if !(d.Count == 1) {
		return "", fmt.Errorf("more than 1 results for Ip %s, skipping", ip)
	}

	result := d.Results[0]

	if result.Interface.Device.Name == "" {
		return "", fmt.Errorf("no device associated with the Ip %s, skipping", ip)
	}

	return result.Interface.Device.Name, nil
}
