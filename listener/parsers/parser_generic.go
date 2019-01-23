package parsers

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
	"time"
)

var sevToLevel = map[string]string{"info": "INFO", "warning": "WARN", "critical": "CRITICAL"}
var statusToAlertStatus = map[string]string{"alerting": listener.Status_ALERTING, "recover": listener.Status_CLEARED}

// alertData corresponds to custom json payload sent by a configurable endpoint
type alertData struct {
	Id          float64                // id from the external system (optional)
	Name        string                 // Name of the alert (required)
	Timestamp   string                 // string timestamp in RFC3339 format (optional)
	Severity    string                 // either "info", "warning", or "critical" (optional)
	Device      string                 // alerting device (optional)
	Entity      string                 // alerting entity (required)
	Description string                 // descriptive msg (required)
	Preamble    string                 // preamble to prepend to the message (optional)
	Status      string                 // either "alerting" or "recover" (required)
	Source      string                 // source of the alert (optional)
	Labels      map[string]interface{} // map of labels to attach to alert (optional)
}

type GenericParser struct {
	name string
}

func (p *GenericParser) Name() string {
	return p.name
}

func (p *GenericParser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := alertData{}
	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, d.Timestamp)
	if err != nil {
		glog.Errorf("Unable to parse time string , using current time")
		t = time.Now()
	}
	var missing []string
	switch {
	case d.Name == "":
		missing = append(missing, "name")
	case d.Entity == "":
		missing = append(missing, "entity")
	case d.Description == "":
		missing = append(missing, "description")
	case d.Status == "":
		missing = append(missing, "status")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("Required fields missing: %v", missing)
	}
	details := d.Description
	if d.Preamble != "" {
		details = d.Preamble + ":" + d.Description
	}
	return &listener.WebHookAlertData{
		Id:      fmt.Sprintf("%d", int64(d.Id)),
		Name:    d.Name,
		Details: details,
		Device:  d.Device,
		Entity:  d.Entity,
		Time:    t,
		Source:  d.Source,
		Level:   sevToLevel[d.Severity],
		Status:  statusToAlertStatus[d.Status],
		Labels:  d.Labels,
	}, nil
}

func init() {
	parser := &GenericParser{name: "generic"}
	listener.AddParser(parser)
}
