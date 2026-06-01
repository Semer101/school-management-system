package models

import "time"

// PasswordResetOTP stores email OTP for forgot-password flow
type PasswordResetOTP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"`
	OTPHash   string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}
