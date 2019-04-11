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

const (
	tokenExpiryTime = 24 * time.Hour
)

type AuthProvider interface {
	Authenticate(userid, password string) (bool, error)
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JwtToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
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
	addr         string
	handler      *ah.AlertHandler
	authProvider AuthProvider
	apiKey       string

	statGets          stats.Stat
	statPosts         stats.Stat
	statPatches       stats.Stat
	statError         stats.Stat
	statsAuthFailures stats.Stat
}

func NewServer(addr, apiKey string, authProvider AuthProvider, handler *ah.AlertHandler) *Server {
	return &Server{
		addr:              addr,
		apiKey:            apiKey,
		handler:           handler,
		authProvider:      authProvider,
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
	router.HandleFunc("/api/auth/refresh", s.Validate(s.RefreshToken)).Methods("GET")
	router.HandleFunc("/api/plugins", s.GetPluginsList).Methods("GET")
	router.HandleFunc("/api/{category}", s.GetItems).Methods("GET")
	router.HandleFunc("/api/{category}/{id}", s.Validate(s.Update)).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/alerts/{id}", s.GetAlert).Methods("GET")
	router.HandleFunc("/api/alerts/{id}/{action}", s.Validate(s.ActionAlert)).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/suppression_rules/persistent", s.GetPersistentRules).Methods("GET")
	router.HandleFunc("/api/suppression_rules", s.Validate(s.CreateSuppRule)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/suppression_rules/{id}/clear", s.Validate(s.ClearSuppRule)).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/users", s.Validate(s.CreateUser)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/users/{name}/delete", s.Validate(s.DeleteUser)).Methods("DELETE", "OPTIONS")

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
				claims := &Claims{}
				token, err := jwt.ParseWithClaims(bearerToken[1], claims, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Error validating token")
					}
					return []byte(s.apiKey), nil
				})
				if err != nil {
					s.statError.Add(1)
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				if token.Valid {
					glog.V(4).Infof("Validated token for user %s", claims.Username)
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
		// try to get basic auth
		u, p, ok := req.BasicAuth()
		if !ok {
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}
		user.Username = u
		user.Password = p
	} else {
		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}
	}
	if s.authProvider != nil {
		auth, err := s.authProvider.Authenticate(user.Username, user.Password)
		if err != nil {
			glog.Errorf("Failed to auth %s: %v", user.Username, err)
			http.Error(w, fmt.Sprintf("Authentication Failed: %v", err), http.StatusUnauthorized)
			return
		}
		if !auth {
			http.Error(w, "Authentication Failed", http.StatusUnauthorized)
			return
		}
	}
	expirationTime := time.Now().Add(tokenExpiryTime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime,
		},
	})
	tokenString, err := token.SignedString([]byte(s.apiKey))
	if err != nil {
		s.statsAuthFailures.Add(1)
		glog.Errorf("Api: Error generating Token: %v", err)
		http.Error(w, fmt.Sprintf("Error generating Token: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	glog.V(2).Infof("Successfully authenticated user: %s", user.Username)
	// TODO save the user /team to db
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString, ExpiresAt: expirationTime})
}

func (s *Server) RefreshToken(w http.ResponseWriter, req *http.Request) {
	decoded := req.Context().Value("decoded")
	claims := decoded.(*Claims)
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		http.Error(w, "Renewal only allowed within 30 seconds of expiry", http.StatusBadRequest)
		return
	}
	claims.ExpiresAt = time.Now().Add(tokenExpiryTime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.apiKey))
	if err != nil {
		s.statsAuthFailures.Add(1)
		glog.Errorf("Api: Error generating Token: %v", err)
		http.Error(w, fmt.Sprintf("Error generating Token: %s", err.Error()), http.StatusInternalServerError)
		s.statError.Add(1)
		return
	}
	glog.V(2).Infof("Successfully renewed claims for user: %s", claims.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString, ExpiresAt: claims.ExpiresAt})
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
	// hack for users until user mgmt is implemented
	vars := mux.Vars(req)
	if vars["category"] == "users" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.handler.GetUsersFromConfig())
		return
	}
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
		var er error
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
			er = s.handler.Suppress(
				ctx, tx, alert,
				creator,
				reason,
				duration,
			)
		case "clear":
			er = s.handler.Clear(ctx, tx, alert)
		case "ack":
			owner, ok := queries["owner"]
			if !ok {
				http.Error(w, "Invalid query: expected non empty owner", http.StatusBadRequest)
				return fmt.Errorf("Invalid query: expected non empty owner")
			}
			team := ""
			teams, ok := queries["team"]
			if ok {
				team = teams[0]
			}
			er = s.handler.SetOwner(ctx, tx, alert, owner[0], team)
		}
		if er != nil {
			glog.Errorf("Api: Unable to Action alerts: %v", er)
			http.Error(w, fmt.Sprintf("Unable to Update alerts: %s", er.Error()), http.StatusInternalServerError)
		}
		return er
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
	if err := json.NewDecoder(req.Body).Decode(rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid parameters for query: %v", err), http.StatusBadRequest)
		return
	}
	if rule.Name == "" {
		rule.Name = fmt.Sprintf("Rule - %s - %v", rule.Creator, rule.Duration)
	}
	rule.CreatedAt = models.MyTime{time.Now()}
	tx := s.handler.Db.NewTx()
	ctx := req.Context()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		id, err := s.handler.AddSuppRule(ctx, tx, rule)
		if err != nil {
			return err
		}
		rule.Id = id
		return nil
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create suppression rule: %v", err), http.StatusInternalServerError)
		return
	}
	s.statPosts.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func (s *Server) ClearSuppRule(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	tx := s.handler.Db.NewTx()
	ctx := req.Context()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		return s.handler.DeleteSuppRule(ctx, tx, id)
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to delete suppression rule: %v", err), http.StatusBadRequest)
		return
	}
}

func (s *Server) GetPluginsList(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins.GetApiPluginsList())
}

func (s *Server) GetPersistentRules(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.handler.Suppressor.GetPersistentRules())
}

func (s *Server) CreateUser(w http.ResponseWriter, req *http.Request) {
	u := &models.User{}
	if err := json.NewDecoder(req.Body).Decode(u); err != nil {
		http.Error(w, fmt.Sprintf("Invalid parameters for query: %v", err), http.StatusBadRequest)
		return
	}
	if u.Team == nil {
		http.Error(w, fmt.Sprintf("Failed to add user: Team is required"), http.StatusBadRequest)
		return
	}
	tx := s.handler.Db.NewTx()
	ctx := req.Context()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		id, err := tx.NewInsert(models.QueryInsertTeam, u.Team)
		if err != nil {
			return fmt.Errorf("Failed to create new team: %v", err)
		}
		u.TeamId = id
		u.Team.Id = id
		id, err = tx.NewInsert(models.QueryInsertUser, u)
		if err != nil {
			return fmt.Errorf("Failed to create new user: %v", err)
		}
		u.Id = id
		return nil
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}
	s.statPosts.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (s *Server) DeleteUser(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tx := s.handler.Db.NewTx()
	ctx := req.Context()
	err := models.WithTx(ctx, tx, func(ctx context.Context, tx models.Txn) error {
		return tx.Exec(models.QueryDeleteUserByName, vars["name"])
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to delete user: %v", err), http.StatusBadRequest)
		return
	}
}
