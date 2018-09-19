package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/types"
)

const queryURL = "/api/rblx/device/dm/v1/"

func getResult(n *Netbox, url string) (map[string]interface{}, error) {
	body, err := n.query(url)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return n.getResult(body)
}

func parseIface(i *types.Interface, ifaceData map[string]interface{}) error {
	if len(ifaceData) == 0 {
		return fmt.Errorf("Interface not found in result data")
	}
	if !ifaceData["is_connected"].(bool) {
		return fmt.Errorf("Link is not connected or inactive")
	}
	i.Description = ifaceData["rblx_description"].(string)
	if ifaceData["is_lag"].(bool) {
		i.Type = "agg"
	} else {
		i.Type = "phy"
		if ifaceData["lag"] != nil {
			lag := ifaceData["lag"].(map[string]interface{})
			i.Agg = lag["name"].(string)
		}
	}
	i.PeerDevice = ifaceData["peer_name"].(string)
	i.PeerIntf = ifaceData["peer_int"].(string)
	if ifaceData["peer_is_lag"].(bool) && i.Type == "phy" {
		i.PeerAgg = ifaceData["peer_lag_name"].(string)
	}
	if ifaceData["peer_role"].(string) == "border-router" {
		i.Role = "bb"
	} else {
		i.Role = "dc"
	}
	return nil
}

func queryInterface(n *Netbox, alert *models.Alert) (*types.Interface, error) {
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=%s", alert.Device.String, alert.Entity)
	result, err := getResult(n, url)
	if err != nil {
		return nil, err
	}
	// add site info to alert
	site := result["site_data"].(map[string]interface{})
	alert.AddSite(site["name"].(string))

	i := types.NewInterface(alert.Device.String, alert.Entity)
	iface := result["interfaces"].(map[string]interface{})
	ifaceData, ok := iface[alert.Entity]
	if !ok {
		return nil, fmt.Errorf("Interface not found in result data")
	}

	return i, parseIface(i, ifaceData.(map[string]interface{}))
}

func queryCircuit(n *Netbox, alert *models.Alert) (*types.Circuit, error) {
	iface := types.NewInterface(alert.Device.String, alert.Entity)
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
	err = parseIface(iface, ifaceData)
	if err != nil {
		return nil, err
	}
	c := types.NewCircuit()
	c.Role = iface.Role

	d, err := parseDevice(n, iface.Device)
	if err != nil {
		return nil, fmt.Errorf("Unable to query Device: %v", err)
	}
	c.ASide.Device = d
	c.ASide.Interface = iface.Interface
	c.ASide.Agg = iface.Agg

	d, err = parseDevice(n, iface.PeerDevice)
	if err != nil {
		return nil, fmt.Errorf("Unable to query Device: %v", err)
	}
	c.ZSide.Device = d
	c.ZSide.Interface = iface.PeerIntf
	c.ZSide.Agg = iface.PeerAgg
	if iface.Type == "agg" {
		// we dont know a/z info for aggs
		return c, nil
	}
	if c.Role == "dc" {
		return c, nil
	}
	term := ifaceData["circuit_termination"].(map[string]interface{})
	if term["term_side"].(string) == "Z" {
		tmp := c.ZSide.Device
		c.ZSide.Device = c.ASide.Device
		c.ZSide.Interface = iface.Interface
		c.ZSide.Agg = iface.Agg
		c.ASide.Device = tmp
		c.ASide.Interface = iface.PeerIntf
		c.ASide.Agg = iface.PeerAgg
	}
	c.CktId = ifaceData["circuit_id"].(string)
	ckt := ifaceData["circuit"].(map[string]interface{})
	provider := ckt["provider"].(map[string]interface{})
	c.Provider = provider["slug"].(string)

	return c, nil
}
