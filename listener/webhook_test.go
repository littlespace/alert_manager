package listener

import (
	"bytes"
	ah "github.com/mayuresh82/alert_manager/handler"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type mockParser struct{}

func (m *mockParser) Name() string { return "mocked" }

func (m *mockParser) Parse(data []byte) (*WebHookAlertData, error) {
	return &WebHookAlertData{
		Id:      "1",
		Name:    "Test Alert",
		Details: "Test Alert 123",
		Device:  "dev1",
		Entity:  "ent1",
		Time:    time.Now(),
		Level:   "WARN",
		Status:  "ACTIVE",
		Source:  "mocked"}, nil
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
	mockChan := make(chan *ah.AlertEvent)
	ah.ListenChan = mockChan

	req, err := http.NewRequest("POST", "/listener/alert/?source=mocked", bytes.NewReader([]byte("blah")))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(lis.httpHandler)

	go handler.ServeHTTP(rr, req)

	event := <-mockChan
	assert.Equal(t, event.Type, ah.EventType_ACTIVE)
	assert.Equal(t, event.Alert.Name, "Test Alert")
	assert.Equal(t, event.Alert.Description, "Test Alert 123")
	assert.Equal(t, event.Alert.Device.String, "dev1")
	assert.Equal(t, event.Alert.Entity, "ent1")
	assert.Equal(t, event.Alert.Severity.String(), "WARN")
	assert.Equal(t, event.Alert.Source, "mocked")
}

func TestMain(m *testing.M) {
	p := &mockParser{}
	AddParser(p)
	tu.InitTests()
	os.Exit(m.Run())
}
