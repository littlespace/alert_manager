package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"net"
	"regexp"
	"strings"
)

type BgpPeer struct {
	Type            string `json:"bgp_type"`
	LocalIp         string `json:"local_ip"`
	LocalDevice     string `json:"local_device"`
	LocalInterface  string `json:"local_interface"`
	RemoteIp        string `json:"remote_ip"`
	RemoteDevice    string `json:"remote_device"`
	RemoteInterface string `json:"remote_interface"`
}

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
	return n.getResults(body), nil
}

func (p *BgpPeer) query(n *Netbox, alert *models.Alert) error {
	d := &NetboxDevice{}
	err := d.query(n, alert)
	if err != nil {
		return err
	}
	p.LocalDevice = d.Device

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
		return fmt.Errorf("Unable to extract peer-ip from alert entity")
	}

	// query the peer IP from netbox
	results, err := queryIfaceResults(n, peerIp)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("No results found for %s in netbox", peerIp)
	}
	result := results[0].(map[string]interface{})
	iface := result["interface"].(map[string]interface{})
	device := iface["device"].(map[string]interface{})
	p.RemoteIp = peerIp
	p.RemoteDevice = device["name"].(string)

	p.RemoteInterface = iface["name"].(string)
	p.RemoteInterface = strings.Replace(p.RemoteInterface, ".0", "", -1)
	if p.RemoteInterface == "lo0" {
		// the bgp session is ibgp
		p.Type = "ibgp"
		p.LocalIp = d.Ip
		p.LocalInterface = "lo0"
	} else {
		p.Type = "ebgp"
		remoteAddr := result["address"].(string)
		// find local IP
		ip, ipnet, _ := net.ParseCIDR(remoteAddr)
		if err != nil {
			return err
		}
		// find the other IP in the subnet
		for ipz := ip.Mask(ipnet.Mask); ipnet.Contains(ipz); inc(ipz) {
			if !ipz.Equal(ip) {
				p.LocalIp = ipz.String()
				break
			}
		}
		// get local interface
		results, err = queryIfaceResults(n, p.LocalIp)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("No results found for %s in netbox", p.LocalIp)
		}
		result = results[0].(map[string]interface{})
		iface := result["interface"].(map[string]interface{})
		p.LocalInterface = iface["name"].(string)
		p.LocalInterface = strings.Replace(p.LocalInterface, ".0", "", -1)
	}

	return nil
}
