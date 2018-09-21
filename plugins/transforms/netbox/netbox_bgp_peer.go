package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"net"
	"regexp"
	"strings"
)

// helper func that increments a net.IP
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func queryIfaceResults(n *Netbox, ip string) ([]interface{}, error) {
	url := fmt.Sprintf("%s/api/ipam/ip-addresses/?q=%s", n.Addr, ip)
	body, err := n.query(url)
	if err != nil {
		return nil, err
	}
	return n.getResults(body)
}

func BgpLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	dLabels, err := DeviceLabels(n, alert)
	if err != nil {
		return nil, err
	}
	labels := make(models.Labels)
	labels["LabelType"] = "Bgp"
	labels["LocalDeviceName"] = dLabels["Name"]
	labels["LocalDeviceIp"] = dLabels["Ip"]
	labels["LocalDeviceStatus"] = dLabels["Status"]

	// extract peer IP from entity
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock
	regEx := regexp.MustCompile(regexPattern)
	peerIp := regEx.FindString(alert.Entity)
	if peerIp == "" {
		// check for v6 : matches specific v6 addrs a:b:c..::x
		regexPattern = `((\w+)\:?)+\:\:\w+`
		regEx = regexp.MustCompile(regexPattern)
		peerIp = regEx.FindString(alert.Entity)
	}
	if peerIp == "" {
		return nil, fmt.Errorf("Unable to extract peer-ip from alert entity")
	}

	// query the peer IP from netbox
	results, err := queryIfaceResults(n, peerIp)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("No results found for %s in netbox", peerIp)
	}
	result := results[0].(map[string]interface{})
	iface := result["interface"].(map[string]interface{})
	device := iface["device"].(map[string]interface{})
	labels["RemoteIp"] = peerIp
	dLabels, err = deviceLabels(n, device["name"].(string))
	if err != nil {
		return nil, fmt.Errorf("Unable to query remote device: %v", err)
	}
	labels["RemoteDeviceName"] = dLabels["Name"]
	labels["RemoteDeviceIp"] = dLabels["Ip"]
	labels["RemoteDeviceStatus"] = dLabels["Status"]

	labels["RemoteInterface"] = strings.Replace(iface["name"].(string), ".0", "", -1)
	if labels["RemoteInterface"] == "lo0" {
		// the bgp session is ibgp
		labels["Type"] = "ibgp"
		labels["LocalIp"] = labels["LocalDeviceIp"]
		labels["LocalInterface"] = "lo0"
	} else {
		labels["Type"] = "ebgp"
		remoteAddr := result["address"].(string)
		// find local IP
		ip, ipnet, _ := net.ParseCIDR(remoteAddr)
		if err != nil {
			return nil, err
		}
		// find the other IP in the subnet
		for ipz := ip.Mask(ipnet.Mask); ipnet.Contains(ipz); inc(ipz) {
			if !ipz.Equal(ip) {
				labels["LocalIp"] = ipz.String()
				break
			}
		}
		// get local interface
		results, err = queryIfaceResults(n, labels["LocalIp"].(string))
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("No results found for %s in netbox", labels["LocalIp"])
		}
		result = results[0].(map[string]interface{})
		iface := result["interface"].(map[string]interface{})
		labels["LocalInterface"] = strings.Replace(iface["name"].(string), ".0", "", -1)
	}

	return labels, nil
}
