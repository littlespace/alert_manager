package netbox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"io/ioutil"
	"net/http"
	"time"
)

type Clienter interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	*http.Client
}

// Netbox transform pulls info from netbox and applies it to alert labels and other fields
type Netbox struct {
	Addr, Token string
	Priority    int
	Register    string
	client      Clienter
	err         error // stores any error encountered while applying the transform
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

func (n *Netbox) getResults(data []byte) ([]interface{}, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return []interface{}{}, fmt.Errorf("Unable to unmarshal data: %v", err)
	}
	return d["results"].([]interface{}), nil
}

func (n *Netbox) getResult(data []byte) (map[string]interface{}, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return d, fmt.Errorf("Unable to unmarshal data: %v", err)
	}
	return d, nil
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unable to query netbox, Got %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (n *Netbox) apply(alert *models.Alert) {
	if !alert.Device.Valid {
		n.err = fmt.Errorf("Unable to get device from alert: field empty !")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			n.err = fmt.Errorf("PANIC while applying netbox transform: %v", r)
		}
	}()
	var l models.Labels
	switch alert.Scope {
	case "device":
		l, n.err = DeviceLabels(n, alert)
	case "phy_interface", "agg_interface":
		l, n.err = InterfaceLabels(n, alert)
	case "link":
		l, n.err = CircuitLabels(n, alert)
	case "bgp_peer":
		l, n.err = BgpLabels(n, alert)
	default:
		n.err = fmt.Errorf("Scope %s is not defined in netbox", alert.Scope)
	}
	if n.err != nil {
		return
	}
	alert.Labels = l
}

func (n *Netbox) Apply(alert *models.Alert) error {
	n.apply(alert)
	lastErr := n.err
	n.err = nil
	return lastErr
}

func init() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	n := &Netbox{client: &Client{&http.Client{Transport: tr, Timeout: 5 * time.Second}}}
	ah.AddTransform(n)
}
