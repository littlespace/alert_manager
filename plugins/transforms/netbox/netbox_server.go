package netbox

import (
	"fmt"

	"github.com/mayuresh82/alert_manager/internal/models"
)

func ServerLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	url := n.Addr + fmt.Sprintf("/api/dcim/devices/?name=%s", alert.Device.String)
	body, err := n.query(url)
	if err != nil {
		return nil, err
	}
	results, err := n.getResults(body)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("Cannot find server %s in netbox", alert.Device.String)
	}
	result := results[0].(map[string]interface{})
	labels := make(models.Labels)
	labels["labelType"] = "Device"
	labels["name"] = result["name"]
	labels["assetTag"] = result["asset_tag"]
	labels["serial"] = result["serial"]
	site := result["site"].(map[string]interface{})
	labels["site"] = site["name"]
	dType := result["device_type"].(map[string]interface{})
	labels["model"] = dType["slug"]
	manu := dType["manufacturer"].(map[string]interface{})
	labels["manufacturer"] = manu["slug"]

	if site, ok := alert.Labels["site"]; ok {
		alert.AddSite(site.(string))
	} else {
		alert.AddSite(labels["site"].(string))
	}
	return labels, nil
}
