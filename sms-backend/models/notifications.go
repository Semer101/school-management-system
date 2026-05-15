package models

import "time"

// Notification stores a broadcast message sent by an Admin.
// target_roles is a comma-separated list e.g. "Student,Parent"
// An empty target_roles means the notification targets all roles.
type Notification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Body        string    `gorm:"not null" json:"body"`
	TargetRoles string    `gorm:"not null;default:''" json:"target_roles"` // comma-separated e.g. "Student,Parent"
	SenderID    uint      `gorm:"not null" json:"sender_id"`
	Sender      User      `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationReceipt tracks per-user read state for each notification.
// A receipt row is created for every (user, notification) pair when a
// broadcast is saved, so we can track is_read per user.
type NotificationReceipt struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UserID         uint         `gorm:"not null;index" json:"user_id"`
	User           User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	NotificationID uint         `gorm:"not null;index" json:"notification_id"`
	Notification   Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
	IsRead         bool         `gorm:"default:false" json:"is_read"`
	ReadAt         *time.Time   `json:"read_at"`
}
