package netbox

import (
	"fmt"
	"net"

	"github.com/mayuresh82/alert_manager/internal/models"
)

func deviceLabels(n *Netbox, device string) (models.Labels, error) {
	queryURL, ok := n.Options["rblx_dm_url"]
	if !ok {
		return nil, fmt.Errorf("Cant find Query URL in options")
	}
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=lo0.0", device)
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
	if primaryIp, ok := result["primary_ip"]; ok {
		ip, _, _ := net.ParseCIDR(primaryIp.(string))
		labels["ip"] = ip.String()
	}
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
	if site, ok := alert.Labels["site"]; ok {
		alert.AddSite(site.(string))
	} else {
		alert.AddSite(labels["site"].(string))
	}
	return labels, nil
}
