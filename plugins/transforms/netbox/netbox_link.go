package netbox

import (
	"fmt"
	"net"

	"github.com/mayuresh82/alert_manager/internal/models"
)

const queryURL = "/api/rblx/device/dm/v1/"

// swap swaps two keys in a map
func swap(in map[string]interface{}, first, second string) map[string]interface{} {
	tmp := in[first]
	in[first] = in[second]
	in[second] = tmp
	return in
}

func getResult(n *Netbox, url string) (map[string]interface{}, error) {
	body, err := n.query(url)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return n.getResult(body)
}

func ifaceLabels(ifaceData map[string]interface{}) (models.Labels, error) {
	if len(ifaceData) == 0 {
		return nil, fmt.Errorf("Interface not found in result data")
	}
	if !ifaceData["is_connected"].(bool) {
		return nil, fmt.Errorf("Link is not connected or inactive")
	}
	labels := make(models.Labels)
	labels["description"] = ifaceData["rblx_description"]
	if ifaceData["is_lag"].(bool) {
		labels["type"] = "agg"
	} else {
		labels["type"] = "phy"
		if ifaceData["lag"] != nil {
			lag := ifaceData["lag"].(map[string]interface{})
			labels["agg"] = lag["name"]
		}
	}
	labels["peerDevice"] = ifaceData["peer_name"]
	labels["peerIntf"] = ifaceData["peer_int"]
	if labels["peerDevice"] == nil || labels["peerIntf"] == nil {
		labels["role"] = "peering"
		return labels, nil
	}
	if ifaceData["peer_is_lag"].(bool) && labels["Type"] == "phy" {
		labels["peerAgg"] = ifaceData["peer_lag_name"]
	}
	if ifaceData["peer_role"].(string) == "border-router" {
		labels["role"] = "bb"
	} else {
		labels["role"] = "dc"
	}
	return labels, nil
}

func InterfaceLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=%s", alert.Device.String, alert.Entity)
	result, err := getResult(n, url)
	if err != nil {
		return nil, err
	}
	// add site info to alert
	site := result["site_data"].(map[string]interface{})
	alert.AddSite(site["name"].(string))

	iface := result["interfaces"].(map[string]interface{})
	ifaceData, ok := iface[alert.Entity]
	if !ok {
		return nil, fmt.Errorf("Interface not found in result data")
	}
	labels, err := ifaceLabels(ifaceData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	labels["labelType"] = "Interface"
	labels["device"] = alert.Device.String
	labels["status"] = result["status"]
	labels["interface"] = alert.Entity
	return labels, nil
}

func CircuitLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=%s", alert.Device.String, alert.Entity)
	result, err := getResult(n, url)
	if err != nil {
		return nil, err
	}
	// add site info to alert
	site := result["site_data"].(map[string]interface{})
	alert.AddSite(site["name"].(string))

	ifc := result["interfaces"].(map[string]interface{})
	ifaceD, ok := ifc[alert.Entity]
	if !ok {
		return nil, fmt.Errorf("Interface not found in result data")
	}
	ifaceData := ifaceD.(map[string]interface{})
	iLabels, err := ifaceLabels(ifaceData)
	if err != nil {
		return nil, err
	}
	iLabels["device"] = alert.Device.String
	iLabels["interface"] = alert.Entity

	labels := make(models.Labels)
	labels["labelType"] = "Circuit"

	labels["role"] = iLabels["role"]

	labels["aSideDeviceName"] = result["name"]
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	labels["aSideDeviceIp"] = ip.String()
	labels["aSideDeviceStatus"] = result["status"]
	labels["aSideInterface"] = iLabels["interface"]
	labels["aSideAgg"] = iLabels["agg"]

	if iLabels["peerDevice"] != nil {
		peerDevice := iLabels["peerDevice"].(string)
		dLabels, err := deviceLabels(n, peerDevice)
		if err != nil {
			return nil, fmt.Errorf("Unable to query peer Device: %s, %v", peerDevice, err)
		}
		labels["zSideDeviceName"] = dLabels["name"]
		labels["zSideDeviceIp"] = dLabels["ip"]
		labels["zSideDeviceStatus"] = dLabels["status"]
		labels["zSideInterface"] = iLabels["peerIntf"]
		labels["zSideAgg"] = iLabels["peerAgg"]
	}

	if labels["role"].(string) == "dc" {
		return labels, nil
	}

	if iLabels["type"].(string) == "agg" {
		// pull a/z, provider info from children
		children := ifaceData["childs"].(map[string]interface{})
		for c, v := range children {
			v := v.(map[string]interface{})
			if !v["is_connected"].(bool) {
				continue
			}
			ifaceD, ok := ifc[c]
			if !ok {
				return nil, fmt.Errorf("Child Interface %s not found in result data", c)
			}
			ifaceData = ifaceD.(map[string]interface{})
			break
		}
	}
	term := ifaceData["circuit_termination"].(map[string]interface{})
	if term["term_side"].(string) == "Z" {
		labels = swap(labels, "aSideDeviceName", "zSideDeviceName")
		labels = swap(labels, "aSideDeviceIp", "zSideDeviceIp")
		labels = swap(labels, "aSideDeviceStatus", "zSideDeviceStatus")
		labels = swap(labels, "aSideInterface", "zSideInterface")
		labels = swap(labels, "aSideAgg", "zSideAgg")
	}
	labels["cktId"] = ifaceData["circuit_id"]
	ckt := ifaceData["circuit"].(map[string]interface{})
	if ckt["status"] != nil {
		status := ckt["status"].(map[string]interface{})
		labels["status"] = status["label"]
	}
	provider := ckt["provider"].(map[string]interface{})
	labels["provider"] = provider["slug"]

	return labels, nil
}
