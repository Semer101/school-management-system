package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"sms-backend/helpers"

	"github.com/gin-gonic/gin"
)

// ══════════════════════════════════════════════════════
//  SECURITY MIDDLEWARE  (OWASP Top 10 — FE-22)
// ══════════════════════════════════════════════════════

// SecurityHeaders adds HTTP security headers to every response (OWASP A05)
// These headers tell browsers how to protect users from common attacks
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking — stops your site being embedded in an iframe
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing — browsers must respect Content-Type header
		c.Header("X-Content-Type-Options", "nosniff")

		// Force HTTPS for 1 year (only enable in production)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Control what information is sent in the Referer header
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy — restricts where scripts/styles can load from.
		// FIX: Added 'unsafe-inline' to script-src (needed for the inline <script> in swagger-ui.html)
		//      Added connect-src to allow the Swagger UI to make fetch() calls back to the API.
		c.Header("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self' http://localhost:8080")

		// Hide server information to make it harder for attackers to identify vulnerabilities
		c.Writer.Header().Del("X-Powered-By")

		c.Next()
	}
}

// CORSMiddleware controls which origins can call your API (OWASP A05)
// CORS = Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	// In production, replace with your actual frontend URL
	allowedOrigins := []string{
		"http://localhost:3000", // React dev server
		"http://localhost:5173", // Vite dev server
		"http://localhost:8080", 
		"http://127.0.0.1:5500",
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if the request origin is in our allowed list
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// For disallowed origins, we can either omit the header or explicitly deny
			c.Writer.Header().Del("Access-Control-Allow-Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours preflight cache

		// Handle preflight OPTIONS requests — browsers send these before actual requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ── Rate Limiting (OWASP A07 — Authentication failure protection) ─────────────

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time // IP -> list of request timestamps
	limit    int                    // max requests
	window   time.Duration          // time window
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Clean up old entries every 5 minutes
	go func() {
		for range time.Tick(5 * time.Minute) {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter out requests outside the window
	var recent []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.requests[ip] = recent
		return false
	}

	rl.requests[ip] = append(recent, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for ip, times := range rl.requests {
		var recent []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = recent
		}
	}
}

// Global rate limiters for different endpoint types
var (
	loginLimiter = newRateLimiter(5, time.Minute)   // 5 login attempts/minute
	apiLimiter   = newRateLimiter(100, time.Minute) // 100 API calls/minute
)

// RateLimitLogin restricts login attempts to prevent brute force (OWASP A07)
func RateLimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !loginLimiter.allow(ip) {
			helpers.Error(c, http.StatusTooManyRequests, "too many login attempts, try again in 1 minute")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitAPI applies general rate limiting to all API endpoints
func RateLimitAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !apiLimiter.allow(ip) {
			helpers.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ── Input Sanitization (OWASP A03 — Injection prevention) ────────────────────

// SanitizeInput strips potentially dangerous characters from string inputs
// This is a defence-in-depth measure — GORM parameterized queries are the primary defence
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	// Trim excessive whitespace
	input = strings.TrimSpace(input)
	return input
}

// ── Request Logger ─────────────────────────────────────────────────────────────

// RequestLogger logs method, path, status code, and duration for every request
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		c.Next() // process the request

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		userID, _ := c.Get("userID")

		// Log format: [METHOD] /path | STATUS | duration | IP | userID
		gin.DefaultWriter.Write([]byte(
			strings.Join([]string{
				"[LOG]",
				method,
				path,
				http.StatusText(statusCode),
				duration.String(),
				ip,
				"user:" + formatUserID(userID),
				"\n",
			}, " | "),
		))
	}
}

func formatUserID(userID any) string {
	if userID == nil {
		return "anonymous"
	}
	return strings.TrimSpace(strings.Replace(strings.Replace(
		strings.Replace(fmt.Sprintf("%v", userID), " ", "", -1), "\n", "", -1), "\r", "", -1))
}
