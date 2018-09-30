package aggregator

import (
	"context"
	"flag"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var mockAlerts = map[string]*models.Alert{
	"bgp_1":      tu.MockAlert(1, "Neteng BGP Down", "Alert1", "d1", "e1", "src1", "scp1", "1", "INFO", []string{"a", "b"}, nil),
	"bgp_2":      tu.MockAlert(2, "Neteng BGP Down", "Alert2", "d2", "e2", "src2", "scp2", "2", "INFO", []string{"c", "d"}, nil),
	"agg_bgp_12": tu.MockAlert(12, "Neteng_Aggregated BGP Down", "Alert1"+"\n"+"Alert2"+"\n", "", "Various", "bgp_session", "scope", "1", "WARN", []string{"neteng", "bgp"}, nil),
}

type MockDb struct{}

func (m *MockDb) NewTx() models.Txn {
	return &MockTx{}
}

func (m *MockDb) Close() error {
	return nil
}

type MockTx struct {
	*models.Tx
}

func (t *MockTx) NewAlert(alert *models.Alert) (int64, error) {
	return mockAlerts["agg_bgp_12"].Id, nil
}

func (t *MockTx) UpdateAlert(alert *models.Alert) error {
	return nil
}

func (t *MockTx) Rollback() error {
	return nil
}

func (t *MockTx) Commit() error {
	return nil
}

func (t *MockTx) GetAlert(query string, args ...interface{}) (*models.Alert, error) {
	return mockAlerts["agg_bgp_12"], nil
}

func (t *MockTx) InQuery(query string, args ...interface{}) error {
	mockAlerts["bgp_1"].SetAggId(mockAlerts["agg_bgp_12"].Id)
	mockAlerts["bgp_2"].SetAggId(mockAlerts["agg_bgp_12"].Id)
	return nil
}

func (t *MockTx) InSelect(query string, to interface{}, arg ...interface{}) error {
	return nil
}

func (t *MockTx) SelectAlerts(query string, args ...interface{}) (models.Alerts, error) {
	return models.Alerts{*mockAlerts["bgp_1"], *mockAlerts["bgp_2"]}, nil
}

type mockGrouper struct {
	name string
}

func (m *mockGrouper) Name() string { return m.name }

func (m *mockGrouper) DoGrouping(alerts []*models.Alert) [][]*models.Alert {
	return [][]*models.Alert{
		[]*models.Alert{mockAlerts["bgp_1"]},
		[]*models.Alert{mockAlerts["bgp_2"]},
	}
}

var notif = make(chan *ah.AlertEvent, 1)

func TestAlertGrouping(t *testing.T) {
	group := []*models.Alert{
		tu.MockAlert(1, "Neteng BGP Down", "Alert1", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil),
		tu.MockAlert(2, "Neteng BGP Down", "Alert2", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil),
	}

	grouper := &mockGrouper{name: "bgp_session"}

	ag := alertGroup{groupedAlerts: group, grouper: grouper}
	agg := ag.aggAlert()
	assert.Equal(t, agg.Name, mockAlerts["agg_bgp_12"].Name)
	assert.Equal(t, agg.Description, mockAlerts["agg_bgp_12"].Description)
	assert.Equal(t, agg.Source, mockAlerts["agg_bgp_12"].Source)
	assert.Equal(t, agg.Entity, mockAlerts["agg_bgp_12"].Entity)
	assert.Equal(t, agg.Severity, mockAlerts["agg_bgp_12"].Severity)
	assert.Equal(t, agg.IsAggregate, true)

	if err := ag.saveAgg(context.Background(), &MockTx{}); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["bgp_1"].AggregatorId.Int64, mockAlerts["agg_bgp_12"].Id)
	assert.Equal(t, mockAlerts["bgp_2"].AggregatorId.Int64, mockAlerts["agg_bgp_12"].Id)

	event := <-notif
	assert.Equal(t, event.Type, ah.EventType_ACTIVE)
	assert.Equal(t, event.Alert.Id, mockAlerts["agg_bgp_12"].Id)
}

func TestAggExpiry(t *testing.T) {
	a := &Aggregator{}
	ctx := context.Background()
	tx := &MockTx{}

	// test no change
	mockAlerts["bgp_1"].Status = models.Status_CLEARED
	if err := a.checkExpired(ctx, tx); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["agg_bgp_12"].Status.String(), "ACTIVE")

	// test all cleared
	mockAlerts["bgp_2"].Status = models.Status_CLEARED
	if err := a.checkExpired(ctx, tx); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["agg_bgp_12"].Status.String(), "CLEARED")
	event := <-notif
	assert.Equal(t, event.Type, ah.EventType_CLEARED)
	assert.Equal(t, event.Alert.Id, mockAlerts["agg_bgp_12"].Id)

	// test all expired
	mockAlerts["bgp_1"].Status = models.Status_EXPIRED
	mockAlerts["bgp_2"].Status = models.Status_EXPIRED
	if err := a.checkExpired(ctx, tx); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["agg_bgp_12"].Status.String(), "EXPIRED")
	event = <-notif
	assert.Equal(t, event.Type, ah.EventType_EXPIRED)
	assert.Equal(t, event.Alert.Id, mockAlerts["agg_bgp_12"].Id)
}

func TestMain(m *testing.M) {
	flag.Parse()
	ah.Outputs["slack"] = notif
	ah.Config = ah.NewConfigHandler("../../../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
