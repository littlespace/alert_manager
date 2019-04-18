package output

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

func TestOutputSlack(t *testing.T) {
	var body []byte
	var err error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	s := &SlackNotifier{
		Url: ts.URL,
		Recipients: map[string]*SlackRecipient{
			"default": &SlackRecipient{Channel: "#test"},
		},
	}
	event := &models.AlertEvent{
		Type:  models.EventType_ACTIVE,
		Alert: tu.MockAlert(0, "Neteng BGP Down", "This alert has fired", "dev1", "PeerX", "src", "scp", "t1", "1", "WARN", []string{}, nil),
	}

	data, err := s.formatBody(&plugins.SendRequest{Name: "default", Event: event}, "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	s.post(data, 2*time.Second)
	res := make(map[string]interface{})
	err = json.Unmarshal(body, &res)
	if err != nil {
		t.Fatal(err)
	}
	a := res["attachments"].([]interface{})[0]
	exp := a.(map[string]interface{})
	assert.Equal(t, exp["title"].(string), "[WARN][ACTIVE] Neteng BGP Down")
	assert.Equal(t, exp["text"].(string), " This alert has fired")
	assert.Equal(t, exp["footer"].(string), "src via Alert Manager")
	fields := exp["fields"].([]interface{})
	for _, f := range fields {
		e := f.(map[string]interface{})
		switch e["title"].(string) {
		case "AlertID":
			assert.Equal(t, e["value"].(float64), float64(0))
		case "Device":
			assert.Equal(t, e["value"].(string), "dev1")
		case "Entity":
			assert.Equal(t, e["value"].(string), "PeerX")
		}
	}
	assert.Equal(t, res["channel"].(string), "#test")
}

type mockEmailer struct {
	subject, body string
	from          string
	to            []string
}

func (m *mockEmailer) send(addr, username, pwd, from, subject, body string, to []string) error {
	m.subject = subject
	m.body = body
	m.from = from
	m.to = to
	return nil
}

var mockTpl = `
  {{.Header}} {{.EventType}} {{.AlertSeverity}}
  {{- range .AlertParams }}
  {{.Name}}: {{.Value}}
  {{- end}}
`

var renderedTpl = `
  [CRITICAL][ACTIVE] Test Alert ACTIVE CRITICAL
  Name: Test Alert
  Description: Test Desc
  Entity: testent
  StartTime: Mon Jan 2 22:04:05 UTC 2006
`

func TestOutputEmail(t *testing.T) {
	emailer := &mockEmailer{}
	n := &EmailNotifier{
		Emailer: emailer,
		rawTpl:  mockTpl,
		Recipients: map[string]*EmailRecipient{
			"default": &EmailRecipient{From: "a@foo.com", To: []string{"b@bar.com"}},
		},
	}
	event := &models.AlertEvent{
		Type: models.EventType_ACTIVE,
		Alert: &models.Alert{
			Id:          1,
			Severity:    models.Sev_CRITICAL,
			Status:      models.Status_ACTIVE,
			Name:        "Test Alert",
			Description: "Test Desc",
			Entity:      "testent",
			Team:        "t1",
			StartTime:   models.MyTime{time.Unix(1136239445, 0)},
		},
	}
	n.start(&plugins.SendRequest{Name: "default", Event: event}, "htt://localhost")
	assert.Equal(t, emailer.subject, "Alert Manager: [ACTIVE] Test Alert: [testent]")
	assert.Equal(t, emailer.body, renderedTpl)
	assert.Equal(t, emailer.from, "a@foo.com")
	assert.Equal(t, emailer.to, []string{"b@bar.com"})
}
