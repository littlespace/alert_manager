package output

import (
	"context"
	"time"

	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/reporting"
	"github.com/mayuresh82/alert_manager/plugins"
)

func (n *InfluxNotifier) parseFromEvent(event *models.AlertEvent) *reporting.Datapoint {
	alert := event.Alert
	tags := map[string]string{
		"name":     alert.Name,
		"entity":   alert.Entity,
		"severity": alert.Severity.String(),
		"status":   alert.Status.String(),
	}
	if alert.Scope != "" {
		tags["scope"] = alert.Scope
	}
	if alert.Device.Valid {
		tags["device"] = alert.Device.String
	}
	if alert.Owner.Valid {
		tags["owner"] = alert.Owner.String
	}
	if alert.Team != "" {
		tags["team"] = alert.Team
	}
	if alert.IsAggregate {
		tags["is_aggregate"] = "true"
	} else {
		tags["is_aggregate"] = "false"
	}
	fields := make(map[string]interface{})
	switch event.Type {
	case models.EventType_ACTIVE:
		fields["num_active"] = 1
		switch alert.Severity.String() {
		case "CRITICAL":
			fields["num_critical"] = 1
		case "WARN":
			fields["num_warn"] = 1
		case "INFO":
			fields["num_info"] = 1
		}
	case models.EventType_SUPPRESSED:
		fields["num_suppressed"] = 1
	case models.EventType_EXPIRED:
		fields["num_expired"] = 1
	case models.EventType_CLEARED:
		fields["num_cleared"] = 1
	case models.EventType_ACKD:
		fields["num_ackd"] = 1
	case models.EventType_ESCALATED:
		fields["num_escalated"] = 1
	}
	return &reporting.Datapoint{
		Measurement: n.Measurement,
		Tags:        tags,
		Fields:      fields,
		TimeStamp:   time.Now(),
	}
}

type InfluxNotifier struct {
	Url         string
	Measurement string
	Notif       chan *models.AlertEvent
}

func (n *InfluxNotifier) Name() string {
	return "influx"
}

func (n *InfluxNotifier) Type() string {
	return "output"
}

func (n *InfluxNotifier) Start(ctx context.Context) {
	for {
		select {
		case event := <-n.Notif:
			d := n.parseFromEvent(event)
			reporting.DataChan <- d
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	n := &InfluxNotifier{Notif: make(chan *models.AlertEvent)}
	ah.RegisterOutput(n.Name(), n.Notif)
	plugins.AddOutput(n)
}
