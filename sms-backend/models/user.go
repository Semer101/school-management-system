package models

import (
	"time"

	"gorm.io/gorm"
)

// Role constants — use these everywhere, never raw strings
const (
	RoleAdmin   = "Admin"
	RoleTeacher = "Teacher"
	RoleStudent = "Student"
	RoleParent  = "Parent"
)

// User covers FE-01, FE-02, FE-03
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // json:"-" means NEVER sent in responses
	Role      string         `gorm:"not null;default:Student" json:"role"`
	Phone     string         `json:"phone"`
	AvatarURL string         `json:"avatar_url"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // soft delete
}

// RefreshToken stores the hashed refresh token for each user session.
// enables token rotation, revocation, and replay-attack prevention.
// One row per user — replaced on each login or refresh.
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"` // one active token per user
	TokenHash string    `gorm:"not null" json:"-"`                   // bcrypt hash — never exposed
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
