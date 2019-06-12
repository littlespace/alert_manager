package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
)

var mockRules = map[string]models.SuppressionRule{
	"rule1": models.SuppressionRule{Id: 1, Name: "rule1", Duration: 60},
	"rule2": models.SuppressionRule{Id: 2, Name: "rule2", Duration: 60},
}

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
	if query == models.QueryDeleteSuppRules {
		delete(mockRules, "rule1")
	}
	return nil
}

func (tx *MockTx) InSelect(query string, to interface{}, arg ...interface{}) error {
	return nil
}

func (tx *MockTx) SelectAlerts(query string, arg ...interface{}) (models.Alerts, error) {
	var alerts models.Alerts
	for i := 1; i <= 2; i++ {
		a := &models.Alert{
			Id:          int64(i),
			Status:      models.Status_ACTIVE,
			Name:        "mock",
			Description: "test",
			Entity:      "e1",
			Source:      "src",
			Scope:       "scp",
			Labels:      make(models.Labels),
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

func (tx *MockTx) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	return &models.Alert{
		Status: models.Status_ACTIVE,
		Id:     args[0].(int64),
		Labels: make(models.Labels)}, nil
}

func (tx *MockTx) SelectRules(query string, args ...interface{}) (models.SuppRules, error) {
	return models.SuppRules{
		&models.SuppressionRule{Id: 1, Name: "rule1", Duration: 60},
		&models.SuppressionRule{Id: 2, Name: "rule2", Duration: 60},
	}, nil
}

func (t *MockTx) NewInsert(query string, item interface{}) (int64, error) {
	return 1, nil
}

func (tx *MockTx) NewRecord(alertId int64, event string) (int64, error) {
	return 1, nil
}

func (tx *MockTx) Commit() error {
	return nil
}

func (tx *MockTx) Rollback() error {
	return nil
}

func (tx *MockTx) Exec(query string, args ...interface{}) error {
	return nil
}

type MockAuthProvider struct{}

func (m *MockAuthProvider) Authenticate(userid, password string) (bool, error) {
	if userid == "baz" {
		return false, nil
	}
	return true, nil
}

func NewMockServer() *Server {
	d := &MockDb{}
	return &Server{
		handler: &ah.AlertHandler{
			Db:         d,
			Suppressor: ah.GetSuppressor(d),
			Teams:      models.Teams{&models.Team{Name: "bar"}},
		},
		authProvider:      &MockAuthProvider{},
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

	// test failed auth
	data, _ := json.Marshal(&User{Username: "baz", Password: "bar"})
	body := bytes.NewReader(data)
	req, err = http.NewRequest("POST", "/api/auth", body)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

	// test successful auth
	data, _ = json.Marshal(&User{Username: "foo", Password: "bar"})
	body = bytes.NewReader(data)
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
	assert.Equal(t, tk.ExpiresAt > 0, true)

	// test admin user
	s.admin = &User{Username: "foo", Password: "bar"}
	body = bytes.NewReader(data)
	req, _ = http.NewRequest("POST", "/api/auth", body)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	tk = JwtToken{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&tk); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, tk.ExpiresAt, int64(0))

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
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

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
	router.HandleFunc("/api/{category}", s.GetItems).Methods("GET")
	router.HandleFunc("/api/{category}/{id}", s.GetAlert).Methods("GET")

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
	assert.Equal(t, a["id"].(float64), float64(2))

	req, err = http.NewRequest("GET", "/api/suppression_rules", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	b := []interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(b), 2)
}

func TestServerUpdate(t *testing.T) {
	s := NewMockServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/alerts/{id}", s.Update).Methods("PATCH")

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
	assert.Equal(t, a["id"].(float64), float64(1))
	assert.Equal(t, a["status"].(string), "SUPPRESSED")

	// test clear
	req, _ = http.NewRequest("PATCH", "/api/alerts/1/clear", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	a = map[string]interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["id"].(float64), float64(1))
	assert.Equal(t, a["status"].(string), "CLEARED")

	// test ack
	req, _ = http.NewRequest("PATCH", "/api/alerts/1/ack?owner=foo&team=bar", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	a = map[string]interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["id"].(float64), float64(1))
	assert.Equal(t, a["status"].(string), "ACTIVE")
	assert.Equal(t, a["owner"].(string), "foo")
	assert.Equal(t, a["team"].(string), "bar")
}

func TestSuppRule(t *testing.T) {
	s := NewMockServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/suppression_rules", s.CreateSuppRule).Methods("POST")
	router.HandleFunc("/api/suppression_rules/{id}/clear", s.ClearSuppRule).Methods("DELETE")
	router.HandleFunc("/api/suppression_rules/persistent", s.GetPersistentRules).Methods("GET")

	req, _ := http.NewRequest("POST", "/api/suppression_rules", nil)
	body, _ := json.Marshal(&map[string]interface{}{
		"Rtype": 0,
		"Name":  "test2",
		"Entities": map[string]interface{}{
			"alert_name": "Test Alert",
		},
		"Duration": 300,
		"Reason":   "foo",
		"Creator":  "test",
	})
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	a := map[string]interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["id"].(float64), float64(1))

	req, _ = http.NewRequest("DELETE", "/api/suppression_rules/1/clear", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	_, ok := mockRules["rule1"]
	assert.Equal(t, ok, false)

	// test persistent rules
	req, _ = http.NewRequest("GET", "/api/suppression_rules/persistent", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	b := []interface{}{}
	if err := json.NewDecoder(rr.Result().Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(b), 1)
	c := b[0].(map[string]interface{})
	assert.Equal(t, c["id"].(float64), float64(0))
	assert.Equal(t, c["name"].(string), "Dummy SuppRule")
	assert.Equal(t, c["creator"].(string), "alert manager")
	assert.Equal(t, c["dont_expire"].(bool), true)
}

func TestMain(m *testing.M) {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
