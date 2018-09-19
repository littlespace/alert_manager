package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/types"
	"net"
)

const queryUrl = "/api/rblx/device/dm/v1/"

func parseDevice(n *Netbox, device string) (*types.Device, error) {
	url := n.Addr + queryUrl + fmt.Sprintf("%s?interfaces=lo0.0", device)
	body, err := n.query(url)
	if err != nil {
		return nil, err
	}
	result, err := n.getResult(body)
	if err != nil {
		return nil, err
	}
	d := types.NewDevice()
	d.Name = result["name"].(string)
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	d.Ip = ip.String()
	site := result["site_data"].(map[string]interface{})
	d.Site = site["name"].(string)
	d.Region = result["region"].(string)
	d.Status = result["status"].(string)
	return d, nil
}

func queryDevice(n *Netbox, alert *models.Alert) (*types.Device, error) {
	device, err := parseDevice(n, alert.Device.String)
	if err != nil {
		return nil, err
	}
	alert.AddSite(device.Site)
	return device, nil
}
