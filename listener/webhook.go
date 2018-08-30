package listener

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	am "github.com/mayuresh82/alert_manager"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	Status_ALERTING = "alerting"
	Status_CLEARED  = "cleared"
)

type WebHookAlertData struct {
	Id      string
	Name    string
	Details string
	Device  string
	Entity  string
	Time    time.Time
	Level   string
	Status  string
	Source  string
}

type WebHookListener struct {
	ListenAddr         string `mapstructure:"listen_addr"`
	UseAuth            bool   `mapstructure:"use_auth"`
	Username, Password string

	statRequestsRecvd stats.Stat
	statRequestsError stats.Stat
	statsAuthFailures stats.Stat
}

func NewWebHookListener() *WebHookListener {
	return &WebHookListener{
		statRequestsRecvd: stats.NewCounter("listener.webhook.requests_recvd"),
		statRequestsError: stats.NewCounter("listener.webhook.requests_err"),
		statsAuthFailures: stats.NewCounter("listener.webhook.auth_failures"),
	}
}

func (k *WebHookListener) formatAlertEvent(d *WebHookAlertData) (*ah.AlertEvent, error) {
	// check if the alert exists in the definition
	defined, ok := ah.Config.GetAlertConfig(d.Name)
	var scope string
	if ok {
		d.Level = defined.Config.Severity
		scope = defined.Config.Scope
	}
	event := &ah.AlertEvent{}
	event.Alert = models.NewAlert(d.Name, d.Details, d.Entity, d.Source, scope, d.Id, d.Time, d.Level, false)
	event.Alert.AddDevice(d.Device)
	switch d.Status {
	case Status_CLEARED:
		event.Type = ah.EventType_CLEARED
	default:
		event.Type = ah.EventType_ACTIVE
	}
	if ok {
		if len(defined.Config.Tags) > 0 {
			event.Alert.AddTags(defined.Config.Tags...)
		}
		if defined.Config.AutoExpire != nil && *defined.Config.AutoExpire {
			event.Alert.SetAutoExpire(defined.Config.ExpireAfter)
		}
		if defined.Config.AutoClear != nil {
			event.Alert.AutoClear = *defined.Config.AutoClear
		}
	}
	return event, nil
}

func (k *WebHookListener) basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer h.ServeHTTP(w, r)
		if !k.UseAuth {
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		username, password, authOK := r.BasicAuth()
		if !authOK {
			k.statsAuthFailures.Add(1)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		if username != k.Username || password != k.Password {
			k.statsAuthFailures.Add(1)
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
	}
}

func (k WebHookListener) pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I-AM-ALIVE")
}

func (k WebHookListener) httpHandler(w http.ResponseWriter, r *http.Request) {
	k.statRequestsRecvd.Add(1)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	glog.V(4).Infof("New alert create request: %v", string(body))

	queries := r.URL.Query()
	source, ok := queries["source"]
	if !ok || len(source) != 1 {
		glog.Errorf("No query found in URL: %v", r.URL)
		k.statRequestsError.Add(1)
		http.Error(w, "No query found in URL", http.StatusBadRequest)
		return
	}
	var parser Parser
	for _, p := range parsers {
		if p.Name() == source[0] {
			parser = p
			break
		}
	}
	if parser == nil {
		k.statRequestsError.Add(1)
		glog.Errorf("No parser found in alert definition")
		http.Error(w, "No parser found in alert definition", http.StatusInternalServerError)
		return
	}

	data, err := parser.Parse(body)
	if err != nil {
		k.statRequestsError.Add(1)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	event, err := k.formatAlertEvent(data)
	if err != nil {
		k.statRequestsError.Add(1)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ah.ListenChan <- event
}

func (k *WebHookListener) Name() string {
	return "webhook"
}

func (k *WebHookListener) Uri() string {
	addr := k.ListenAddr
	if strings.HasPrefix(k.ListenAddr, ":") {
		addr = "localhost" + k.ListenAddr
	}
	return fmt.Sprintf("http://%s/listener/alert/", addr)
}

func (k WebHookListener) Listen(ctx context.Context) {
	http.HandleFunc("/listener/alert/", k.basicAuth(k.httpHandler))
	http.HandleFunc("/listener/ping/", k.pingHandler)
	srv := &http.Server{Addr: k.ListenAddr}
	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			glog.Errorf("Webhook HTTP server Shutdown Error: %v", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener
		glog.Errorf("HTTP server ListenAndServe Error: %v", err)
	}
	<-idleConnsClosed
}

func init() {
	listener := NewWebHookListener()
	am.AddListener(listener)
}
