package controllers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ───────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name  string `json:"name"     binding:"required,min=2,max=100" example:"John Doe"`
	Email string `json:"email"    binding:"required,email"         example:"john@school.com"`

	Password string `json:"password" binding:"required,min=8"         example:"secret123"`
	// Valid roles: Admin, Teacher, Student, Parent
	Role  string `json:"role" binding:"required,oneof=Admin Teacher Student Parent" example:"Student"`
	Phone string `json:"phone" example:"09xxxxxxxx"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email" example:"john@school.com"`
	Password string `json:"password" binding:"required"       example:"secret123"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"      example:"oldpass123"`
	NewPassword     string `json:"new_password"     binding:"required,min=8" example:"newpass456"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGci..."`
}

// ══════════════════════════════════════════════════════
//  REGISTER
// ══════════════════════════════════════════════════════

// Register godoc
// @Summary      Register a new user (Admin only)
// @Description  Creates a new user account. Requires Admin JWT.
// @Description  Valid roles: Admin, Teacher, Student, Parent.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterInput        true  "Registration data"
// @Success      201   {object}  helpers.APIResponse  "User created"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Failure      401   {object}  helpers.APIResponse  "Unauthorized"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      409   {object}  helpers.APIResponse  "Email already registered"
// @Router       /api/admin/register [post]
// func Register(c *gin.Context) {
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	var existing models.User
	if result := config.DB.Where("email = ?", input.Email).First(&existing); result.Error == nil {
		helpers.Error(c, http.StatusConflict, "email already registered")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role,
		Phone:    input.Phone,
		IsActive: true,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	helpers.Success(c, http.StatusCreated, "user registered successfully", gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// ══════════════════════════════════════════════════════
//  LOGIN
// ══════════════════════════════════════════════════════

// Login godoc
// @Summary      Login and get JWT tokens
// @Description  Authenticates a user and returns an access token (1 hour) and a rotated refresh token (7 days).
// @Description  Use the access_token in the Authorization header: `Bearer <token>`.
// @Description  The refresh token is also set as an HttpOnly cookie (`sms_refresh`) for browser clients.
// @Description  When the access_token expires, call POST /api/token/refresh.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginInput           true  "Login credentials"
// @Success      200   {object}  helpers.APIResponse  "Tokens returned"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Failure      401   {object}  helpers.APIResponse  "Invalid credentials"
// @Router       /api/login [post]
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		log.Printf("[login] User not found: %s (error: %v)", input.Email, err)
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	log.Printf("[login] Found user: %s, active: %v", user.Email, user.IsActive)

	if !user.IsActive {
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		log.Printf("[login] Password mismatch for: %s", input.Email)
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	accessToken, err := config.GenerateAccessToken(user.ID, user.Role, user.Email)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	// generate a new refresh token and persist it so we can validate/revoke it.
	refreshToken, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// SHA-256 the JWT first — bcrypt rejects inputs >72 bytes (golang.org/x/crypto v0.17+)
	rtSum := sha256.Sum256([]byte(refreshToken))
	hashedRT, err := bcrypt.GenerateFromPassword(rtSum[:], 10)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to store refresh token")
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Use a transaction to ensure atomicity
	tx := config.DB.Begin()
	if tx.Error != nil {
		helpers.Error(c, http.StatusInternalServerError, "database error")
		return
	}

	// Delete previous tokens for this user (allows multiple sessions)
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{}).Error; err != nil {
		tx.Rollback()
		helpers.Error(c, http.StatusInternalServerError, "failed to clear old tokens")
		return
	}

	rt := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: string(hashedRT),
		ExpiresAt: expiresAt,
	}
	if err := tx.Create(&rt).Error; err != nil {
		tx.Rollback()
		helpers.Error(c, http.StatusInternalServerError, "failed to persist refresh token")
		return
	}

	if err := tx.Commit().Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	// HttpOnly cookies for browser clients — never readable from JavaScript.
	helpers.SetAccessCookie(c, accessToken)
	helpers.SetRefreshCookie(c, refreshToken)

	helpers.Success(c, http.StatusOK, "login successful", gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600, // 1 hour in seconds
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		// refresh_token also returned in body for API/mobile clients that cannot use cookies.
		"refresh_token": refreshToken,
	})
}

// ══════════════════════════════════════════════════════
//  REFRESH TOKEN
// ══════════════════════════════════════════════════════

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access token AND a new refresh token (rotation).
// @Description  The old refresh token is invalidated on use (prevents replay attacks).
// @Description  Token is read from the `sms_refresh` HttpOnly cookie (browser) or the JSON body (API clients).
// @Description  The new refresh token is returned in both the response body and a new cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RefreshTokenInput    false "Refresh token (API clients only — browser uses cookie)"
// @Success      200   {object}  helpers.APIResponse  "New tokens issued"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Failure      401   {object}  helpers.APIResponse  "Invalid or expired refresh token"
// @Router       /api/token/refresh [post]
func RefreshToken(c *gin.Context) {
	// Accept token from cookie (browser) or body (API/mobile clients).
	tokenStr, _ := c.Cookie("sms_refresh")
	if tokenStr == "" {
		var input RefreshTokenInput
		if err := c.ShouldBindJSON(&input); err != nil {
			helpers.Error(c, http.StatusBadRequest, "refresh_token required in body or sms_refresh cookie")
			return
		}
		tokenStr = input.RefreshToken
	}

	// Parse and validate the JWT signature/expiry.
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
		},
	)
	if err != nil || !token.Valid {
		helpers.Error(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "invalid token claims")
		return
	}

	var userID uint
	if _, err := fmt.Sscanf(claims.Subject, "%d", &userID); err != nil {
		helpers.Error(c, http.StatusUnauthorized, "invalid token subject")
		return
	}

	// validate against the stored hash (prevents replay of a previously used token).
	// Search all active tokens for this user to find the one matching this hash
	// (supports multiple concurrent sessions, e.g. multiple browser tabs).
	tokenSum := sha256.Sum256([]byte(tokenStr))
	var allActive []models.RefreshToken
	if err := config.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Find(&allActive).Error; err != nil {
		helpers.Error(c, http.StatusUnauthorized, "refresh token not found or expired — please log in again")
		return
	}
	var stored models.RefreshToken
	var matched bool
	for _, rt := range allActive {
		if bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), tokenSum[:]) == nil {
			stored = rt
			matched = true
			break
		}
	}
	if !matched {
		helpers.Error(c, http.StatusUnauthorized, "refresh token mismatch — please log in again")
		return
	}

	// Verify the user still exists and is active.
	var user models.User
	if err := config.DB.Where("id = ? AND is_active = true AND deleted_at IS NULL", userID).
		First(&user).Error; err != nil {
		helpers.Error(c, http.StatusUnauthorized, "account not found or deactivated")
		return
	}

	// issue a NEW refresh token (rotation) and invalidate the old one.
	newAccessToken, err := config.GenerateAccessToken(user.ID, user.Role, user.Email)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	newRefreshToken, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	newRTSum := sha256.Sum256([]byte(newRefreshToken))
	newHashedRT, err := bcrypt.GenerateFromPassword(newRTSum[:], 10)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	// Replace old record atomically.
	newExpiry := time.Now().Add(7 * 24 * time.Hour)
	if err := config.DB.Model(&stored).Updates(map[string]any{
		"token_hash": string(newHashedRT),
		"expires_at": newExpiry,
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to rotate refresh token")
		return
	}

	helpers.SetAccessCookie(c, newAccessToken)
	helpers.SetRefreshCookie(c, newRefreshToken)

	helpers.Success(c, http.StatusOK, "token refreshed", gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
	})
}

// ══════════════════════════════════════════════════════
//  LOGOUT
// ══════════════════════════════════════════════════════

// Logout godoc
// @Summary      Logout — revoke refresh token
// @Description  Deletes the stored refresh token for this user and clears the sms_refresh cookie.
// @Description  After logout, the access token remains valid until it expires (max 1 hour).
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "Logged out"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Router       /api/logout [post]
func Logout(c *gin.Context) {
	userID := c.GetUint("userID")
	// Revoke all refresh tokens for this user.
	config.DB.Where("user_id = ?", userID).Delete(&models.RefreshToken{})
	helpers.ClearAuthCookies(c)
	helpers.Success(c, http.StatusOK, "logged out successfully", nil)
}

// ══════════════════════════════════════════════════════
//  SSE TOKEN — short-lived token for EventSource
// ══════════════════════════════════════════════════════

// IssueSSEToken godoc
// @Summary      Issue a short-lived SSE token
// @Description  Returns a 60-second token for use in the EventSource URL (?sse_token=...).
// @Description  This avoids passing the long-lived access JWT in the query string (which would appear in logs).
// @Description  The token is signed with a separate SSE secret and only valid for the /notifications/stream endpoint.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "SSE token"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Router       /api/notifications/sse-token [post]
func IssueSSEToken(c *gin.Context) {
	userID := c.GetUint("userID")
	role := c.GetString("role")
	email := c.GetString("email")

	sseToken, err := config.GenerateSSEToken(userID, role, email)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate SSE token")
		return
	}

	helpers.Success(c, http.StatusOK, "SSE token issued", gin.H{
		"sse_token":  sseToken,
		"expires_in": 60,
	})
}

// ══════════════════════════════════════════════════════
//  GET ME
// ══════════════════════════════════════════════════════

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the currently authenticated user.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "User profile"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Failure      404  {object}  helpers.APIResponse  "User not found"
// @Router       /api/me [get]
func GetMe(c *gin.Context) {
	userID := c.GetUint("userID")
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "user not found")
		return
	}
	helpers.Success(c, http.StatusOK, "user profile", gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"phone":      user.Phone,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
	})
}

// GetPortalContext returns active academic year and semester for the portal dashboard.
func GetPortalContext(c *gin.Context) {
	var academicYear int
	config.DB.Model(&models.Class{}).
		Where("status = ?", "Active").
		Select("COALESCE(MAX(year), 0)").
		Scan(&academicYear)
	if academicYear == 0 {
		config.DB.Model(&models.Student{}).
			Select("COALESCE(MAX(academic_year), 0)").
			Scan(&academicYear)
	}
	if academicYear == 0 {
		academicYear = time.Now().Year()
	}

	month := time.Now().Month()
	activeSemester := "Semester 1"
	switch {
	case month >= time.September || month <= time.January:
		activeSemester = "Semester 1"
	case month >= time.February && month <= time.May:
		activeSemester = "Semester 2"
	default:
		activeSemester = "Semester 3"
	}

	helpers.Success(c, http.StatusOK, "portal context", gin.H{
		"academic_year":   academicYear,
		"active_semester": activeSemester,
	})
}

// ══════════════════════════════════════════════════════
//  CHANGE PASSWORD
// ══════════════════════════════════════════════════════

// ChangePassword godoc
// @Summary      Change password
// @Description  Updates the authenticated user's password. Requires the current password for verification.
// @Description  Also revokes all existing refresh tokens (forces re-login on other devices).
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      ChangePasswordInput  true  "Password change data"
// @Success      200   {object}  helpers.APIResponse  "Password updated"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Failure      401   {object}  helpers.APIResponse  "Current password incorrect"
// @Failure      404   {object}  helpers.APIResponse  "User not found"
// @Router       /api/me/password [put]
func ChangePassword(c *gin.Context) {
	userID := c.GetUint("userID")
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "user not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		helpers.Error(c, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := config.DB.Model(&user).Updates(map[string]any{
		"password":   string(hashed),
		"updated_at": time.Now(),
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	// revoke all refresh tokens after password change — forces re-login on all devices.
	config.DB.Where("user_id = ?", userID).Delete(&models.RefreshToken{})
	helpers.ClearAuthCookies(c)

	helpers.Success(c, http.StatusOK, "password updated successfully — please log in again", nil)
}