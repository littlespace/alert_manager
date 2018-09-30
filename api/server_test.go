package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type MockDb struct{}

func (d MockDb) NewTx() models.Txn {
	return &MockTx{}
}

func (d MockDb) Close() error {
	return nil
}

type MockTx struct {
	*models.Tx
}

func (tx *MockTx) InQuery(query string, arg ...interface{}) error {
	return nil
}

func (tx *MockTx) SelectAlerts(query string, arg ...interface{}) (models.Alerts, error) {
	var alerts models.Alerts
	for i := 1; i <= 2; i++ {
		a := models.Alert{
			Id:          int64(i),
			Status:      models.Status_ACTIVE,
			Name:        "mock",
			Description: "test",
			Entity:      "e1",
			Source:      "src",
			Scope:       "scp",
		}
		if len(arg) == 0 {
			alerts = append(alerts, a)
			continue
		}
		for _, ar := range arg {
			if ar.(string) == strconv.Itoa(int(a.Id)) {
				alerts = append(alerts, a)
			}
		}
	}
	return alerts, nil
}

func (tx *MockTx) UpdateAlert(alert *models.Alert) error {
	return nil
}

func (t *MockTx) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	return &models.Alert{
		Status: models.Status_ACTIVE,
		Id:     args[0].(int64)}, nil
}

func (t *MockTx) NewSuppRule(r *models.SuppressionRule) (int64, error) {
	return 1, nil
}

func (tx *MockTx) Commit() error {
	return nil
}

func (t *MockTx) Rollback() error {
	return nil
}

func NewMockServer() *Server {
	return &Server{
		handler:           &ah.AlertHandler{Db: &MockDb{}},
		statGets:          &tu.MockStat{},
		statPosts:         &tu.MockStat{},
		statPatches:       &tu.MockStat{},
		statError:         &tu.MockStat{},
		statsAuthFailures: &tu.MockStat{},
	}
}

func TestServerAuth(t *testing.T) {
	s := NewMockServer()

	// test empty auth request
	req, err := http.NewRequest("POST", "/api/auth", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.CreateToken)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusBadRequest)

	// test valid token
	data, _ := json.Marshal(&User{Username: "foo", Password: "bar"})
	body := bytes.NewReader(data)
	req, err = http.NewRequest("POST", "/api/auth", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	tk := JwtToken{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&tk); err != nil {
		t.Fatal(err)
	}

	var testFunc = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	// no auth header
	req, err = http.NewRequest("PATCH", "/api/alerts/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(s.Validate(testFunc))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusBadRequest)

	// invalid token
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Im1nYWl0b25kZSIsInBhc3N3b3JkIjoiYWJjMTIzIn0.N5yn1XuagRR_pbhpDTsKReOETWpbmy_wNTvC1bJi2D4")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(s.Validate(testFunc))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusInternalServerError)

	// valid token
	req, err = http.NewRequest("PATCH", "/api/alerts/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tk.Token))
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(s.Validate(testFunc))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestServerGet(t *testing.T) {
	s := NewMockServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/alerts", s.GetAlerts).Methods("GET")
	router.HandleFunc("/api/alerts/{id}", s.GetAlert).Methods("GET")

	req, err := http.NewRequest("GET", "/api/alerts", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var m []interface{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(m), 2)

	req, err = http.NewRequest("GET", "/api/alerts?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	m = []interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(m), 1)
	alert := m[0].(map[string]interface{})
	assert.Equal(t, alert["Id"].(float64), float64(1))

	req, err = http.NewRequest("GET", "/api/alerts/2", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var a map[string]interface{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["Id"].(float64), float64(2))
}

func TestServerUpdate(t *testing.T) {
	s := NewMockServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/alerts/{id}", s.UpdateAlert).Methods("PATCH")

	// test update invalid query
	req, _ := http.NewRequest("PATCH", "/api/alerts/1?owner=foo&owner=bar", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusBadRequest)

	// test valid update
	req, _ = http.NewRequest("PATCH", "/api/alerts/1?owner=foo&status=2", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestServerAlertAction(t *testing.T) {
	ah.Config = ah.NewConfigHandler("../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	s := NewMockServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/alerts/{id}/{action}", s.ActionAlert).Methods("PATCH")

	// test suppress invalid query
	req, _ := http.NewRequest("PATCH", "/api/alerts/1/suppress", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusBadRequest)

	// test suppress valid
	req, _ = http.NewRequest("PATCH", "/api/alerts/1/suppress?duration=2m", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var a map[string]interface{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["Id"].(float64), float64(1))
	assert.Equal(t, a["Status"].(string), "SUPPRESSED")

	// test clear
	req, _ = http.NewRequest("PATCH", "/api/alerts/1/clear", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	a = map[string]interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["Id"].(float64), float64(1))
	assert.Equal(t, a["Status"].(string), "CLEARED")

	// test ack
	req, _ = http.NewRequest("PATCH", "/api/alerts/1/ack?owner=foo&team=bar", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	a = map[string]interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["Id"].(float64), float64(1))
	assert.Equal(t, a["Status"].(string), "ACTIVE")
	assert.Equal(t, a["Owner"].(string), "foo")
	assert.Equal(t, a["Team"].(string), "bar")
}
