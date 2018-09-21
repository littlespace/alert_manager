package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"net"
)

const queryUrl = "/api/rblx/device/dm/v1/"

func deviceLabels(n *Netbox, device string) (models.Labels, error) {
	url := n.Addr + queryUrl + fmt.Sprintf("%s?interfaces=lo0.0", device)
	body, err := n.query(url)
	if err != nil {
		return nil, err
	}
	result, err := n.getResult(body)
	if err != nil {
		return nil, err
	}
	labels := make(models.Labels)
	labels["LabelType"] = "Device"
	labels["Name"] = result["name"]
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	labels["Ip"] = ip.String()
	site := result["site_data"].(map[string]interface{})
	labels["Site"] = site["name"]
	labels["Region"] = result["region"]
	labels["Status"] = result["status"]
	return labels, nil
}

func DeviceLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	labels, err := deviceLabels(n, alert.Device.String)
	if err != nil {
		return nil, err
	}
	alert.AddSite(labels["Site"].(string))
	return labels, nil
}
