package output

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/plugins"
	"github.com/streadway/amqp"
)

const (
	exchangeName = "alerts"
	exchangeType = "direct"
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
	AmqpRoutingKey string        `mapstructure:"amqp_routing_key"`
	ConnectRetry  time.Duration `mapstructure:"connect_retry"`
	ready         bool
	Notif         chan *plugins.SendRequest
	channel       *amqp.Channel

	sync.RWMutex
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
		p.AmqpRoutingKey,   // routing to 0 or more queues
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
	p.RLock()
	if p.ready {
		return nil
	}
	p.RUnlock()
	var conn *amqp.Connection
	var err error

	conn, err = amqp.Dial(p.uri())
	if err != nil {
		glog.V(2).Infof("Rabbitmq: Error connecting to server %s: %v", p.AmqpAddr, err)
		go func() {
			time.Sleep(p.ConnectRetry)
			p.Setup()
		}()
		return err
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
	p.Lock()
	p.ready = true
	p.Unlock()
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
		case req := <-p.Notif:
			p.RLock()
			if !p.ready {
				glog.V(2).Infof("Amqp publisher not ready, skipping event")
				p.RUnlock()
				break
			}
			p.RUnlock()
			if req.Event.Type != models.EventType_ACTIVE && req.Event.Type != models.EventType_CLEARED {
				break
			}
			//TODO dont publish repeat notifications
			if err := p.Publish(p.toIncident(req.Event)); err != nil {
				glog.Errorf("Amqp: Failed to publish incident: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	p := &Publisher{ConnectRetry: 60 * time.Second, Notif: make(chan *plugins.SendRequest)}
	plugins.AddOutput(p, p.Notif)
}
