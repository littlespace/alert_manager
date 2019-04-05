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
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
)

type VoRecipient struct {
	Team        string
	Url         string
	AutoResolve bool `mapstructure:"auto_resolve"`
	SendAck     bool `mapstructure:"send_ack"`
}

type victorOpsMsg struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	StartTime         string `json:"state_start_time"`
}

type VictorOpsNotifier struct {
	Notif      chan *models.AlertEvent
	Recipients []*VoRecipient
}

func (n *VictorOpsNotifier) Name() string {
	return "victorops"
}

func (n *VictorOpsNotifier) getRecipient(team string) *VoRecipient {
	for _, recp := range n.Recipients {
		if recp.Team == team {
			return recp
		}
	}
	return nil
}

func (n *VictorOpsNotifier) formatBody(event *models.AlertEvent, weburl string) ([]byte, error) {
	m := &victorOpsMsg{}
	switch event.Type {
	case models.EventType_ACTIVE, models.EventType_ESCALATED:
		m.MessageType = "CRITICAL"
	case models.EventType_CLEARED:
		m.MessageType = "RECOVERY"
	case models.EventType_ACKD:
		m.MessageType = "ACKNOWLEDGEMENT"
	}

	var device string
	if event.Alert.Device.Valid {
		device = event.Alert.Device.String
	}
	stateMsg := fmt.Sprintf("AM Url: %s/%d", weburl, event.Alert.Id) + "\n" + event.Alert.Description
	m.EntityID = fmt.Sprintf("%s:%s:%s", event.Alert.Name, device, event.Alert.Entity)
	m.EntityDisplayName = fmt.Sprintf("[%s][%s] %s , Device: %s, Entity: %s",
		event.Alert.Severity.String(), event.Alert.Status.String(), event.Alert.Name, device, event.Alert.Entity)
	m.StateMessage = stateMsg
	m.StartTime = event.Alert.StartTime.String()

	return json.Marshal(m)
}

func (n *VictorOpsNotifier) post(data []byte, url string, timeout time.Duration) {
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		glog.Errorf("Output: Unable to post to victorops: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		glog.Errorf("Output: Unable to post to victorops: Got HTTP %d: %v", resp.StatusCode, string(body))
	}
}

func (n *VictorOpsNotifier) Start(ctx context.Context, opts *plugins.Options) {
	for {
		select {
		case event := <-n.Notif:
			recp := n.getRecipient(event.Alert.Team)
			if recp == nil {
				glog.Errorf("Failed to get recipient for team %s", event.Alert.Team)
				break
			}
			if event.Type == models.EventType_CLEARED && !recp.AutoResolve {
				break
			}
			if event.Type == models.EventType_ACKD && !recp.SendAck {
				break
			}
			body, err := n.formatBody(event, opts.WebUrl)
			if err != nil {
				glog.Errorf("Output: Victorops: Cant get json body for alert: %v", err)
				break
			}
			n.post(body, recp.Url, opts.ClientTimeout)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	n := &VictorOpsNotifier{Notif: make(chan *models.AlertEvent)}
	ah.RegisterOutput(n.Name(), n.Notif)
	plugins.AddOutput(n)
}
