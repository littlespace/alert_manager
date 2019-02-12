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

type SlackRecipient struct {
	Team    string
	Channel string
	Upload  bool
	Token   string
	Action  string
	Mention string
}

type SlackNotifier struct {
	Url        string
	Recipients []*SlackRecipient
	Notif      chan *models.AlertEvent

	//statPostsSent stat.Stat
	//statPostsError stat.Stat
	//statsAuthFailures stat.Stat
	//statsNotFounds stat.Stat
}

func (n *SlackNotifier) Name() string {
	return "slack"
}

func (n *SlackNotifier) getRecipient(team string) *SlackRecipient {
	for _, recp := range n.Recipients {
		if recp.Team == team {
			return recp
		}
	}
	return nil
}

func (n *SlackNotifier) formatBody(event *models.AlertEvent) ([]byte, error) {
	recipient := n.getRecipient(event.Alert.Team)
	if recipient == nil {
		return []byte{}, fmt.Errorf("Failed to get recipient for team %s", event.Alert.Team)
	}
	message := recipient.Mention
	// dont send message on clear
	if event.Type != models.EventType_CLEARED {
		message += " " + event.Alert.Description
	}
	device := "None"
	if event.Alert.Device.Valid {
		device = event.Alert.Device.String
	}
	fields := []map[string]interface{}{
		map[string]interface{}{
			"title": "AlertID", "value": event.Alert.Id, "short": false,
		},
		map[string]interface{}{
			"title": "Device", "value": device, "short": false,
		},
		map[string]interface{}{
			"title": "Entity", "value": event.Alert.Entity, "short": false,
		},
	}

	title := fmt.Sprintf("[%s][%s] %s", event.Alert.Severity.String(), event.Alert.Status.String(), event.Alert.Name)
	body := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"title": title,
				// "title_link": http://alert_manager/alert?id=xxx
				"text":   message,
				"fields": fields,
				"footer": fmt.Sprintf("%s via Alert Manager", event.Alert.Source),
				"ts":     event.Alert.LastActive.Unix(),
			},
		},
		"parse": "full", // to linkify urls, users and channels in alert message.
	}
	if recipient.Channel != "" {
		body["channel"] = recipient.Channel
	}
	// TODO send imageURL via token, and uplaod file
	// https://github.com/grafana/grafana/blob/master/pkg/services/alerting/notifiers/slack.go

	return json.Marshal(&body)
}

func (n *SlackNotifier) post(data []byte) {
	c := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := c.Post(n.Url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		//n.statsPostError.Add(1)
		glog.Errorf("Output: Unable to post to slack: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		//n.statsPostError.Add(1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		glog.Errorf("Output: Unable to post to slack: Got HTTP %d: %v", resp.StatusCode, string(body))
	}
	//n.statPostsSent.Add(1)
}

func (n *SlackNotifier) Start(ctx context.Context) {
	for {
		select {
		case event := <-n.Notif:
			body, err := n.formatBody(event)
			if err != nil {
				glog.Errorf("Output: Slack: Cant get json body for alert %s: %v", event.Alert.Name, err)
				break
			}
			n.post(body)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	n := &SlackNotifier{Notif: make(chan *models.AlertEvent)}
	ah.RegisterOutput(n.Name(), n.Notif)
	plugins.AddOutput(n)
}
