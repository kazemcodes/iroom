package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iroom/iroom/internal/config"
	"github.com/iroom/iroom/internal/database"
	"github.com/iroom/iroom/internal/handlers"
	"github.com/iroom/iroom/internal/middleware"
	"github.com/iroom/iroom/internal/pkg/hash"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type testEnv struct {
	e              *echo.Echo
	api            *echo.Group
	cfg            *config.Config
	token          string
	db             *sql.DB
	userRepo       *repository.UserRepo
	classRepo      *repository.ClassRepo
	sessionRepo    *repository.SessionRepo
	messageRepo    *repository.MessageRepo
	recordingRepo  *repository.RecordingRepo
	logRepo        *repository.ActivityLogRepo
	settingsRepo   *repository.SettingsRepo
	ticketRepo     *repository.TicketRepo
	sessionLogRepo *repository.SessionLogRepo
	resetRepo      *repository.PasswordResetRepo
	healthHandler  *handlers.HealthHandler
	totpSvc        *services.TOTPService
}

func setup(t *testing.T) *testEnv {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.Seed(db)

	cfg := config.Default()
	cfg.External.APIKey = "test-api-key"
	e := echo.New()
	e.HideBanner = true

	api := e.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWT.Secret))

	token, _ := jwt.Generate(cfg.JWT.Secret, jwt.Claims{UserID: 1, Email: "admin@iroom.local", Role: "admin"}, cfg.JWT.AccessExpiry)

	livekitSvc := services.NewLiveKitService("test-key", "test-secret", "")
	healthHandler := handlers.NewHealthHandler(db, ":memory:", livekitSvc)

	resetRepo := repository.NewPasswordResetRepo(db)
	totpSvc := services.NewTOTPService("IRoom")

	return &testEnv{
		e:              e,
		api:            api,
		cfg:            cfg,
		token:          token,
		db:             db,
		userRepo:       repository.NewUserRepo(db),
		classRepo:      repository.NewClassRepo(db),
		sessionRepo:    repository.NewSessionRepo(db),
		messageRepo:    repository.NewMessageRepo(db),
		recordingRepo:  repository.NewRecordingRepo(db),
		logRepo:        repository.NewActivityLogRepo(db),
		settingsRepo:   repository.NewSettingsRepo(db),
		ticketRepo:     repository.NewTicketRepo(db),
		sessionLogRepo: repository.NewSessionLogRepo(db),
		resetRepo:      resetRepo,
		healthHandler:  healthHandler,
		totpSvc:        totpSvc,
	}
}

func req(e *echo.Echo, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	r := httptest.NewRequest(method, path, buf)
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w
}

func jsonBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &m)
	return m
}

// ==================== AUTH ====================

func TestRegister(t *testing.T) {
	env := setup(t)
	h := handlers.NewAuthHandler(env.userRepo, env.logRepo, env.resetRepo, "", env.cfg.JWT.Secret, env.cfg.JWT.AccessExpiry, env.cfg.JWT.RefreshExpiry, env.totpSvc)
	env.e.POST("/api/v1/auth/register", h.Register)

	w := req(env.e, "POST", "/api/v1/auth/register", map[string]string{
		"email": "new@test.com", "password": "pass123", "display_name": "کاربر جدید",
	}, "")
	if w.Code != 201 {
		t.Errorf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin(t *testing.T) {
	env := setup(t)
	h := handlers.NewAuthHandler(env.userRepo, env.logRepo, env.resetRepo, "", env.cfg.JWT.Secret, env.cfg.JWT.AccessExpiry, env.cfg.JWT.RefreshExpiry, env.totpSvc)
	env.e.POST("/api/v1/auth/login", h.Login)

	w := req(env.e, "POST", "/api/v1/auth/login", map[string]string{
		"email": "admin@iroom.local", "password": "admin123",
	}, "")
	if w.Code != 200 {
		t.Errorf("login: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := jsonBody(t, w)
	if !resp["success"].(bool) {
		t.Error("login should succeed")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	env := setup(t)
	h := handlers.NewAuthHandler(env.userRepo, env.logRepo, env.resetRepo, "", env.cfg.JWT.Secret, env.cfg.JWT.AccessExpiry, env.cfg.JWT.RefreshExpiry, env.totpSvc)
	env.e.POST("/api/v1/auth/login", h.Login)

	w := req(env.e, "POST", "/api/v1/auth/login", map[string]string{
		"email": "admin@iroom.local", "password": "wrong",
	}, "")
	if w.Code != 401 {
		t.Errorf("wrong pass: expected 401, got %d", w.Code)
	}
}

func TestRefresh(t *testing.T) {
	env := setup(t)
	h := handlers.NewAuthHandler(env.userRepo, env.logRepo, env.resetRepo, "", env.cfg.JWT.Secret, env.cfg.JWT.AccessExpiry, env.cfg.JWT.RefreshExpiry, env.totpSvc)
	env.e.POST("/api/v1/auth/login", h.Login)
	env.e.POST("/api/v1/auth/refresh", h.Refresh)

	loginResp := jsonBody(t, req(env.e, "POST", "/api/v1/auth/login", map[string]string{
		"email": "admin@iroom.local", "password": "admin123",
	}, ""))
	tokens := loginResp["data"].(map[string]interface{})["tokens"].(map[string]interface{})

	w := req(env.e, "POST", "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tokens["refresh_token"].(string),
	}, "")
	if w.Code != 200 {
		t.Errorf("refresh: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ==================== CLASSES ====================

func TestClassCRUD(t *testing.T) {
	env := setup(t)
	ch := handlers.NewClassHandler(env.classRepo, env.sessionRepo)
	env.api.POST("/classes", ch.Create)
	env.api.GET("/classes", ch.List)
	env.api.GET("/classes/:id", ch.GetByID)
	env.api.PUT("/classes/:id", ch.Update)
	env.api.DELETE("/classes/:id", ch.Delete)
	env.api.POST("/classes/:id/enroll", ch.Enroll)
	env.api.GET("/classes/:id/students", ch.GetStudents)

	// Create
	w := req(env.e, "POST", "/api/v1/classes", map[string]interface{}{
		"name": "ریاضی", "description": "پایه دهم", "color": "#3B82F6", "max_students": 30,
	}, env.token)
	if w.Code != 201 {
		t.Fatalf("create class: %d: %s", w.Code, w.Body.String())
	}

	// List
	w = req(env.e, "GET", "/api/v1/classes", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("list classes: %d", w.Code)
	}

	// Get by ID
	w = req(env.e, "GET", "/api/v1/classes/1", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("get class: %d: %s", w.Code, w.Body.String())
	}

	// Get not found
	w = req(env.e, "GET", "/api/v1/classes/999", nil, env.token)
	if w.Code != 404 {
		t.Errorf("get class not found: expected 404, got %d", w.Code)
	}

	// Update
	w = req(env.e, "PUT", "/api/v1/classes/1", map[string]interface{}{
		"name": "ریاضی پیشرفته",
	}, env.token)
	if w.Code != 200 {
		t.Fatalf("update class: %d: %s", w.Code, w.Body.String())
	}

	// Enroll (use admin user ID 1 since that's the only user)
	w = req(env.e, "POST", "/api/v1/classes/1/enroll", map[string]interface{}{
		"student_id": 1,
	}, env.token)
	if w.Code != 200 {
		t.Fatalf("enroll: %d: %s", w.Code, w.Body.String())
	}

	// Students
	w = req(env.e, "GET", "/api/v1/classes/1/students", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("students: %d: %s", w.Code, w.Body.String())
	}

	// Delete
	w = req(env.e, "DELETE", "/api/v1/classes/1", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("delete class: %d: %s", w.Code, w.Body.String())
	}

	// No auth
	w = req(env.e, "POST", "/api/v1/classes", map[string]interface{}{"name": "test"}, "")
	if w.Code != 401 {
		t.Errorf("no auth: expected 401, got %d", w.Code)
	}
}

// ==================== SESSIONS ====================

func TestSessionCRUD(t *testing.T) {
	env := setup(t)
	ch := handlers.NewClassHandler(env.classRepo, env.sessionRepo)
	sh := handlers.NewSessionHandler(env.sessionRepo, env.classRepo)
	env.api.POST("/classes", ch.Create)
	env.api.POST("/sessions", sh.Create)
	env.api.GET("/sessions", sh.List)
	env.api.GET("/sessions/:id", sh.GetByID)
	env.api.POST("/sessions/:id/start", sh.Start)
	env.api.POST("/sessions/:id/end", sh.End)
	env.api.DELETE("/sessions/:id", sh.Delete)

	// Create class first
	req(env.e, "POST", "/api/v1/classes", map[string]interface{}{"name": "فیزیک"}, env.token)

	// Create session
	w := req(env.e, "POST", "/api/v1/sessions", map[string]interface{}{
		"class_id": 1, "title": "جلسه اول", "scheduled_at": "2026-07-01T10:00:00Z", "duration": 60,
	}, env.token)
	if w.Code != 201 {
		t.Fatalf("create session: %d: %s", w.Code, w.Body.String())
	}

	// Get session
	w = req(env.e, "GET", "/api/v1/sessions/1", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("get session: %d: %s", w.Code, w.Body.String())
	}

	// Start session
	w = req(env.e, "POST", "/api/v1/sessions/1/start", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("start session: %d: %s", w.Code, w.Body.String())
	}
	resp := jsonBody(t, w)
	status := resp["data"].(map[string]interface{})["status"]
	if status != "live" {
		t.Errorf("expected status=live, got %v", status)
	}

	// End session
	w = req(env.e, "POST", "/api/v1/sessions/1/end", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("end session: %d: %s", w.Code, w.Body.String())
	}

	// Delete
	w = req(env.e, "DELETE", "/api/v1/sessions/1", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("delete session: %d: %s", w.Code, w.Body.String())
	}
}

// ==================== MESSAGES ====================

func TestMessages(t *testing.T) {
	env := setup(t)
	ch := handlers.NewClassHandler(env.classRepo, env.sessionRepo)
	sh := handlers.NewSessionHandler(env.sessionRepo, env.classRepo)
	mh := handlers.NewMessageHandler(env.messageRepo)
	env.api.POST("/classes", ch.Create)
	env.api.POST("/sessions", sh.Create)
	env.api.POST("/sessions/:id/messages", mh.Send)
	env.api.GET("/sessions/:id/messages", mh.List)

	req(env.e, "POST", "/api/v1/classes", map[string]interface{}{"name": "تست"}, env.token)
	req(env.e, "POST", "/api/v1/sessions", map[string]interface{}{
		"class_id": 1, "title": "جلسه تست", "scheduled_at": "2026-07-01T10:00:00Z",
	}, env.token)

	// Send
	w := req(env.e, "POST", "/api/v1/sessions/1/messages", map[string]string{"content": "سلام دنیا"}, env.token)
	if w.Code != 201 {
		t.Fatalf("send msg: %d: %s", w.Code, w.Body.String())
	}

	// List
	w = req(env.e, "GET", "/api/v1/sessions/1/messages", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("list msgs: %d: %s", w.Code, w.Body.String())
	}
	msgs := jsonBody(t, w)["data"].([]interface{})
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

// ==================== ADMIN ====================

func TestAdmin(t *testing.T) {
	env := setup(t)
	ah := handlers.NewAdminHandler(env.userRepo, env.classRepo, env.sessionRepo, env.messageRepo, env.recordingRepo, env.logRepo, env.settingsRepo, env.ticketRepo, env.sessionLogRepo, env.cfg.JWT.Secret, env.cfg.JWT.AccessExpiry, env.cfg.JWT.RefreshExpiry)
	env.api.GET("/admin/dashboard/stats", ah.DashboardStats)
	env.api.GET("/admin/users", ah.ListUsers)
	env.api.GET("/admin/settings", ah.GetSettings)
	env.api.PUT("/admin/settings", ah.UpdateSettings)

	// Stats
	w := req(env.e, "GET", "/api/v1/admin/dashboard/stats", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("stats: %d", w.Code)
	}

	// Users
	w = req(env.e, "GET", "/api/v1/admin/users", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("users: %d", w.Code)
	}

	// Settings
	w = req(env.e, "GET", "/api/v1/admin/settings", nil, env.token)
	if w.Code != 200 {
		t.Fatalf("settings: %d", w.Code)
	}
	s := jsonBody(t, w)["data"].(map[string]interface{})
	if _, ok := s["recording_enabled"]; !ok {
		t.Error("missing recording_enabled")
	}

	// Update settings
	w = req(env.e, "PUT", "/api/v1/admin/settings", map[string]interface{}{
		"recording_enabled": false, "max_users_per_room": 50,
	}, env.token)
	if w.Code != 200 {
		t.Fatalf("update settings: %d", w.Code)
	}
}

// ==================== EXTERNAL API ====================

func TestExternalAPI(t *testing.T) {
	env := setup(t)
	eh := handlers.NewExternalHandler(env.userRepo, env.classRepo, env.sessionRepo, env.cfg.External.APIKey)

	ext := env.e.Group("/api/v1/external")
	ext.Use(middleware.APIKeyAuth(env.cfg.External.APIKey))
	ext.GET("/status", eh.Status)
	ext.GET("/stats", eh.Stats)
	ext.POST("/users", eh.CreateUser)

	// No API key
	w := req(env.e, "GET", "/api/v1/external/status", nil, "")
	if w.Code != 401 {
		t.Errorf("no key: expected 401, got %d", w.Code)
	}

	// Wrong key
	r := httptest.NewRequest("GET", "/api/v1/external/status", nil)
	r.Header.Set("X-API-Key", "wrong")
	w2 := httptest.NewRecorder()
	env.e.ServeHTTP(w2, r)
	if w2.Code != 401 {
		t.Errorf("wrong key: expected 401, got %d", w2.Code)
	}

	// Correct key
	r = httptest.NewRequest("GET", "/api/v1/external/status", nil)
	r.Header.Set("X-API-Key", env.cfg.External.APIKey)
	w3 := httptest.NewRecorder()
	env.e.ServeHTTP(w3, r)
	if w3.Code != 200 {
		t.Errorf("correct key: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}

	// Create user
	createBody, _ := json.Marshal(map[string]string{
		"email": "ext@test.com", "password": "pass123", "display_name": "خارجی",
	})
	r = httptest.NewRequest("POST", "/api/v1/external/users", bytes.NewBuffer(createBody))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", env.cfg.External.APIKey)
	w4 := httptest.NewRecorder()
	env.e.ServeHTTP(w4, r)
	if w4.Code != 201 {
		t.Errorf("create ext user: expected 201, got %d: %s", w4.Code, w4.Body.String())
	}
}

type nopCloser struct{ *bytes.Buffer }

func (n *nopCloser) Close() error { return nil }

// ==================== HEALTH ====================

func TestHealth(t *testing.T) {
	env := setup(t)
	env.e.GET("/api/v1/health", env.healthHandler.Health)
	w := req(env.e, "GET", "/api/v1/health", nil, "")
	if w.Code != 200 {
		t.Errorf("health: expected 200, got %d", w.Code)
	}
	resp := jsonBody(t, w)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}

// ==================== MIDDLEWARE ====================

func TestMiddlewareAuth(t *testing.T) {
	env := setup(t)

	g := env.e.Group("/api/v1/test-auth")
	g.Use(middleware.Auth(env.cfg.JWT.Secret))
	g.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"msg": "ok"})
	})

	t.Run("valid token", func(t *testing.T) {
		w := req(env.e, "GET", "/api/v1/test-auth/ping", nil, env.token)
		if w.Code != 200 {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no token", func(t *testing.T) {
		w := req(env.e, "GET", "/api/v1/test-auth/ping", nil, "")
		if w.Code != 401 {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		w := req(env.e, "GET", "/api/v1/test-auth/ping", nil, "not.a.valid.jwt")
		if w.Code != 401 {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken, err := jwt.Generate(env.cfg.JWT.Secret, jwt.Claims{UserID: 1, Email: "test@test.com", Role: "admin"}, -60)
		if err != nil {
			t.Fatalf("generate expired token: %v", err)
		}
		w := req(env.e, "GET", "/api/v1/test-auth/ping", nil, expiredToken)
		if w.Code != 401 {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}

func TestMiddlewareAdminOnly(t *testing.T) {
	env := setup(t)

	g := env.e.Group("/api/v1/test-admin")
	g.Use(middleware.Auth(env.cfg.JWT.Secret))
	g.Use(middleware.AdminOnly())
	g.GET("/secret", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"msg": "admin ok"})
	})

	t.Run("admin passes", func(t *testing.T) {
		w := req(env.e, "GET", "/api/v1/test-admin/secret", nil, env.token)
		if w.Code != 200 {
			t.Errorf("admin: expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("student gets 403", func(t *testing.T) {
		studentToken, err := jwt.Generate(env.cfg.JWT.Secret, jwt.Claims{UserID: 2, Email: "student@test.com", Role: "student"}, 60)
		if err != nil {
			t.Fatalf("generate student token: %v", err)
		}
		w := req(env.e, "GET", "/api/v1/test-admin/secret", nil, studentToken)
		if w.Code != 403 {
			t.Errorf("student: expected 403, got %d", w.Code)
		}
	})
}

func TestRateLimit(t *testing.T) {
	env := setup(t)

	g := env.e.Group("/api/v1/test-rl")
	g.Use(middleware.RateLimit(5, time.Minute))
	g.GET("/ok", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"msg": "ok"})
	})

	t.Run("5 requests pass", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			w := req(env.e, "GET", "/api/v1/test-rl/ok", nil, "")
			if w.Code != 200 {
				t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
			}
		}
	})

	t.Run("6th request gets 429", func(t *testing.T) {
		w := req(env.e, "GET", "/api/v1/test-rl/ok", nil, "")
		if w.Code != 429 {
			t.Errorf("expected 429, got %d", w.Code)
		}
	})

	t.Run("100 requests hit rate limit", func(t *testing.T) {
		rl := env.e.Group("/api/v1/test-rl2")
		rl.Use(middleware.RateLimit(5, time.Minute))
		rl.GET("/ok", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"msg": "ok"})
		})

		limited := 0
		for i := 0; i < 100; i++ {
			w := req(env.e, "GET", "/api/v1/test-rl2/ok", nil, "")
			if w.Code == 429 {
				limited++
			}
		}
		if limited == 0 {
			t.Error("expected some 429 responses in 100 requests")
		}
	})
}

func TestPasswordHash(t *testing.T) {
	h, err := hash.Hash("password")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	if h == "" {
		t.Fatal("Hash returned empty string")
	}

	if !hash.Check("password", h) {
		t.Error("Check with correct password should return true")
	}

	if hash.Check("wrong", h) {
		t.Error("Check with wrong password should return false")
	}
}

func TestJWTGenerate(t *testing.T) {
	secret := "test-secret"
	claims := jwt.Claims{UserID: 42, Email: "test@test.com", Role: "teacher"}

	tokenStr, err := jwt.Generate(secret, claims, 60)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("Generate returned empty string")
	}

	got, err := jwt.Validate(secret, tokenStr)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if got.UserID != 42 {
		t.Errorf("expected UserID=42, got %d", got.UserID)
	}
	if got.Email != "test@test.com" {
		t.Errorf("expected Email=test@test.com, got %s", got.Email)
	}
	if got.Role != "teacher" {
		t.Errorf("expected Role=teacher, got %s", got.Role)
	}

	_, err = jwt.Validate("wrong-secret", tokenStr)
	if err == nil {
		t.Error("Validate with wrong secret should return error")
	}
}
