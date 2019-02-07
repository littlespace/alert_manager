package aggregator

import (
	"context"
	"flag"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins/processors/aggregator/groupers"
	tu "github.com/mayuresh82/alert_manager/testutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var mockAlerts = map[string]*models.Alert{
	"bgp_1":      tu.MockAlert(1, "Neteng BGP Down", "Alert1", "d1", "e1", "src1", "scp1", "1", "INFO", []string{"a", "b"}, nil),
	"bgp_2":      tu.MockAlert(2, "Neteng BGP Down", "Alert2", "d2", "e2", "src2", "scp2", "2", "INFO", []string{"c", "d"}, nil),
	"agg_bgp_12": tu.MockAlert(12, "Neteng_Aggregated BGP Down", "Alert1"+"\n"+"Alert2"+"\n", "", "Various", "bgp_session", "scope", "1", "WARN", []string{"neteng", "bgp"}, nil),
	"a3":         tu.MockAlert(3, "Test Alert 3", "Alert3", "d3", "e3", "src3", "scp3", "3", "INFO", []string{}, nil),
	"a4":         tu.MockAlert(4, "Test Alert 4", "Alert4", "d4", "e4", "src4", "device", "4", "INFO", []string{}, nil),
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
	switch query {
	case models.QueryUpdateAggId:
		mockAlerts["bgp_1"].AggregatorId = mockAlerts["agg_bgp_12"].Id
		mockAlerts["bgp_2"].AggregatorId = mockAlerts["agg_bgp_12"].Id
	case models.QueryUpdateManyStatus:
		mockAlerts["bgp_1"].Status = models.Status_SUPPRESSED
		mockAlerts["bgp_2"].Status = models.Status_SUPPRESSED
	}
	return nil
}

func (t *MockTx) InSelect(query string, to interface{}, arg ...interface{}) error {
	return nil
}

func (t *MockTx) SelectAlerts(query string, args ...interface{}) (models.Alerts, error) {
	return models.Alerts{mockAlerts["bgp_1"], mockAlerts["bgp_2"]}, nil
}

func (tx *MockTx) NewSuppRule(rule *models.SuppressionRule) (int64, error) {
	return 1, nil
}

func (tx *MockTx) SelectRules(query string, args ...interface{}) (models.SuppRules, error) {
	return models.SuppRules{}, nil
}

func (tx *MockTx) NewRecord(alertId int64, event string) (int64, error) {
	return 1, nil
}

type mockGrouper struct {
	name string
}

func (m *mockGrouper) Name() string { return m.name }

func (m *mockGrouper) AggDesc(alerts []*models.Alert) string {
	return mockAlerts["agg_bgp_12"].Description
}

func (m *mockGrouper) Valid(alerts []*models.Alert) []*models.Alert {
	return alerts
}

func (m *mockGrouper) GrouperFunc() groupers.GroupingFunc {
	return func(i, j *models.Alert) bool {
		return i.Labels["s1"] == j.Labels["s2"]
	}
}

func TestAlertGrouping(t *testing.T) {
	group := []*models.Alert{
		tu.MockAlert(1, "Neteng BGP Down", "Alert1", "d1", "e1", "src1", "scp1", "1", "WARN", []string{"a", "b"}, nil),
		tu.MockAlert(2, "Neteng BGP Down", "Alert2", "d2", "e2", "src2", "scp2", "2", "WARN", []string{"c", "d"}, nil),
	}

	grouper := &mockGrouper{name: "bgp_session"}

	ag := alertGroup{groupedAlerts: group, grouper: grouper, ruleName: "bgp_session"}
	agg := ag.aggAlert()
	id, err := ag.saveAgg(&MockTx{}, agg)
	if err != nil {
		t.Fatal(err)
	}
	agg.Id = id
	assert.Equal(t, agg.Name, mockAlerts["agg_bgp_12"].Name)
	assert.Equal(t, agg.Description, mockAlerts["agg_bgp_12"].Description)
	assert.Equal(t, agg.Source, mockAlerts["agg_bgp_12"].Source)
	assert.Equal(t, agg.Entity, mockAlerts["agg_bgp_12"].Entity)
	assert.Equal(t, agg.Severity, mockAlerts["agg_bgp_12"].Severity)
	assert.Equal(t, agg.IsAggregate, true)

	assert.Equal(t, mockAlerts["bgp_1"].AggregatorId, mockAlerts["agg_bgp_12"].Id)
	assert.Equal(t, mockAlerts["bgp_2"].AggregatorId, mockAlerts["agg_bgp_12"].Id)

	assert.Equal(t, agg.Labels["device"], []string{"d1", "d2"})
	assert.Equal(t, agg.Labels["entity"], []string{"e1", "e2"})

	a := &Aggregator{db: &MockDb{}, statAggsActive: &tu.MockStat{}, statError: &tu.MockStat{}}
	ctx := context.Background()
	supp := ah.GetSuppressor(&MockDb{})
	out := make(chan *models.AlertEvent, 1)
	// test notif
	a.handleGrouped(ctx, &ag, out)
	event := <-out
	assert.Equal(t, event.Alert.Status.String(), "ACTIVE")
	assert.Equal(t, int64(event.Alert.Id), mockAlerts["agg_bgp_12"].Id)

	// test suppressed
	r := models.NewSuppRule(
		models.Labels{"alert_name": "Neteng_Aggregated BGP Down"},
		models.MatchCond_ALL,
		"test", "test", 5*time.Minute)
	supp.SaveRule(ctx, &MockTx{}, r)
	if err := a.handleGrouped(ctx, &ag, out); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["bgp_1"].Status, models.Status_SUPPRESSED)
	assert.Equal(t, mockAlerts["bgp_2"].Status, models.Status_SUPPRESSED)
}

func TestAggExpiry(t *testing.T) {
	a := &Aggregator{db: &MockDb{}, statAggsActive: &tu.MockStat{}, statError: &tu.MockStat{}}
	ctx := context.Background()
	out := make(chan *models.AlertEvent, 1)

	// test no change
	mockAlerts["bgp_1"].Status = models.Status_EXPIRED
	if err := a.checkExpired(ctx, out); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["agg_bgp_12"].Status.String(), "ACTIVE")

	// test all expired
	mockAlerts["bgp_2"].Status = models.Status_EXPIRED
	if err := a.checkExpired(ctx, out); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mockAlerts["agg_bgp_12"].Status.String(), "EXPIRED")
	event := <-out
	assert.Equal(t, event.Type, models.EventType_EXPIRED)
	assert.Equal(t, event.Alert.Id, mockAlerts["agg_bgp_12"].Id)
}

func TestGrouperMatch(t *testing.T) {
	a := &Aggregator{db: &MockDb{}, statAggsActive: &tu.MockStat{}, statError: &tu.MockStat{}}
	for _, a := range mockAlerts {
		a.ExtendLabels()
	}
	// test configured grouper
	grouper := a.grouperForAlert(mockAlerts["bgp_1"], "bgp_session")
	assert.Equal(t, grouper.Name(), "bgp_session")
	// test no grouper
	grouper = a.grouperForAlert(mockAlerts["a3"], "label_group")
	assert.Nil(t, grouper)
	// test generic grouper
	grouper = a.grouperForAlert(mockAlerts["a4"], "label_group")
	assert.Equal(t, grouper.Name(), "default_label_grouper")
	g := grouper.(*groupers.LabelGrouper)
	rule, _ := ah.Config.GetAggregationRuleConfig("label_group")
	assert.ElementsMatch(t, g.Groupby, rule.GroupBy)
}

func TestMain(m *testing.M) {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../../../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
