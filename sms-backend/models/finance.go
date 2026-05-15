package models

import "time"

// Transaction tracks all financial movements (FE-20, FE-21)
type Transaction struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	StudentID   uint       `json:"student_id"`
	Student     Student    `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Amount      float64    `gorm:"not null" json:"amount"`
	ReceiptID   string     `gorm:"uniqueIndex" json:"receipt_id"` // Ethiopian bank receipt Tx ID
	Type        string     `gorm:"not null" json:"type"`          // "Tuition", "Expense", "Payroll"
	Status      string     `gorm:"default:Pending" json:"status"` // "Pending", "Verified", "Rejected"
	Description string     `json:"description"`
	CreatedBy   uint       `json:"created_by"`  // UserID of who created this record
	VerifiedBy  uint       `json:"verified_by"` // Admin UserID who verified
	VerifiedAt  *time.Time `json:"verified_at"` // pointer = nullable
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Payroll tracks staff salary payments (FE-20)
type Payroll struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TeacherID uint       `gorm:"not null" json:"teacher_id"`
	Teacher   Teacher    `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Amount    float64    `gorm:"not null" json:"amount"`
	Month     int        `gorm:"not null" json:"month"` // 1-12
	Year      int        `gorm:"not null" json:"year"`
	Status    string     `gorm:"default:Pending" json:"status"` // "Pending", "Paid"
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at"`
}
