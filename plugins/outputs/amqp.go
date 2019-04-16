package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

const (
	exchangeName = "alerts"
	exchangeType = "direct"
	routingKey   = "alerts"
)

type Incident struct {
	Name        string
	Type        string
	Id          int64
	StartTime   time.Time `json:"start_time"`
	Data        map[string]interface{}
	AddedAt     time.Time `json:"added_at"`
	IsAggregate bool      `json:"is_aggregate"`
}

type Publisher struct {
	AmqpAddr      string        `mapstructure:"amqp_addr"`
	AmqpUser      string        `mapstructure:"amqp_username"`
	AmqpPass      string        `mapstructure:"amqp_password"`
	AmqpQueueName string        `mapstructure:"amqp_queue_name"`
	ConnectRetry  time.Duration `mapstructure:"connect_retry"`
	ready         bool
	Notif         chan *models.AlertEvent
	channel       *amqp.Channel

	sync.Mutex
}

func (p *Publisher) Name() string {
	return "amqp"
}

func (p *Publisher) Publish(incident *Incident) error {
	data, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("Unable to marshal incident json: %v", err)
	}
	msg := amqp.Publishing{
		Timestamp:       time.Now(),
		DeliveryMode:    1, // non-persistent : TODO : make it persistent ?
		Priority:        0,
		Headers:         amqp.Table{},
		ContentType:     "application/json",
		ContentEncoding: "",
		Body:            data,
	}
	return p.channel.Publish(
		exchangeName, // publish to an exchange
		routingKey,   // routing to 0 or more queues
		false,        // mandatory
		false,        // immediate
		msg,
	)
}

func (p *Publisher) uri() string {
	if p.AmqpPass == "" {
		return fmt.Sprintf("amqp://%s", p.AmqpAddr)
	}
	if p.AmqpPass == "" {
		return fmt.Sprintf("amqp://%s@%s", p.AmqpUser, p.AmqpAddr)
	}
	return fmt.Sprintf("amqp://%s:%s@%s", p.AmqpUser, p.AmqpPass, p.AmqpAddr)
}

func (p *Publisher) Setup() error {
	p.Lock()
	defer p.Unlock()
	if p.ready {
		return nil
	}
	var conn *amqp.Connection
	var err error
	for {
		conn, err = amqp.Dial(p.uri())
		if err != nil {
			glog.V(2).Infof("Rabbimq: Error connecting to server %s: %v", p.uri, err)
			time.Sleep(p.ConnectRetry)
			continue
		}
		break
	}
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Error getting channel: %v", err)
	}
	p.channel = channel
	if err = p.channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		false,        // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("Error declaring exchange: %v", err)
	}
	p.ready = true
	return nil
}

func (p *Publisher) toIncident(event *models.AlertEvent) *Incident {
	alert := event.Alert
	incident := &Incident{
		Name:        alert.Name,
		Type:        event.Type.String(),
		Id:          alert.Id,
		StartTime:   alert.StartTime.Time,
		AddedAt:     time.Now(),
		IsAggregate: alert.IsAggregate,
	}
	data, _ := json.Marshal(alert)
	json.Unmarshal(data, &incident.Data)
	return incident
}

func (p *Publisher) Start(ctx context.Context, options *plugins.Options) {
	if err := p.Setup(); err != nil {
		glog.Errorf("Failed to start amqp publisher: %v", err)
	}
	for {
		select {
		case event := <-p.Notif:
			p.Lock()
			if !p.ready {
				glog.V(2).Infof("Amqp publisher not ready, skipping event")
				p.Unlock()
				break
			}
			p.Unlock()
			if event.Type != models.EventType_ACTIVE && event.Type != models.EventType_CLEARED {
				break
			}
			//TODO dont publish repeat notifications
			if err := p.Publish(p.toIncident(event)); err != nil {
				glog.Errorf("Amqp: Failed to publish incident: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	p := &Publisher{Notif: make(chan *models.AlertEvent)}
	ah.RegisterOutput(p.Name(), p.Notif)
	plugins.AddOutput(p)
}
