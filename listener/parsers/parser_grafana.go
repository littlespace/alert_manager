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

type grafanaData struct {
	Title       string
	RuleId      int
	RuleName    string
	RuleUrl     string
	State       string
	Message     string
	EvalMatches []struct {
		Metric string
		Tags   map[string]string
		Value  float64
	}
}

type GrafanaParser struct {
	name string
}

func (p *GrafanaParser) Name() string {
	return p.name
}

func (p *GrafanaParser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := grafanaData{}
	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}
	metricText := fmt.Sprintf("\nMetric: %v, Value: %v\n", d.EvalMatches[0].Metric, d.EvalMatches[0].Value)
	l := &listener.WebHookAlertData{
		Id:      strconv.FormatInt(int64(d.RuleId), 10),
		Name:    d.RuleName,
		Details: d.Message + metricText,
		Time:    time.Now(),
		Status:  listener.Status_ALERTING,
		Source:  "grafana",
	}
	var tags []string
	for tagName, tagValue := range d.EvalMatches[0].Tags {
		if strings.ToLower(tagName) == "device" {
			l.Device = tagValue
			continue
		}
		tags = append(tags, tagValue)
	}
	l.Entity = strings.Join(tags, ":")
	if strings.ToLower(d.State) != "alerting" {
		l.Status = listener.Status_CLEARED
	}
	return l, nil
}

func init() {
	parser := &GrafanaParser{name: "grafana"}
	listener.AddParser(parser)
}
