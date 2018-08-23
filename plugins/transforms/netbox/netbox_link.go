package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type NetboxInterface struct {
	Device      string
	Interface   string
	Description string
	Agg         string  `json:",omitempty"`
	TermSide    string  `json:",omitempty"`
	CktId       string  `json:",omitempty"`
	Cid         float64 `json:",omitempty"`
}

func (i *NetboxInterface) parse(raw map[string]interface{}) {
	device := raw["device"].(map[string]interface{})
	i.Device = device["name"].(string)
	i.Interface = raw["name"].(string)
	i.Description = raw["description"].(string)

	if raw["lag"] == nil {
		// ckt is not in lag, error ?!
		return
	}
	lag := raw["lag"].(map[string]interface{})
	term := raw["circuit_termination"].(map[string]interface{})
	ckt := term["circuit"].(map[string]interface{})
	i.Agg = lag["name"].(string)
	i.TermSide = term["term_side"].(string)
	i.Cid = ckt["id"].(float64)
	i.CktId = ckt["cid"].(string)
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
	ASide, ZSide *NetboxInterface
	CktId        string `json:"circuit_id"`
	Provider     string
}

func (c *NetboxCircuit) query(n *Netbox, alert *models.Alert) error {
	iface := &NetboxInterface{}
	err := iface.query(n, alert)
	if err != nil {
		return err
	}
	// pull circuit info from circuits table
	url := fmt.Sprintf("%s/api/circuits/circuits/?cid=%s", n.Addr, iface.CktId)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	results := n.getResults(body)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	raw := results[0].(map[string]interface{})
	c.CktId = iface.CktId
	provider := raw["provider"].(map[string]interface{})
	c.Provider = provider["name"].(string)

	// pull the other side from the circuit terminations table
	url = fmt.Sprintf("%s/api/circuits/circuit-terminations/?circuit_id=%v", n.Addr, iface.Cid)
	body, err = n.query(url)
	if err != nil {
		return err
	}
	var aSide, zSide *NetboxInterface
	if iface.TermSide == "A" {
		aSide = iface
		err = zSide.parseTerm(n, body, "Z")
	} else {
		zSide = iface
		err = aSide.parseTerm(n, body, "A")
	}
	if err != nil {
		return err
	}
	c.ASide = aSide
	c.ZSide = zSide
	return nil
}
