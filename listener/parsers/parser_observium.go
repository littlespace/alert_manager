package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
)

type observiumData struct {
	State             string `json:"alert_state"`
	Url               string `json:"alert_url"`
	Id                string `json:"alert_id"`
	Message           string `json:"alert_message"`
	Timestamp         string `json:"alert_timestamp_rfc3339"`
	EntityName        string `json:"entity_name"`
	EntityType        string `json:"entity_type"`
	EntityDescription string `json:"entity_description"`
	DeviceName        string `json:"device_sysname"`
	DeviceOs          string `json:"device_os"`
	Title             string
}

type ObserviumParser struct {
	name string
}

func (p *ObserviumParser) Name() string {
	return p.name
}

func (p *ObserviumParser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := observiumData{}
	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}
	if d.Message == "" || d.EntityName == "" {
		return nil, fmt.Errorf("Invalid alert data received")
	}
	t, err := time.Parse(time.RFC3339, d.Timestamp)
	if err != nil {
		glog.Errorf("Unable to parse time string , using current time")
		t = time.Now()
	}
	l := &listener.WebHookAlert{
		Id:      d.Id,
		Name:    d.Message,
		Details: d.Title + " / " + d.EntityDescription,
		Device:  d.DeviceName,
		Entity:  d.EntityName,
		Time:    t,
		Source:  "observium",
		Status:  listener.Status_ALERTING,
	}
	if strings.ToLower(d.State) == "recover" {
		l.Status = listener.Status_CLEARED
	}
	return &listener.WebHookAlertData{Alerts: []*listener.WebHookAlert{l}}, nil
}

func init() {
	parser := &ObserviumParser{name: "observium"}
	listener.AddParser(parser)
}
