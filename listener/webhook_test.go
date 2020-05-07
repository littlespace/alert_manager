package listener

import (
	"bytes"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

type mockParser struct{}

func (m *mockParser) Name() string { return "mocked" }

func (m *mockParser) Parse(data []byte) (*WebHookAlertData, error) {
	alert := &WebHookAlert{
		Id:      "1",
		Name:    "Test Alert",
		Details: "Test Alert 123",
		Device:  "dev1",
		Entity:  "ent1",
		Time:    time.Now(),
		Level:   "WARN",
		Status:  "ACTIVE",
		Source:  "mocked"}
	return &WebHookAlertData{Alerts: []*WebHookAlert{alert}}, nil
}

func TestAlertHandlerBadRequest(t *testing.T) {
	lis := &WebHookListener{statRequestsRecvd: &tu.MockStat{}, statRequestsError: &tu.MockStat{}}

	// test empty request
	req, err := http.NewRequest("POST", "/listener/alert", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(lis.httpHandler)
	handler.ServeHTTP(rr, req)
	expected := http.StatusInternalServerError
	assert.Equal(t, rr.Code, expected)

	// test no query
	body := bytes.NewReader([]byte("blah"))
	req, err = http.NewRequest("POST", "/listener/alert", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(lis.httpHandler)
	handler.ServeHTTP(rr, req)
	expected = http.StatusBadRequest
	assert.Equal(t, rr.Code, expected)

	// test no valid parser
	req, err = http.NewRequest("POST", "/listener/alert/?source=abcd", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(lis.httpHandler)
	handler.ServeHTTP(rr, req)
	expected = http.StatusInternalServerError
	assert.Equal(t, rr.Code, expected)
}

func TestAlertHandlerParsing(t *testing.T) {
	lis := &WebHookListener{statRequestsRecvd: &tu.MockStat{}, statRequestsError: &tu.MockStat{}}

	req, err := http.NewRequest("POST", "/listener/alert/?source=mocked&team=foo", bytes.NewReader([]byte("blah")))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(lis.httpHandler)

	go handler.ServeHTTP(rr, req)

	event := <-ah.ListenChan
	assert.Equal(t, event.Type, models.EventType_ACTIVE)
	assert.Equal(t, event.Alert.Name, "Test Alert")
	assert.Equal(t, event.Alert.Description, "Test Alert 123")
	assert.Equal(t, event.Alert.Device.String, "dev1")
	assert.Equal(t, event.Alert.Entity, "ent1")
	assert.Equal(t, event.Alert.Severity.String(), "WARN")
	assert.Equal(t, event.Alert.Source, "mocked")
	assert.Equal(t, event.Alert.Team, "foo")
}

func TestMain(m *testing.M) {
	p := &mockParser{}
	AddParser(p)
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
