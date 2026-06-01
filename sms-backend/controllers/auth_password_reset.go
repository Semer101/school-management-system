package controllers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPInput struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// ForgotPassword sends a 6-digit OTP to the user's email
func ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))

	var user models.User
	if err := config.DB.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		// Don't reveal whether email exists
		helpers.Success(c, http.StatusOK, "if the email exists, an OTP has been sent", nil)
		return
	}

	otp := generateOTP()
	hash, _ := bcrypt.GenerateFromPassword([]byte(otp), 10)

	// Invalidate previous OTPs
	config.DB.Where("email = ?", email).Delete(&models.PasswordResetOTP{})

	record := models.PasswordResetOTP{
		Email:     email,
		OTPHash:   string(hash),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	config.DB.Create(&record)

	body := fmt.Sprintf(`<h2>Password Reset</h2><p>Your OTP code is: <strong>%s</strong></p><p>Valid for 15 minutes.</p>`, otp)
	helpers.SendEmail(email, "SMS: Password Reset OTP", body)

	helpers.Success(c, http.StatusOK, "if the email exists, an OTP has been sent", nil)
}

// ResetPasswordWithOTP verifies OTP and sets new password
func ResetPasswordWithOTP(c *gin.Context) {
	var input VerifyOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))

	var record models.PasswordResetOTP
	if err := config.DB.Where("email = ? AND used = ? AND expires_at > ?", email, false, time.Now()).
		Order("created_at DESC").First(&record).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid or expired OTP")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.OTPHash), []byte(input.OTP)); err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid OTP")
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "user not found")
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	config.DB.Model(&user).Update("password", string(hashed))
	config.DB.Model(&record).Update("used", true)
	config.DB.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})

	helpers.Success(c, http.StatusOK, "password reset successfully", nil)
}

// UploadAvatar handles profile picture upload
func UploadAvatar(c *gin.Context) {
	userID := c.GetUint("userID")
	file, err := c.FormFile("avatar")
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "avatar file required")
		return
	}

	avatarDir := os.Getenv("AVATAR_DIR")
	if avatarDir == "" {
		avatarDir = "./uploads/avatars"
	}
	os.MkdirAll(avatarDir, 0750)

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		helpers.Error(c, http.StatusBadRequest, "only jpg, png, webp allowed")
		return
	}

	filename := fmt.Sprintf("user_%d%s", userID, ext)
	dest := filepath.Join(avatarDir, filename)
	if err := c.SaveUploadedFile(file, dest); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "upload failed")
		return
	}

	url := "/uploads/avatars/" + filename
	config.DB.Model(&models.User{}).Where("id = ?", userID).Update("avatar_url", url)

	helpers.Success(c, http.StatusOK, "avatar updated", gin.H{"avatar_url": url})
}
