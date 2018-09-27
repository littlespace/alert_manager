package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	ah "github.com/mayuresh82/alert_manager/handler"
	"github.com/mayuresh82/alert_manager/internal/models"
	"github.com/mayuresh82/alert_manager/internal/stats"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const key = "al3rtMana63r"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JwtToken struct {
	Token string `json:"token"`
}

func buildSelectQuery(queries map[string][]string) (models.Query, error) {
	query := models.Query{}
	for q, v := range queries {
		switch q {
		case "limit":
			query.Limit, _ = strconv.Atoi(v[0])
		case "offset":
			query.Offset, _ = strconv.Atoi(v[0])
		default:
			if strings.HasSuffix(q, "__in") {
				parts := strings.Split(q, "__")
				q = parts[0]
				v = strings.Split(v[0], ",")
			}
			op := models.Op_EQUAL
			if len(v) > 1 {
				op = models.Op_IN
			}
			query.Params = append(query.Params, models.Param{Field: q, Values: v, Op: op})
		}
	}
	return query, nil
}

func buildUpdateQuery(queries, matches map[string][]string) (models.UpdateQuery, error) {
	query := models.UpdateQuery{}
	for q, v := range queries {
		if len(v) > 1 {
			return query, fmt.Errorf("Error: Invalid query params")
		}
		query.Set = append(query.Set, models.Field{Name: q, Value: v[0]})
	}
	for m, v := range matches {
		op := models.Op_EQUAL
		if len(v) > 1 {
			op = models.Op_IN
		}
		query.Where = append(query.Where, models.Param{Field: m, Values: v, Op: op})
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
	router.HandleFunc("/api/alerts", s.GetAlerts).Methods("GET")
	router.HandleFunc("/api/alerts/{id}", s.GetAlert).Methods("GET")
	router.HandleFunc("/api/alerts/{id}", s.Validate(s.UpdateAlert)).Methods("PATCH")
	router.HandleFunc("/api/alerts/{id}/{action}", s.Validate(s.ActionAlert)).Methods("PATCH")

	// CORS specific headers
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH"})

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

func (s *Server) fetchResults(q models.Querier) (models.Alerts, error) {
	tx := s.handler.Db.NewTx()
	ctx := context.Background()
	var (
		alerts models.Alerts
		er     error
	)
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		alerts, er = q.Run(tx)
		if er != nil {
			return er
		}
		return nil
	})
	return alerts, err
}

func (s *Server) GetAlerts(w http.ResponseWriter, req *http.Request) {
	queries := req.URL.Query()
	q, err := buildSelectQuery(queries)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch alerts: %s", err.Error()), http.StatusBadRequest)
	}
	alerts, err := s.fetchResults(q)
	if err != nil {
		glog.Errorf("Api: Unable to fetch alerts: %v", err)
		http.Error(w, fmt.Sprintf("Unable to fetch alerts: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	s.statGets.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
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

func (s *Server) UpdateAlert(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	queries := req.URL.Query()
	q, err := buildUpdateQuery(queries, map[string][]string{"id": []string{vars["id"]}})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to Update alerts: %s", err.Error()), http.StatusBadRequest)
		return
	}
	_, err = s.fetchResults(q)
	if err != nil {
		glog.Errorf("Api: Unable to Update alerts: %v", err)
		http.Error(w, fmt.Sprintf("Unable to Update alerts: %s", err.Error()), http.StatusInternalServerError)
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
			rule := models.NewSuppRule(
				models.Labels{"alert_id": alert.Id},
				"alert",
				"Alert suppressed via API",
				"alert manager",
				duration)
			err = s.handler.Suppress(ctx, tx, alert, rule)
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
