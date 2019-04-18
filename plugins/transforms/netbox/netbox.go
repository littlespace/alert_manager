package netbox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
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

func (n *Netbox) Apply(alert *models.Alert) error {
	var l models.Labels
	scope := alert.Scope
	if scope == "" {
		if scp, ok := alert.Labels["scope"]; ok {
			scope = scp.(string)
		}
	}
	var err error
	switch scope {
	case "device":
		l, err = DeviceLabels(n, alert)
	case "phy_interface", "agg_interface":
		l, err = InterfaceLabels(n, alert)
	case "link":
		l, err = CircuitLabels(n, alert)
	case "bgp_peer":
		l, err = BgpLabels(n, alert)
	case "dns_monitor":
		if val, ok := alert.Labels["vipIp"]; ok {
			deviceName, er := IptoDevice(n, val.(string))
			if er == nil {
				alert.AddDevice(deviceName)
			}
		}
		l, err = DeviceLabels(n, alert)

	default:
		glog.V(2).Infof("Not applying transform: Scope %s is not defined in netbox", alert.Scope)
		return nil

	}
	if err != nil {
		return err
	}
	for k, v := range l {
		alert.Labels[k] = v
	}
	return nil
}

func init() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	n := &Netbox{client: &Client{&http.Client{Transport: tr, Timeout: 5 * time.Second}}}
	ah.AddTransform(n)
}
