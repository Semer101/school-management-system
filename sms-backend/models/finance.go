package models

import "time"

// Transaction tracks all financial movements (FE-20, FE-21)
//
// A receipt can be uploaded in two ways:
//  1. By a Parent/Student via the JSON API (`SubmitBankReceipt`) — uses just a
//     bank-supplied `ReceiptID` text. Image is optional.
//  2. By a Parent via the receipt-upload API (`UploadPaymentReceipt`) —
//     uploads a JPG/JPEG/PNG/WEBP image of the bank receipt to
//     `ReceiptImageURL` and links it to the student via `StudentID`.
//
// On the wire, `CreatedBy` is the userID of the uploader:
//   - Parent upload → parent's UserID
//   - Student upload → student's UserID
//
// Admin moderation:
//   - `Status`   = "Pending" | "Verified" | "Rejected"
//   - `VerifiedBy` / `VerifiedAt` = admin that took the action
//   - `RejectionNotes` = admin-only field populated on Reject for the parent
type Transaction struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	StudentID       uint       `json:"student_id"`
	Student         Student    `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	ParentID        *uint      `json:"parent_id"`        // pointer = nullable; set when a parent uploads
	Amount          float64    `gorm:"not null" json:"amount"`
	ReceiptID       string     `gorm:"uniqueIndex" json:"receipt_id"` // Ethiopian bank receipt Tx ID
	ReceiptImageURL string     `json:"receipt_image_url"`            // path relative to /uploads, served by ServeReceiptImage
	Type            string     `gorm:"not null" json:"type"`         // "Tuition", "Expense", "Payroll"
	Status          string     `gorm:"default:Pending" json:"status"` // "Pending", "Verified", "Rejected"
	Description     string     `json:"description"`
	RejectionNotes  string     `json:"rejection_notes"` // free-text admin note on Reject
	CreatedBy       uint       `json:"created_by"`      // UserID of uploader (parent or student)
	VerifiedBy      uint       `json:"verified_by"`     // Admin UserID who verified/rejected
	VerifiedAt      *time.Time `json:"verified_at"`     // pointer = nullable
	AcademicYear    int        `json:"academic_year"`
	Semester        string     `json:"semester"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
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
