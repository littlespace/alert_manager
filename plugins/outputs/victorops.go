package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"io/ioutil"
	"net/http"
	"time"
)

type victorOpsMsg struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	StartTime         string `json:"state_start_time"`
}

type VictorOpsNotifier struct {
	Notif       chan *ah.AlertEvent
	Url         string
	AutoResolve bool `mapstructure:"auto_resolve"`
}

func (n *VictorOpsNotifier) Name() string {
	return "victorops"
}

func (n *VictorOpsNotifier) formatBody(event *ah.AlertEvent) ([]byte, error) {
	m := &victorOpsMsg{}
	switch event.Type {
	case ah.EventType_ACTIVE, ah.EventType_ESCALATED:
		m.MessageType = "CRITICAL"
	case ah.EventType_CLEARED:
		m.MessageType = "RECOVERY"
	case ah.EventType_ACKD:
		m.MessageType = "ACKNOWLEDGEMENT"
	}

	var device string
	if event.Alert.Device.Valid {
		device = event.Alert.Device.String
	}
	m.EntityID = fmt.Sprintf("%s:%s:%s", event.Alert.Name, device, event.Alert.Entity)
	m.EntityDisplayName = fmt.Sprintf("[%s][%s] %s , Device: %s, Entity: %s",
		event.Alert.Severity.String(), event.Alert.Status.String(), event.Alert.Name, device, event.Alert.Entity)
	m.StateMessage = event.Alert.Description
	m.StartTime = event.Alert.StartTime.String()

	return json.Marshal(m)
}

func (n *VictorOpsNotifier) post(data []byte) {
	c := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := c.Post(n.Url, "application/json", bytes.NewBuffer(data))
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

func (n *VictorOpsNotifier) Start(ctx context.Context) {
	for {
		select {
		case event := <-n.Notif:
			if event.Type == ah.EventType_CLEARED && !n.AutoResolve {
				break
			}
			body, err := n.formatBody(event)
			if err != nil {
				glog.Errorf("Output: Victorops: Cant get json body for alert: %v", err)
				break
			}
			n.post(body)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	n := &VictorOpsNotifier{Notif: make(chan *ah.AlertEvent)}
	ah.RegisterOutput(n.Name(), n.Notif)
	am.AddOutput(n)
}
