package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/streadway/amqp"
	"time"
)

const (
	exchangeName = "alerts"
	exchangeType = "direct"
	routingKey   = "alerts"
)

type IncidentInfo struct {
	Description string
	Entity      string
	Device      string
	Site        string
	Owner, Team string
	Tags        []string
	StartTime   time.Time `json:"start_time"`
	Severity    string
	Status      string
}

type Incident struct {
	Name    string
	Id      int64
	Infos   []IncidentInfo
	AddedAt time.Time `json:"added_at"`
}

type Publisher struct {
	AmqpAddr      string        `mapstructure:"amqp_addr"`
	AmqpUser      string        `mapstructure:"amqp_username"`
	AmqpPass      string        `mapstructure:"amqp_password"`
	AmqpQueueName string        `mapstructure:"amqp_queue_name"`
	ConnectRetry  time.Duration `mapstructure:"connect_retry"`
	ready         bool
	Notif         chan *ah.AlertEvent
	channel       *amqp.Channel
}

func (p *Publisher) Name() string {
	return "rabbitmq"
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

func (p *Publisher) toIncident(event *ah.AlertEvent) *Incident {
	alert := event.Alert
	return &Incident{
		Name: alert.Name,
		Id:   alert.Id,
		Infos: []IncidentInfo{
			IncidentInfo{
				Description: alert.Description,
				Entity:      alert.Entity,
				Device:      alert.Device.String,
				Site:        alert.Site.String,
				Owner:       alert.Owner.String,
				Team:        alert.Team.String,
				Tags:        alert.Tags,
				StartTime:   alert.StartTime.Time,
				Severity:    alert.Severity.String(),
				Status:      alert.Status.String(),
			},
		},
		AddedAt: time.Now(),
	}
}

func (p *Publisher) Start(ctx context.Context) {
	if !p.ready {
		go func() {
			if err := p.Setup(); err != nil {
				glog.Errorf("Failed to start amqp publisher: %v", err)
			}
		}()
	}
	for {
		select {
		case event := <-p.Notif:
			if !p.ready {
				glog.V(2).Infof("Amqp publisher not ready, skipping event")
				break
			}
			if event.Type != ah.EventType_ACTIVE {
				break
			}
			//TODO dont publish repeat notifications
			if err := p.Publish(p.toIncident(event)); err != nil {
				glog.Errorf("RabbitMq: Failed to publish incident: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	p := &Publisher{}
	ah.RegisterOutput(p.Name(), p.Notif)
	am.AddOutput(p)
}
