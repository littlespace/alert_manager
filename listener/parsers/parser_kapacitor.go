package parsers

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
	"time"
)

type kapacitorData struct {
	ID      string // any unique ID
	Message string // Name of the alert
	Device  string
	Entity  string // entity that can be constructed from Tags
	Details string // Info message of the alert
	Time    string // time triggered
	Level   string // INFO, WARN or CRITICAL
	Data    string // time series data
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
	details := d.Details + "\n" + d.Data
	timelayout := "2006-01-02T15:04:05.000Z"
	t, err := time.Parse(timelayout, d.Time)
	if err != nil {
		glog.Errorf("Unable to parse time string , using current time")
		t = time.Now()
	}
	return &listener.WebHookAlertData{
		Id:      d.ID,
		Name:    d.Message,
		Details: details,
		Device:  d.Device,
		Entity:  d.Entity,
		Time:    t,
		//Status:  ,
		Level:  d.Level,
		Source: "kapacitor",
	}, nil
}

func init() {
	parser := &KapacitorParser{name: "kapacitor"}
	listener.AddParser(parser)
}
