package inhibitor

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
	"dev_1": tu.MockAlert(1, "Neteng Device Down", "Alert1", "d1", "e1", "src1", "scp1", "1", "INFO", []string{"a", "b"},
		models.Labels{"Name": "d1"}),
	"bgp_1": tu.MockAlert(2, "Neteng BGP Down", "Alert2", "d2", "e2", "src2", "scp2", "2", "INFO", []string{"c", "d"},
		models.Labels{"RemoteDeviceName": "d1"}),
	"link_1": tu.MockAlert(3, "Neteng DC Link Down", "Alert3", "d2", "e2", "src2", "scp3", "3", "WARN", []string{"e", "f"},
		models.Labels{"ZSideDeviceName": "d1"}),
	"link_2": tu.MockAlert(4, "Neteng DC Link Down", "Alert3", "d2", "e2", "src2", "scp2", "4", "INFO", []string{"c", "d"},
		models.Labels{"ZSideDeviceName": "d4"}),
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

func (tx *MockTx) InSelect(q string, to interface{}, args ...interface{}) error {
	if to, ok := to.(*models.Alerts); ok {
		*to = append(*to, mockAlerts["dev_1"])
	}
	return nil
}

func (tx *MockTx) UpdateAlert(a *models.Alert) error {
	return nil
}

func (tx *MockTx) NewRecord(alertId int64, event string) (int64, error) {
	return 1, nil
}

func (tx *MockTx) Rollback() error {
	return nil
}

func (tx *MockTx) Commit() error {
	return nil
}

func TestInhibit(t *testing.T) {
	i := &Inhibitor{
		alertBuf:            make(map[string][]*models.Alert),
		db:                  &MockDb{},
		statAlertsInhibited: &tu.MockStat{},
		statError:           &tu.MockStat{},
	}
	ctx := context.Background()
	rule, ok := ah.Config.GetInhibitRuleConfig("Device down")
	if !ok {
		t.Fatal("Rule not found")
	}
	out := make(chan *ah.AlertEvent, 1)
	// test rule no match
	i.addAlert("Device down", mockAlerts["link_2"])
	i.checkRule(ctx, rule, out)
	assert.Equal(t, mockAlerts["link_2"].Status, models.Status_ACTIVE)
	assert.Equal(t, len(i.alertBuf["Device down"]), 0)
	event := <-out
	assert.Equal(t, event.Type, ah.EventType_ACTIVE)
	assert.Equal(t, event.Alert.Id, mockAlerts["link_2"].Id)

	// test rule match
	i.addAlert("Device down", mockAlerts["bgp_1"])
	i.addAlert("Device down", mockAlerts["link_1"])
	i.checkRule(ctx, rule, out)
	assert.Equal(t, mockAlerts["bgp_1"].Status, models.Status_SUPPRESSED)
	assert.Equal(t, mockAlerts["link_1"].Status, models.Status_SUPPRESSED)
	assert.Equal(t, len(i.alertBuf["Device down"]), 0)
}

func TestMain(m *testing.M) {
	flag.Parse()
	ah.Config = ah.NewConfigHandler("../../../testutil/testdata/test_config.yaml")
	ah.Config.LoadConfig()
	os.Exit(m.Run())
}
