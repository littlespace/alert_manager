package listener

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
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
	Labels  map[string]interface{}
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

func (k *WebHookListener) sanityCheck(d *WebHookAlertData) error {
	// alert name should only contain alpha-numeric chars, spaces or underscores
	reg, err := regexp.Compile("[^a-zA-Z0-9_\\s]+")
	if err != nil {
		return err
	}
	processedString := reg.ReplaceAllString(d.Name, "")
	if processedString != d.Name {
		return fmt.Errorf("Invalid alert name: %s. Name should only contain alpha-numeric chars, spaces or underscores", d.Name)
	}
	return nil
}

func (k *WebHookListener) formatAlertEvent(d *WebHookAlertData, team string) (*models.AlertEvent, error) {
	if err := k.sanityCheck(d); err != nil {
		return nil, err
	}
	// check if the alert exists in the definition
	defined, ok := ah.Config.GetAlertConfig(d.Name)
	var scope string
	if ok {
		if defined.Config.Severity != "" {
			d.Level = defined.Config.Severity
		}
		scope = defined.Config.Scope
	}
	if d.Level == "" {
		d.Level = "INFO"
	}
	event := &models.AlertEvent{}
	event.Alert = models.NewAlert(d.Name, d.Details, d.Entity, d.Source, scope, team, d.Id, d.Time, d.Level, false)
	if d.Device != "" {
		event.Alert.AddDevice(d.Device)
	}
	switch d.Status {
	case Status_CLEARED:
		event.Type = models.EventType_CLEARED
	default:
		event.Type = models.EventType_ACTIVE
	}
	if d.Labels != nil {
		for k, v := range d.Labels {
			event.Alert.Labels[k] = v
		}
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
		for k, v := range defined.Config.StaticLabels {
			event.Alert.Labels[k] = v
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
	teams, ok := queries["team"]
	team := "default"
	if ok {
		team = teams[0]
	} else {
		glog.V(2).Infof("No team specified in URL, using 'default'")
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
		glog.Error(err)
		k.statRequestsError.Add(1)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	event, err := k.formatAlertEvent(data, team)
	if err != nil {
		glog.Error(err)
		k.statRequestsError.Add(1)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ah.ListenChan <- event
}

func (k *WebHookListener) Name() string {
	return "webhook"
}

func (k *WebHookListener) GetParsersList() []string {

	var plist []string
	for _, p := range parsers {
		plist = append(plist, p.Name())
	}
	return plist
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
	srv := &http.Server{Addr: k.ListenAddr, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
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
	plugins.AddListener(listener)
}
