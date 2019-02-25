package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"net"
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
	labels["Description"] = ifaceData["rblx_description"]
	if ifaceData["is_lag"].(bool) {
		labels["Type"] = "agg"
	} else {
		labels["Type"] = "phy"
		if ifaceData["lag"] != nil {
			lag := ifaceData["lag"].(map[string]interface{})
			labels["Agg"] = lag["name"]
		}
	}
	labels["PeerDevice"] = ifaceData["peer_name"]
	labels["PeerIntf"] = ifaceData["peer_int"]
	if labels["PeerDevice"] == nil || labels["PeerIntf"] == nil {
		labels["Role"] = "peering"
		return labels, nil
	}
	if ifaceData["peer_is_lag"].(bool) && labels["Type"] == "phy" {
		labels["PeerAgg"] = ifaceData["peer_lag_name"]
	}
	if ifaceData["peer_role"].(string) == "border-router" {
		labels["Role"] = "bb"
	} else {
		labels["Role"] = "dc"
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
	labels["LabelType"] = "Interface"
	labels["Device"] = alert.Device.String
	labels["Status"] = result["status"]
	labels["Interface"] = alert.Entity
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
	iLabels["Device"] = alert.Device.String
	iLabels["Interface"] = alert.Entity

	labels := make(models.Labels)
	labels["LabelType"] = "Circuit"

	labels["Role"] = iLabels["Role"]

	labels["ASideDeviceName"] = result["name"]
	ip, _, _ := net.ParseCIDR(result["primary_ip"].(string))
	labels["ASideDeviceIp"] = ip.String()
	labels["ASideDeviceStatus"] = result["status"]
	labels["ASideInterface"] = iLabels["Interface"]
	labels["ASideAgg"] = iLabels["Agg"]

	if iLabels["PeerDevice"] != nil {
		peerDevice := iLabels["PeerDevice"].(string)
		dLabels, err := deviceLabels(n, peerDevice)
		if err != nil {
			return nil, fmt.Errorf("Unable to query peer Device: %s, %v", peerDevice, err)
		}
		labels["ZSideDeviceName"] = dLabels["Name"]
		labels["ZSideDeviceIp"] = dLabels["Ip"]
		labels["ZSideDeviceStatus"] = dLabels["Status"]
		labels["ZSideInterface"] = iLabels["PeerIntf"]
		labels["ZSideAgg"] = iLabels["PeerAgg"]
	}

	if iLabels["Type"].(string) == "agg" {
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
	if ifaceData["circuit_termination"] == nil || ifaceData["circuit_id"] == nil {
		return labels, nil
	}
	term := ifaceData["circuit_termination"].(map[string]interface{})
	if term["term_side"].(string) == "Z" {
		labels = swap(labels, "ASideDeviceName", "ZSideDeviceName")
		labels = swap(labels, "ASideDeviceIp", "ZSideDeviceIp")
		labels = swap(labels, "ASideDeviceStatus", "ZSideDeviceStatus")
		labels = swap(labels, "ASideInterface", "ZSideInterface")
		labels = swap(labels, "ASideAgg", "ZSideAgg")
	}
	labels["CktId"] = ifaceData["circuit_id"]
	ckt := ifaceData["circuit"].(map[string]interface{})
	provider := ckt["provider"].(map[string]interface{})
	labels["Provider"] = provider["slug"]

	return labels, nil
}
