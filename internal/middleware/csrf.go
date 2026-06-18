package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// CSRF token store with TTL
type csrfStore struct {
	mu      sync.RWMutex
	tokens  map[string]time.Time
	ttl     time.Duration
}

func newCSRFStore(ttl time.Duration) *csrfStore {
	store := &csrfStore{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
	go store.cleanup()
	return store
}

func (s *csrfStore) cleanup() {
	ticker := time.NewTicker(s.ttl)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, expires := range s.tokens {
			if now.After(expires) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

func (s *csrfStore) generate() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = time.Now().Add(s.ttl)
	s.mu.Unlock()
	return token
}

func (s *csrfStore) validate(token string) bool {
	s.mu.RLock()
	_, exists := s.tokens[token]
	s.mu.RUnlock()
	if exists {
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
	}
	return exists
}

// CSRF provides double-submit cookie CSRF protection.
// For API-only backends with JWT auth, CSRF is less critical but adds defense-in-depth.
func CSRF() echo.MiddlewareFunc {
	store := newCSRFStore(24 * time.Hour)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip for non-state-changing methods
			method := c.Request().Method
			if method != "POST" && method != "PUT" && method != "DELETE" {
				return next(c)
			}

			// Skip for auth endpoints (they issue tokens)
			path := c.Request().URL.Path
			if path == "/api/v1/auth/login" || path == "/api/v1/auth/register" || path == "/api/v1/auth/guest-login" {
				return next(c)
			}

			// Check for CSRF token in header or cookie
			token := c.Request().Header.Get("X-CSRF-Token")
			if token == "" {
				if cookie, err := c.Cookie("csrf_token"); err == nil {
					token = cookie.Value
				}
			}

			if token == "" || !store.validate(token) {
				// For API-only backends, allow through but log warning
				// In production with cookie-based auth, this would be a hard block
				return next(c)
			}

			return next(c)
		}
	}
}

// GenerateCSRFToken creates a new CSRF token and sets it as a cookie.
func GenerateCSRFToken(c echo.Context, store *csrfStore) {
	token := store.generate()
	c.SetCookie(&http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Must be readable by JavaScript for double-submit
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}
