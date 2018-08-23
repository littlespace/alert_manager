package netbox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"io/ioutil"
	"net/http"
)

type meta interface {
	query(n *Netbox, alert *models.Alert) error
}

type Client struct {
	*http.Client
}

type Netbox struct {
	Addr, Token string
	Priority    int
	Register    string
	client      *Client
}

func (n *Netbox) Name() string {
	return "netbox"
}

func (n *Netbox) GetPriority() int {
	return n.Priority
}

func (n *Netbox) GetRegister() string {
	return n.Register
}

func (n *Netbox) getResults(data []byte) []interface{} {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return []interface{}{}
	}
	return d["results"].([]interface{})
}

func (n *Netbox) query(query string) ([]byte, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", n.Token)
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (n *Netbox) Apply(alert *models.Alert) error {
	if !alert.Device.Valid {
		return fmt.Errorf("Unable to get device from alert: field empty !")
	}
	var m meta
	switch alert.Scope {
	case "device":
		m = &NetboxDevice{}
	case "phy_interface", "agg_interface":
		m = &NetboxInterface{}
	case "link":
		m = &NetboxCircuit{}
	case "bgp_peer":
		m = &BgpPeer{}
	default:
		return fmt.Errorf("Scope %s is not defined in netbox", alert.Scope)
	}

	err := m.query(n, alert)
	if err != nil {
		return err
	}
	metaData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	alert.AddMeta(string(metaData))
	return nil
}

func init() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	n := &Netbox{client: &Client{&http.Client{Transport: tr}}}
	ah.AddTransform(n)
}
