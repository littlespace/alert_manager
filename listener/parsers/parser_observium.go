package parsers

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/listener"
	"strconv"
	"strings"
	"time"
)

type observiumData struct {
	State             string `json:"alert_state"`
	Url               string `json:"alert_url"`
	Id                int    `json"alert_id"`
	Message           string `json:"alert_message"`
	Timestamp         string
	EntityName        string `json:"entity_name"`
	EntityDescription string `json:"entity_description"`
	DeviceName        string `json:"device_hostname"`
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
	timelayout := "2006-01-02 04:05:22"
	t, err := time.Parse(timelayout, d.Timestamp)
	if err != nil {
		glog.Errorf("Unable to parse time string , using current time")
		t = time.Now()
	}
	l := &listener.WebHookAlertData{
		Id:      strconv.FormatInt(int64(d.Id), 10),
		Name:    d.Title,
		Details: d.Message + fmt.Sprintf(" Url: %s", d.Url),
		Device:  d.DeviceName,
		Entity:  d.EntityName,
		Time:    t,
		Source:  "observium",
		Status:  listener.Status_ALERTING,
	}
	if strings.ToLower(d.State) == "recover" {
		l.Status = listener.Status_CLEARED
	}
	return l, nil
}

func init() {
	parser := &ObserviumParser{name: "observium"}
	listener.AddParser(parser)
}
