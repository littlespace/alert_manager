package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"net"
)

const queryUrl = "/api/rblx/device/dm/v1/"

type NetboxDevice struct {
	Device string
	Ip     string
	Site   string
	Region string `json:",omitempty"`
}

func (d *NetboxDevice) query(n *Netbox, alert *models.Alert) error {
	url := n.Addr + queryUrl + fmt.Sprintf("%s?interfaces=lo0.0", alert.Device.String)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	result, err := n.getResult(body)
	if err != nil {
		return err
	}
	d.Device = result["name"].(string)
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	d.Ip = ip.String()
	site := result["site_data"].(map[string]interface{})
	d.Site = site["name"].(string)
	d.Region = result["region"].(string)

	alert.AddSite(d.Site)
	return nil
}
