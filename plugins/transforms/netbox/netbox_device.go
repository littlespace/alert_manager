package netbox

import (
	"fmt"
	"github.com/mayuresh82/alert_manager/internal/models"
)

type NetboxDevice struct {
	Device string
	Ip     string
	Site   string
	Region string `json:",omitempty"`
}

func (d *NetboxDevice) query(n *Netbox, alert *models.Alert) error {
	url := fmt.Sprintf("%s/api/dcim/devices/?name=%s", n.Addr, alert.Device.String)
	body, err := n.query(url)
	if err != nil {
		return err
	}
	results := n.getResults(body)
	if len(results) == 0 {
		return fmt.Errorf("No results found in netbox")
	}
	result := results[0].(map[string]interface{})
	site := result["site"].(map[string]interface{})
	ip := result["primary_ip"].(map[string]interface{})
	// TODO get region and other
	d.Device = result["name"].(string)
	d.Ip = ip["address"].(string)
	d.Site = site["name"].(string)
	return nil
}
