package netbox

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/mayuresh82/alert_manager/internal/models"
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

func BgpLabels(n *Netbox, alert *models.Alert, addSite bool) (models.Labels, error) {
	dLabels, err := DeviceLabels(n, alert, addSite)
	if err != nil {
		return nil, err
	}
	labels := make(models.Labels)
	labels["labelType"] = "Bgp"
	labels["localDeviceName"] = dLabels["name"]
	labels["localDeviceIp"] = dLabels["ip"]
	labels["localDeviceStatus"] = dLabels["status"]

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
	labels["remoteIp"] = peerIp
	dLabels, err = deviceLabels(n, device["name"].(string))
	if err != nil {
		return nil, fmt.Errorf("Unable to query remote device: %v", err)
	}
	labels["remoteDeviceName"] = dLabels["name"]
	labels["remoteDeviceIp"] = dLabels["ip"]
	labels["remoteDeviceStatus"] = dLabels["status"]

	labels["remoteInterface"] = strings.Replace(iface["name"].(string), ".0", "", -1)
	if labels["remoteInterface"] == "lo0" {
		// the bgp session is ibgp
		labels["type"] = "ibgp"
		labels["localIp"] = labels["localDeviceIp"]
		labels["localInterface"] = "lo0"
	} else {
		labels["type"] = "ebgp"
		remoteAddr := result["address"].(string)
		// find local IP
		ip, ipnet, _ := net.ParseCIDR(remoteAddr)
		if err != nil {
			return nil, err
		}
		// find the other IP in the subnet
		for ipz := ip.Mask(ipnet.Mask); ipnet.Contains(ipz); inc(ipz) {
			if !ipz.Equal(ip) {
				labels["localIp"] = ipz.String()
				break
			}
		}
		// get local interface
		results, err = queryIfaceResults(n, labels["localIp"].(string))
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("No results found for %s in netbox", labels["localIp"])
		}
		result = results[0].(map[string]interface{})
		iface := result["interface"].(map[string]interface{})
		labels["localInterface"] = strings.Replace(iface["name"].(string), ".0", "", -1)
	}

	return labels, nil
}
