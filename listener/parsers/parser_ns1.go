package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/listener"
)

type ns1Data struct {
	Job struct {
		JobID   string `json:"id"`
		JobType string `json:"job_type"`
		Name    string
		Config  struct {
			Host string
			Port int
		}
	}
	Region string
	Since  int
	State  string
}

type Ns1Parser struct {
	name string
}

func (p *Ns1Parser) Name() string {
	return p.name
}

func (p *Ns1Parser) Parse(data []byte) (*listener.WebHookAlertData, error) {
	d := ns1Data{}

	if err := json.Unmarshal(data, &d); err != nil {
		glog.Errorf("Unable to decode json: %v", err)
		return nil, err
	}

	if !(d.Job.JobType == "tcp") {
		return nil, fmt.Errorf("only TCP monitor are supported")
	}
	if d.Job.Config.Host == "" {
		return nil, fmt.Errorf("invalid data received, Config/Host address is mandatory")
	}
	if d.Job.Name == "" {
		return nil, fmt.Errorf("Invalid data received, Name is mandatory")
	}

	l := &listener.WebHookAlertData{
		Id:      d.Job.JobID,
		Name:    "Neteng DNS Monitor Down",
		Details: d.Job.Name,
		Status:  listener.Status_ALERTING,
		Time:    time.Now(),
		Source:  "ns1",
		Entity:  d.Job.Name,
		Labels:  make(models.Labels),
	}

	l.Labels["vipIp"] = d.Job.Config.Host

	if strings.ToLower(d.State) == "up" {
		l.Status = listener.Status_CLEARED
	}

	return l, nil
}

func init() {
	parser := &Ns1Parser{name: "ns1"}
	listener.AddParser(parser)
}
