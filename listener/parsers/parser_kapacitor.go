package parsers

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
	"time"
)

type kapacitorData struct {
	Id            string // alert name encoded as an ID
	Message       string
	Details       string // Info message of the alert
	Time          string // time triggered
	Level         string // INFO, WARN, CRITICAL, OK
	PreviousLevel string
	Data          map[string]interface{} // time series data
}

type KapacitorParser struct {
	name string
}

func (p *KapacitorParser) Name() string {
	return p.name
}

func (p *KapacitorParser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := kapacitorData{}
	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}
	details := d.Message + "\n" + d.Details
	t, err := time.Parse(time.RFC3339, d.Time)
	if err != nil {
		glog.Errorf("Unable to parse time string , using current time")
		t = time.Now()
	}
	status := listener.Status_ALERTING
	if d.Level == "WARNING" {
		d.Level = "WARN"
	}
	if d.Level == "OK" {
		status = listener.Status_CLEARED
	}
	r := &listener.WebHookAlertData{
		Name:    d.Id,
		Details: details,
		Time:    t,
		Status:  status,
		Level:   d.Level,
		Source:  "kapacitor",
	}
	// extract tags
	series, ok := d.Data["series"]
	if !ok {
		return nil, fmt.Errorf("Invalid data received, no tags found")
	}
	s := series.([]interface{})
	if len(s) == 0 {
		return nil, fmt.Errorf("Invalid data received, no tags found")
	}
	sMap := s[0].(map[string]interface{})
	tags, ok := sMap["tags"]
	if !ok || tags == nil {
		return nil, fmt.Errorf("Invalid data received, no tags found")
	}
	tagMap := tags.(map[string]interface{})
	if device, ok := tagMap["device"]; ok {
		r.Device = device.(string)
	}
	if entity, ok := tagMap["entity"]; ok {
		r.Entity = entity.(string)
	}
	return r, nil
}

func init() {
	parser := &KapacitorParser{name: "kapacitor"}
	listener.AddParser(parser)
}
