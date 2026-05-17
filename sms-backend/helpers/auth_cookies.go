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

// SetAccessCookie stores the JWT in an HttpOnly cookie (browser clients).
func SetAccessCookie(c *gin.Context, accessToken string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(AccessCookieName, accessToken, accessMaxAge, "/", "", cookieSecure(), true)
}

// SetRefreshCookie stores the refresh token (scoped to refresh path).
func SetRefreshCookie(c *gin.Context, refreshToken string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(RefreshCookieName, refreshToken, refreshMaxAge, "/api/token", "", cookieSecure(), true)
}

// ClearAuthCookies removes session cookies on logout.
func ClearAuthCookies(c *gin.Context) {
	secure := cookieSecure()
	c.SetSameSite(http.SameSiteStrictMode)
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
