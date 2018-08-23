package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type BgpPeer struct {
	LocalIp      string `json:"local_ip"`
	LocalDevice  string `json:"local_device"`
	RemoteIp     string `json:"remote_ip"`
	RemoteDevice string `json:"remote_device"`
}

func (p *BgpPeer) query(n *Netbox, alert *models.Alert) error {
	d := &NetboxDevice{}
	err := d.query(n, alert)
	if err != nil {
		return err
	}
	p.LocalDevice = d.Device
	p.LocalIp = d.Ip

	peerIp := alert.Entity
	url := fmt.Sprintf("%s/api/ipam/ip-addresses/?q=%s", n.Addr, peerIp)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	results := n.getResults(body)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	result := results[0].(map[string]interface{})
	iface := result["interface"].(map[string]interface{})
	device := iface["device"].(map[string]interface{})
	p.RemoteIp = peerIp
	p.RemoteDevice = device["name"].(string)
	return nil
}
