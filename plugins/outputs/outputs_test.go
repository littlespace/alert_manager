package output

import (
	"encoding/json"
	"fmt"
	ah "github.com/mayuresh82/alert_manager/handler"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOutputs(t *testing.T) {
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
	s := &SlackNotifier{Url: ts.URL, Recipient: "#test"}
	event := tu.MockAlertEvent(
		ah.EventType_ACTIVE,
		"Neteng BGP Down",
		"This alert has fired",
		"WARN",
		"dev1", "PeerX", "src", "scp", "1")

	data, err := s.formatBody(event)
	if err != nil {
		t.Fatal(err)
	}
	s.post(data)
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
