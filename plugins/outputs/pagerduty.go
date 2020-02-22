package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
)

// Based on https://v2.developer.pagerduty.com/docs/events-api-v2

type PagerDutyRecipient struct {
	RoutingKey   string
	SendResolved bool
	SendAck      bool
}

type PagerDutyMessage struct {
	RoutingKey    string `json:"routing_key"`
	EventAction   string `json:"event_action"`
	DedupKey      string `json:"dedup_key,omitempty"`
	Summary       string
	Severity      string
	Source        string
	Timestamp     string
	Component     string
	CustomDetails map[string]interface{} `json:"custom_details"`
}

type PagerDutyNotifier struct {
	URL        string
	Notif      chan *plugins.SendRequest
	Recipients map[string]*PagerDutyRecipient
}

func (p *PagerDutyNotifier) Name() string {
	return "pagerduty"
}

func (p *PagerDutyNotifier) formatBody(event *models.AlertEvent, routingKey string, weburl string) ([]byte, error) {
	component := event.Alert.Entity
	if event.Alert.Device.Valid {
		component = event.Alert.Device.String + ":" + event.Alert.Entity
	}
	var action string
	switch event.Type {
	case models.EventType_ACTIVE:
		action = "trigger"
	case models.EventType_ACKD:
		action = "acknowledge"
	case models.EventType_CLEARED:
		action = "resolve"
	default:
		return nil, fmt.Errorf("Unrecognized event")
	}
	m := &PagerDutyMessage{
		RoutingKey:  routingKey,
		EventAction: action,
		Summary:     event.Alert.Description,
		Severity:    event.Alert.Severity.String(),
		Source:      event.Alert.Source,
		Timestamp:   event.Alert.StartTime.Format(time.RFC3339),
		Component:   component,
	}
	return json.Marshal(m)
}

func (n *PagerDutyNotifier) post(data []byte, timeout time.Duration) {
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Post(n.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		glog.Errorf("Output: Unable to post to pagerduty: %v", err)
		return
	}
	if resp.StatusCode != http.StatusAccepted {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		glog.Errorf("Output: Unable to post to pagerduty: Got HTTP %d: %v", resp.StatusCode, string(body))
	}
}

func (n *PagerDutyNotifier) Start(ctx context.Context, opts *plugins.Options) {
	for {
		select {
		case req := <-n.Notif:
			event := req.Event
			recp, ok := n.Recipients[req.Name]
			if !ok {
				glog.Errorf("Failed to get recipient for output %s", req.Name)
				break
			}
			if event.Type == models.EventType_CLEARED && !recp.SendResolved {
				break
			}
			if event.Type == models.EventType_ACKD && !recp.SendAck {
				break
			}
			body, err := n.formatBody(event, recp.RoutingKey, opts.WebUrl)
			if err != nil {
				glog.Errorf("Output: Victorops: Cant get json body for alert: %v", err)
				break
			}
			n.post(body, opts.ClientTimeout)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	n := &PagerDutyNotifier{Notif: make(chan *plugins.SendRequest)}
	plugins.AddOutput(n, n.Notif)
}
