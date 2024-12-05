package x

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// RateWindow is the time window for rate limiting (1 minute)
	RateWindow = 60 * time.Second
	// MaxRequests is the maximum number of requests allowed per window
	MaxRequests = 60
	// RateLimitPrefix is the Redis key prefix for rate limiting
	RateLimitPrefix = "ratelimit:"
)

// GetRealIP extracts the real client IP from the request
// It checks X-Real-IP and X-Forwarded-For headers before falling back to RemoteAddr
func GetRealIP(r *http.Request) string {
	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For header
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For may contain multiple IPs, use the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return as-is if splitting fails
	}
	return ip
}

// RateLimitMiddleware creates a new rate limiting middleware
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP from X-Real-IP or X-Forwarded-For header, fallback to RemoteAddr
		clientIP := GetRealIP(r)
		key := fmt.Sprintf("%s%s", RateLimitPrefix, clientIP)

		// Use Redis to track request count
		count, err := RConn.Incr(key).Result()
		if err != nil {
			http.Error(w, "Rate limit error", http.StatusInternalServerError)
			return
		}

		// Set expiry on first request
		if count == 1 {
			RConn.Expire(key, RateWindow)
		}

		// Check if rate limit exceeded
		if count > MaxRequests {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", MaxRequests))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", MaxRequests))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", MaxRequests-int(count)))

		next.ServeHTTP(w, r)
	})
}
