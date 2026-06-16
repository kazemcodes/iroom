package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	reqs := rl.requests[key]
	valid := make([]time.Time, 0, len(reqs))
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

func RateLimit(limit int, window time.Duration) echo.MiddlewareFunc {
	limiter := newRateLimiter(limit, window)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			if !limiter.allow(key) {
				return response.Error(c, http.StatusTooManyRequests, "درخواست‌ها بیش از حد مجاز است")
			}
			return next(c)
		}
	}
}

func AuthRateLimit() echo.MiddlewareFunc {
	return RateLimit(20, time.Minute)
}

func APIKeyRateLimit() echo.MiddlewareFunc {
	return RateLimit(60, time.Minute)
}
