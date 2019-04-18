package netbox

import (
	"fmt"
	"net"

	"github.com/mayuresh82/alert_manager/internal/models"
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
	labels["labelType"] = "Device"
	labels["name"] = result["name"]
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	labels["ip"] = ip.String()
	site := result["site_data"].(map[string]interface{})
	labels["site"] = site["name"]
	labels["region"] = result["region"]
	labels["status"] = result["status"]
	return labels, nil
}

func DeviceLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	labels, err := deviceLabels(n, alert.Device.String)
	if err != nil {
		return nil, err
	}
	alert.AddSite(labels["site"].(string))
	return labels, nil
}
