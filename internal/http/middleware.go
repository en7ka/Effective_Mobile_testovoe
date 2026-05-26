package http

import (
	"log"
	nethttp "net/http"
	"sync"
	"time"
)

type statusRecorder struct {
	nethttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(logger *log.Logger) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			start := time.Now()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         nethttp.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			logger.Printf(
				"request method=%s path=%s status=%d duration=%s",
				r.Method,
				r.URL.Path,
				recorder.status,
				time.Since(start),
			)
		})
	}
}

func rateLimitMiddleware(requestsPerSecond int, burst int) func(nethttp.Handler) nethttp.Handler {
	limiter := newTokenBucket(requestsPerSecond, burst)

	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			if !limiter.allow() {
				writeError(w, nethttp.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type tokenBucket struct {
	mu               sync.Mutex
	tokens           float64
	capacity         float64
	refillPerSecond  float64
	lastRefill       time.Time
}

func newTokenBucket(requestsPerSecond int, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:          float64(burst),
		capacity:        float64(burst),
		refillPerSecond: float64(requestsPerSecond),
		lastRefill:      time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = minFloat(b.capacity, b.tokens+elapsed*b.refillPerSecond)
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}
