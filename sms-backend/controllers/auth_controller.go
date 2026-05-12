package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ──────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name     string `json:"name"     binding:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email"    binding:"required,email"         example:"john@school.com"`
	Password string `json:"password" binding:"required,min=8"         example:"secret123"`
	Role     string `json:"role"     binding:"required,oneof=Admin Teacher Student Parent" example:"Student"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email" example:"john@school.com"`
	Password string `json:"password" binding:"required"       example:"secret123"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"      example:"oldpass123"`
	NewPassword     string `json:"new_password"     binding:"required,min=8" example:"newpass456"`
}

// ── Register ─────────────────────────────────────────────────────────────────

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account. Available roles: Admin, Teacher, Student, Parent
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterInput      true  "Registration data"
// @Success      201   {object}  helpers.APIResponse  "User created"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Failure      409   {object}  helpers.APIResponse  "Email already registered"
// @Router       /api/register [post]
func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	var existing models.User
	result := config.DB.Where("email = ?", input.Email).First(&existing)
	if result.Error == nil {
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

// ── Login ─────────────────────────────────────────────────────────────────────

// Login godoc
// @Summary      Login and get JWT tokens
// @Description  Authenticates a user and returns an access token (1 hour) and a refresh token (7 days)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginInput         true  "Login credentials"
// @Success      200   {object}  helpers.APIResponse  "Tokens returned"
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
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !user.IsActive {
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		helpers.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	accessToken, err := config.GenerateAccessToken(user.ID, user.Role, user.Email)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	helpers.Success(c, http.StatusOK, "login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 1 hour in seconds
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// ── GetMe ─────────────────────────────────────────────────────────────────────

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the currently authenticated user
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

// ── ChangePassword ────────────────────────────────────────────────────────────

// ChangePassword godoc
// @Summary      Change password
// @Description  Updates the authenticated user's password. Requires the current password.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      ChangePasswordInput  true  "Password change data"
// @Success      200   {object}  helpers.APIResponse    "Password updated"
// @Failure      400   {object}  helpers.APIResponse    "Validation error"
// @Failure      401   {object}  helpers.APIResponse    "Current password incorrect"
// @Failure      404   {object}  helpers.APIResponse    "User not found"
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

	if err := config.DB.Model(&user).Update("password", string(hashed)).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	now := time.Now()
	config.DB.Model(&user).Update("updated_at", now)

	helpers.Success(c, http.StatusOK, "password updated successfully", nil)
}
