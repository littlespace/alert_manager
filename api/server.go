package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"github.com/mayuresh82/alert_manager/plugins"
)

const key = "al3rtMana63r"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JwtToken struct {
	Token string `json:"token"`
}

func buildSelectQuery(req *http.Request) (models.Query, error) {
	queries := req.URL.Query()
	vars := mux.Vars(req)
	query := models.NewQuery(vars["category"])
	for q, v := range queries {
		switch q {
		case "limit":
			query.Limit, _ = strconv.Atoi(v[0])
		case "offset":
			query.Offset, _ = strconv.Atoi(v[0])
		case "timerange":
			query.TimeRange = v[0]
		case "history":
			query.IncludeHistory = true
		default:
			if strings.HasSuffix(q, "__in") {
				parts := strings.Split(q, "__")
				q = parts[0]
				v = strings.Split(v[0], ",")
			}
			query.Params = append(query.Params, models.Param{Field: q, Values: v})
		}
	}
	return query, nil
}

func buildUpdateQuery(req *http.Request, matches map[string][]string) (models.UpdateQuery, error) {
	vars := mux.Vars(req)
	queries := req.URL.Query()
	query := models.NewUpdateQuery(vars["category"])
	for q, v := range queries {
		if len(v) > 1 {
			return query, fmt.Errorf("Error: Invalid query params")
		}
		query.Set = append(query.Set, models.Field{Name: q, Value: v[0]})
	}
	for m, v := range matches {
		query.Where = append(query.Where, models.Param{Field: m, Values: v})
	}
	return query, nil
}

type Server struct {
	addr    string
	handler *ah.AlertHandler

	statGets          stats.Stat
	statPosts         stats.Stat
	statPatches       stats.Stat
	statError         stats.Stat
	statsAuthFailures stats.Stat
}

func NewServer(addr string, handler *ah.AlertHandler) *Server {
	return &Server{
		addr:              addr,
		handler:           handler,
		statGets:          stats.NewCounter("api.gets"),
		statPosts:         stats.NewCounter("api.posts"),
		statPatches:       stats.NewCounter("api.patches"),
		statError:         stats.NewCounter("api.errors"),
		statsAuthFailures: stats.NewCounter("api.auth_failures"),
	}
}

func (s *Server) Start(ctx context.Context) {
	router := mux.NewRouter()

	router.HandleFunc("/api/auth", s.CreateToken).Methods("POST")
	router.HandleFunc("/api/plugins", s.GetPluginsList).Methods("GET")
	router.HandleFunc("/api/{category}", s.GetItems).Methods("GET")
	router.HandleFunc("/api/{category}/{id}", s.Validate(s.Update)).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/alerts/{id}", s.GetAlert).Methods("GET")
	router.HandleFunc("/api/alerts/{id}/{action}", s.Validate(s.ActionAlert)).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/suppression_rules", s.Validate(s.CreateSuppRule)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/suppression_rules/{id}/clear", s.Validate(s.ClearSuppRule)).Methods("DELETE", "OPTIONS")

	// CORS specific headers
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"})

	// set up the router
	srv := &http.Server{
		Handler: handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router),
		Addr:    s.addr,
		// set some sane timeouts
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	srv.ListenAndServe()
}

func (s *Server) Validate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Error validating token")
					}
					return []byte(key), nil
				})
				if err != nil {
					s.statError.Add(1)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if token.Valid {
					ctx := context.WithValue(req.Context(), "decoded", token.Claims)
					next(w, req.WithContext(ctx))
				} else {
					s.statsAuthFailures.Add(1)
					http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
				}
			}
		} else {
			s.statError.Add(1)
			http.Error(w, "An authorization header is required", http.StatusBadRequest)
		}
	})
}

func (s *Server) CreateToken(w http.ResponseWriter, req *http.Request) {
	var user User
	if req.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	//TODO Authenticate user provided creds against AD

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		s.statsAuthFailures.Add(1)
		glog.Errorf("Api: Error generating Token: %v", err)
		http.Error(w, fmt.Sprintf("Error generating Token: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	glog.V(2).Infof("Successfully authenticated user: %s", user.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

func (s *Server) fetchResults(q models.Querier) ([]interface{}, error) {
	tx := s.handler.Db.NewTx()
	ctx := context.Background()
	var (
		items []interface{}
		er    error
	)
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		items, er = q.Run(tx)
		if er != nil {
			return er
		}
		return nil
	})
	return items, err
}

func (s *Server) GetItems(w http.ResponseWriter, req *http.Request) {
	q, err := buildSelectQuery(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch items: %s", err.Error()), http.StatusBadRequest)
	}
	items, err := s.fetchResults(q)
	if err != nil {
		glog.Errorf("Api: Unable to fetch items: %v", err)
		http.Error(w, fmt.Sprintf("Unable to fetch items: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	s.statGets.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (s *Server) GetAlert(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	a := &models.Alert{Id: id}
	tx := s.handler.Db.NewTx()
	err := models.WithTx(req.Context(), tx, func(ctx context.Context, tx models.Txn) error {
		alert, err := s.handler.GetExisting(tx, a)
		if err != nil {
			return err
		}
		s.statError.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alert)
		return nil
	})
	if err != nil {
		glog.Errorf("Api: Unable to fetch alerts: %v", err)
		http.Error(w, fmt.Sprintf("Unable to fetch alerts: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
	}
}

func (s *Server) Update(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	q, err := buildUpdateQuery(req, map[string][]string{"id": []string{vars["id"]}})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to Update item: %s", err.Error()), http.StatusBadRequest)
		return
	}
	_, err = s.fetchResults(q)
	if err != nil {
		glog.Errorf("Api: Unable to Update item: %v", err)
		http.Error(w, fmt.Sprintf("Unable to Update item: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	s.statPatches.Add(1)
}

func (s *Server) ActionAlert(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	a := &models.Alert{Id: id}
	tx := s.handler.Db.NewTx()
	alert, err := s.handler.GetExisting(tx, a)
	if err != nil {
		http.Error(w, fmt.Sprintf("Alert %d not found", id), http.StatusBadRequest)
		s.statError.Add(1)
		return
	}
	ctx := req.Context()
	err = models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		queries := req.URL.Query()
		var err error
		switch vars["action"] {
		case "suppress":
			if alert.Status != models.Status_ACTIVE {
				http.Error(w, fmt.Sprintf("Alert %d is not ACTIVE", id), http.StatusBadRequest)
				return fmt.Errorf("Invalid query: Alert %d is not ACTIVE", id)
			}
			durationStr, ok := queries["duration"]
			if !ok {
				http.Error(w, "Invalid query: expected duration", http.StatusBadRequest)
				return fmt.Errorf("Invalid query: expected duration")
			}
			duration, err := time.ParseDuration(durationStr[0])
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid duration: %s", err.Error()), http.StatusBadRequest)
				return fmt.Errorf("Invalid duration: %s", err.Error())
			}
			creator := "alert_manager"
			reason := "alert suppressed via API"
			err = s.handler.Suppress(
				ctx, tx, alert,
				creator,
				reason,
				duration,
			)
		case "clear":
			err = s.handler.Clear(ctx, tx, alert)
		case "ack":
			owner, ok := queries["owner"]
			team, ok := queries["team"]
			if !ok {
				http.Error(w, "Invalid query: expected owner and team", http.StatusBadRequest)
				return fmt.Errorf("Invalid query: expected owner and team")
			}
			err = s.handler.SetOwner(ctx, tx, alert, owner[0], team[0])
		}
		if err != nil {
			glog.Errorf("Api: Unable to Action alerts: %v", err)
			http.Error(w, fmt.Sprintf("Unable to Update alerts: %s", err.Error()), http.StatusInternalServerError)
		}
		return err
	})
	if err != nil {
		return
	}
	s.statPatches.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

func (s *Server) CreateSuppRule(w http.ResponseWriter, req *http.Request) {
	rule := &models.SuppressionRule{}
	if err := json.NewDecoder(req.Body).Decode(&rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid parameters for query: %v", err), http.StatusBadRequest)
		return
	}
	if rule.Name == "" {
		rule.Name = fmt.Sprintf("Rule - %s - %v", rule.Creator, rule.Duration)
	}
	rule.CreatedAt = models.MyTime{time.Now()}
	tx := s.handler.Db.NewTx()
	id, err := s.handler.AddSuppRule(req.Context(), tx, rule)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create suppression rule: %v", err), http.StatusInternalServerError)
		return
	}
	rule.Id = id
	s.statPosts.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func (s *Server) ClearSuppRule(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	tx := s.handler.Db.NewTx()
	if err := s.handler.DeleteSuppRule(req.Context(), tx, id); err != nil {
		http.Error(w, fmt.Sprintf("Unable to delete suppression rule: %v", err), http.StatusBadRequest)
		return
	}
}

func (s *Server) GetPluginsList(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins.GetApiPluginsList())
}
