package parsers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
)

var promStatusMap = map[string]string{"firing": listener.Status_ALERTING, "resolved": listener.Status_CLEARED}

// PromAmData defines the data sent by Prometheus Alertmanager
type promAmData struct {
	Alerts       []map[string]interface{}
	CommonLabels map[string]interface{}
}

type PromParser struct {
	name string
}

func (p *PromParser) Name() string {
	return p.name
}

func (p *PromParser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := promAmData{}
	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}
	if len(d.Alerts) == 0 {
		return nil, fmt.Errorf("Invalid alert data received")
	}
	// Prom alertmanager sends a bunch of alerts in the http message.
	// There can also be a firing and a resolved for the same alert in the same message. Its stupid.
	alerts := make(map[string]*listener.WebHookAlert)
	for _, alertData := range d.Alerts {
		t, err := time.Parse(time.RFC3339, alertData["startsAt"].(string))
		if err != nil {
			glog.Errorf("Unable to parse time string , using current time")
			t = time.Now()
		}
		labels := alertData["labels"].(map[string]interface{})
		var device string
		if labels["device"] != nil {
			device = labels["device"].(string)
		}
		if labels["entity"] == nil || labels["description"] == nil {
			return nil, fmt.Errorf("Entity and description are required")
		}
		source := "prometheus"
		if labels["source"] != nil {
			source = labels["source"].(string)
		}
		sev := "info"
		if labels["severity"] != nil {
			sev = labels["severity"].(string)
		}
		alert := &listener.WebHookAlert{
			Id:      "None",
			Name:    labels["alertname"].(string),
			Details: labels["description"].(string),
			Device:  device,
			Entity:  labels["entity"].(string),
			Time:    t,
			Source:  source,
			Level:   sevToLevel[sev],
			Status:  promStatusMap[alertData["status"].(string)],
			Labels:  labels,
		}
		key := fmt.Sprintf("%s:%s:%s", alert.Name, alert.Device, alert.Entity)
		alerts[key] = alert
	}
	var uniqueAlerts []*listener.WebHookAlert
	for _, a := range alerts {
		uniqueAlerts = append(uniqueAlerts, a)
	}
	return &listener.WebHookAlertData{Alerts: uniqueAlerts}, nil
}

func init() {
	parser := &PromParser{name: "prometheus"}
	listener.AddParser(parser)
}
