package output

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-mail/mail"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	tpl "github.com/mayuresh82/alert_manager/template"
	"html/template"
	"net"
	"strconv"
	"time"
)

type Emailer interface {
	send(addr, username, pwd, from, subject, body string, recipients []string) error
}

type EmailSender struct{}

func (e *EmailSender) send(addr, username, pwd, from, subject, body string, recipients []string) error {
	m := mail.NewMessage()
	m.SetAddressHeader("From", from, "Alert Manager")
	m.SetHeader("To", recipients...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if port == "" {
		port = "25"
	}
	p, _ := strconv.Atoi(port)
	d := mail.NewDialer(host, p, username, pwd)
	if p == 587 {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

type EmailNotifier struct {
	Notif        chan *ah.AlertEvent
	rawTpl       string
	Emailer      Emailer
	SmtpAddr     string `mapstructure:"smtp_addr"`
	UseAuth      bool   `mapstructure:"use_auth"`
	SmtpUsername string `mapstructure:"smtp_username"`
	SmtpPassword string `mapstructure:"smtp_password"`
	From         string
	Recipients   []string
}

type TplData struct {
	Subject       string
	AlertMgrURL   string
	SentAt        string
	EventType     string
	AlertSeverity string
	Header        string
	AlertParams   []struct{ Name, Value string }
}

func (e *EmailNotifier) Name() string {
	return "email"
}

func (e *EmailNotifier) renderTemplate(data *TplData) (string, error) {
	tmpl, err := template.New("email").Parse(e.rawTpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *EmailNotifier) subject(event *ah.AlertEvent) string {
	alert := event.Alert
	var subject string
	if alert.Device.Valid {
		subject = fmt.Sprintf("Alert Manager: [%s] %s: [%s][%s]", alert.Status.String(), alert.Name, alert.Device.String, alert.Entity)
	} else {
		subject = fmt.Sprintf("Alert Manager: [%s] %s: [%s]", alert.Status.String(), alert.Name, alert.Entity)
	}
	return subject
}

func (e *EmailNotifier) start(event *ah.AlertEvent) {
	data := &TplData{
		Subject:       e.subject(event),
		AlertMgrURL:   "http://TODO",
		SentAt:        time.Now().Format("Mon Jan 2 15:04:05 MST 2006"),
		EventType:     event.Type.String(),
		AlertSeverity: event.Alert.Severity.String(),
		AlertParams: []struct{ Name, Value string }{
			struct{ Name, Value string }{"Name", event.Alert.Name},
			struct{ Name, Value string }{"Description", event.Alert.Description},
			struct{ Name, Value string }{"Entity", event.Alert.Entity},
			struct{ Name, Value string }{"StartTime", event.Alert.StartTime.String()},
		},
	}
	data.Header = fmt.Sprintf("[%s][%s] %s", data.AlertSeverity, data.EventType, event.Alert.Name)
	if event.Alert.Device.Valid {
		data.AlertParams = append(data.AlertParams, struct{ Name, Value string }{"Device", event.Alert.Device.String})
	}
	body, err := e.renderTemplate(data)
	if err != nil {
		glog.Errorf("Output: Email: Failed to render template: %v", err)
		return
	}
	if err := e.Emailer.send(
		e.SmtpAddr,
		e.SmtpUsername,
		e.SmtpPassword,
		e.From,
		data.Subject,
		body,
		e.Recipients); err != nil {
		glog.Errorf("Output: Email : Unable to send email : %v", err)
	}
}

func (e *EmailNotifier) Start(ctx context.Context) {
	for {
		select {
		case event := <-e.Notif:
			e.start(event)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	e := &EmailNotifier{
		Notif:   make(chan *ah.AlertEvent),
		rawTpl:  tpl.EmailTemplate,
		Emailer: &EmailSender{},
	}
	ah.RegisterOutput(e.Name(), e.Notif)
	am.AddOutput(e)
}
