package models

import "time"

// LockerFile represents a file stored in a student's Digital Locker (FE-12, FE-13, FE-14)
type LockerFile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StudentID  uint      `gorm:"not null" json:"student_id"`
	Student    Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	FileName   string    `gorm:"not null" json:"file_name"`
	FilePath   string    `gorm:"not null" json:"-"`              // internal path, never exposed
	FileSize   int64     `json:"file_size"`                      // bytes
	FileType   string    `json:"file_type"`                      // "pdf", "jpg", etc.
	Category   string    `json:"category"`                       // "Certificate", "Assignment", "Portfolio"
	IsPublic   bool      `gorm:"default:false" json:"is_public"` // teacher access (FE-14)
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
}
