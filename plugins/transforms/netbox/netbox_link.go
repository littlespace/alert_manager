package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type NetboxInterface struct {
	Device      string
	Interface   string
	Description string
	Role        string
	Type        string
	Agg         string `json:",omitempty"`
	Term        struct {
		TermSide string  `json:",omitempty"`
		CktId    string  `json:",omitempty"`
		Cid      float64 `json:",omitempty"`
	}
	Conn struct {
		Device      string `json:",omitempty"`
		Interface   string `json:",omitempty"`
		Description string `json:",omitempty"`
	}
}

func (i *NetboxInterface) parse(raw map[string]interface{}) {
	device := raw["device"].(map[string]interface{})
	i.Device = device["name"].(string)
	i.Interface = raw["name"].(string)
	i.Description = raw["description"].(string)
	i.Type = "phy"

	if raw["circuit_termination"] == nil {
		// check if ckt is a DC ckt
		if raw["interface_connection"] == nil {
			// check if iface is a lag
			ff := raw["form_factor"].(map[string]interface{})
			if ff["value"].(float64) == 200 {
				i.Type = "lag"
				return
			}
		}
		i.Role = "dc"
		conn := raw["interface_connection"].(map[string]interface{})
		iface := conn["interface"].(map[string]interface{})
		i.Conn.Interface = iface["name"].(string)
		i.Conn.Description = iface["description"].(string)
		dev := iface["device"].(map[string]interface{})
		i.Conn.Device = dev["name"].(string)
		return
	}
	if raw["lag"] == nil {
		// ckt is not in lag, error ?!
		return
	}
	i.Role = "bb"
	lag := raw["lag"].(map[string]interface{})
	term := raw["circuit_termination"].(map[string]interface{})
	ckt := term["circuit"].(map[string]interface{})
	i.Agg = lag["name"].(string)
	i.Term.TermSide = term["term_side"].(string)
	i.Term.Cid = ckt["id"].(float64)
	i.Term.CktId = ckt["cid"].(string)
}

func (i *NetboxInterface) parseTerm(n *Netbox, data []byte, termSide string) error {
	results := n.getResults(data)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	for _, result := range results {
		res := result.(map[string]interface{})
		if res["term_side"].(string) == termSide {
			i.parse(res["interface"].(map[string]interface{}))
			return nil
		}
	}
	return fmt.Errorf("Circuit termination not found: %s", termSide)
}

func (i *NetboxInterface) query(n *Netbox, alert *models.Alert) error {
	url := fmt.Sprintf("%s/api/dcim/interfaces/?name=%s&device=%s", n.Addr, alert.Entity, alert.Device.String)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	results := n.getResults(body)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	raw := results[0].(map[string]interface{})
	i.parse(raw)
	return nil
}

type NetboxCircuit struct {
	ASide, ZSide struct{ Device, Interface, Agg string }
	Role         string
	CktId        string `json:"circuit_id,omitempty"`
	Provider     string `json:",omitempty"`
}

func (c *NetboxCircuit) query(n *Netbox, alert *models.Alert) error {
	iface := &NetboxInterface{}
	err := iface.query(n, alert)
	if err != nil {
		return err
	}
	if iface.Role == "" {
		if iface.Type == "lag" {
			// lag has no endpoint info. Just save it as the ASide
			c.ASide.Device = iface.Device
			c.ASide.Interface = iface.Interface
			return nil
		}
		return fmt.Errorf("Unable to get circuit role")
	}
	c.Role = iface.Role
	if c.Role == "dc" {
		c.ASide.Device = iface.Device
		c.ASide.Interface = iface.Interface
		c.ZSide.Device = iface.Conn.Device
		c.ZSide.Interface = iface.Conn.Interface
		return nil
	}

	// pull circuit info from circuits table
	url := fmt.Sprintf("%s/api/circuits/circuits/?cid=%s", n.Addr, iface.Term.CktId)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	results := n.getResults(body)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	raw := results[0].(map[string]interface{})
	c.CktId = iface.Term.CktId
	provider := raw["provider"].(map[string]interface{})
	c.Provider = provider["name"].(string)

	// pull the other side from the circuit terminations table
	url = fmt.Sprintf("%s/api/circuits/circuit-terminations/?circuit_id=%v", n.Addr, iface.Term.Cid)
	body, err = n.query(url)
	if err != nil {
		return err
	}
	var aSide, zSide *NetboxInterface
	if iface.Term.TermSide == "A" {
		aSide = iface
		err = zSide.parseTerm(n, body, "Z")
	} else {
		zSide = iface
		err = aSide.parseTerm(n, body, "A")
	}
	if err != nil {
		return err
	}
	c.ASide.Device = aSide.Device
	c.ASide.Interface = aSide.Interface
	c.ASide.Agg = aSide.Agg
	c.ZSide.Device = zSide.Device
	c.ZSide.Interface = zSide.Interface
	c.ZSide.Agg = zSide.Agg
	return nil
}
