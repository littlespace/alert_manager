package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

const queryURL = "/api/rblx/device/dm/v1/"

type NetboxInterface struct {
	Device                        string
	Interface                     string
	Description                   string
	Role                          string
	Type                          string
	Agg                           string `json:",omitempty"`
	PeerDevice, PeerIntf, PeerAgg string
}

func getResult(n *Netbox, url string) (map[string]interface{}, error) {
	body, err := n.query(url)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return n.getResult(body)
}

func (i *NetboxInterface) parse(ifaceData map[string]interface{}) error {
	if len(ifaceData) == 0 {
		return fmt.Errorf("Interface not found in result data")
	}
	if !ifaceData["is_connected"].(bool) {
		return fmt.Errorf("Link is not connected or inactive")
	}
	i.Description = ifaceData["rblx_description"].(string)
	i.Role = "dc"
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
	if ifaceData["peer_is_lag"].(bool) {
		i.PeerAgg = ifaceData["peer_lag_name"].(string)
	}
	if ifaceData["peer_role"].(string) == "border-router" {
		i.Role = "bb"
	}
	return nil
}

func (i *NetboxInterface) query(n *Netbox, alert *models.Alert) error {
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=%s", alert.Device.String, alert.Entity)
	result, err := getResult(n, url)
	if err != nil {
		return err
	}
	// add site info to alert
	site := result["site_data"].(map[string]interface{})
	alert.AddSite(site["name"].(string))

	i.Device = alert.Device.String
	i.Interface = alert.Entity
	iface := result["interfaces"].(map[string]interface{})
	ifaceData, ok := iface[alert.Entity]
	if !ok {
		return fmt.Errorf("Unable to get interface data from netbox")
	}

	return i.parse(ifaceData.(map[string]interface{}))
}

type NetboxCircuit struct {
	ASide, ZSide struct{ Device, Interface, Agg string }
	Role         string
	CktId        string `json:"circuit_id,omitempty"`
	Provider     string `json:",omitempty"`
}

func (c *NetboxCircuit) query(n *Netbox, alert *models.Alert) error {
	iface := &NetboxInterface{Device: alert.Device.String, Interface: alert.Entity}
	url := n.Addr + queryURL + fmt.Sprintf("%s?interfaces=%s", alert.Device.String, alert.Entity)
	result, err := getResult(n, url)
	if err != nil {
		return err
	}
	// add site info to alert
	site := result["site_data"].(map[string]interface{})
	alert.AddSite(site["name"].(string))

	ifc := result["interfaces"].(map[string]interface{})
	ifaceD, ok := ifc[alert.Entity]
	if !ok {
		return fmt.Errorf("Unable to get interface data from netbox")
	}
	ifaceData := ifaceD.(map[string]interface{})
	err = iface.parse(ifaceData)
	if err != nil {
		return err
	}
	c.Role = iface.Role
	c.ASide.Device = iface.Device
	c.ASide.Interface = iface.Interface
	c.ASide.Agg = iface.Agg
	c.ZSide.Device = iface.PeerDevice
	c.ZSide.Interface = iface.PeerIntf
	c.ZSide.Agg = iface.PeerAgg
	if iface.Type == "agg" {
		// we dont know a/z info for aggs
		return nil
	}
	if c.Role == "dc" {
		return nil
	}
	term := ifaceData["circuit_termination"].(map[string]interface{})
	if term["term_side"].(string) == "Z" {
		c.ZSide.Device = iface.Device
		c.ZSide.Interface = iface.Interface
		c.ZSide.Agg = iface.Agg
		c.ASide.Device = iface.PeerDevice
		c.ASide.Interface = iface.PeerIntf
		c.ASide.Agg = iface.PeerAgg
	}
	c.CktId = ifaceData["circuit_id"].(string)
	ckt := ifaceData["circuit"].(map[string]interface{})
	provider := ckt["provider"].(map[string]interface{})
	c.Provider = provider["slug"].(string)

	return nil
}
