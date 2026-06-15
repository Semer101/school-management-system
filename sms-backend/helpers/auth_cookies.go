package helpers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AccessCookieName  = "sms_access"
	RefreshCookieName = "sms_refresh"
	accessMaxAge      = 3600 // 1 hour
	refreshMaxAge     = 7 * 24 * 60 * 60
)

func cookieSecure() bool {
	return os.Getenv("ENV") == "production"
}

func cookieSameSite() http.SameSite {
	if os.Getenv("ENV") == "production" && os.Getenv("FRONTEND_URL") != "" {
		return http.SameSiteNoneMode
	}
	return http.SameSiteStrictMode
}

func SetAccessCookie(c *gin.Context, accessToken string) {
	c.SetSameSite(cookieSameSite())
	c.SetCookie(AccessCookieName, accessToken, accessMaxAge, "/", "", cookieSecure(), true)
}

func SetRefreshCookie(c *gin.Context, refreshToken string) {
	c.SetSameSite(cookieSameSite())
	// Path must be "/api/token" so the cookie is sent to /api/token/refresh
	// and other paths under /api/ will not send it (security best practice)
	c.SetCookie(RefreshCookieName, refreshToken, refreshMaxAge, "/api/token", "", cookieSecure(), true)
}

// ClearAuthCookies removes session cookies on logout.
func ClearAuthCookies(c *gin.Context) {
	secure := cookieSecure()
	c.SetSameSite(cookieSameSite())
	c.SetCookie(AccessCookieName, "", -1, "/", "", secure, true)
	c.SetCookie(RefreshCookieName, "", -1, "/api/token", "", secure, true)
}

// ExtractAccessToken reads JWT from Authorization header or sms_access cookie.
func ExtractAccessToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") && parts[1] != "" {
			return parts[1], true
		}
	}
	if token, err := c.Cookie(AccessCookieName); err == nil && token != "" {
		return token, true
	}
	return "", false
}
