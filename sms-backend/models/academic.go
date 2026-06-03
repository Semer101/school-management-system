package models

import (
	"time"

	"gorm.io/gorm"
)

// Class represents a school class/grade level (e.g. "Grade 10A")
// Covers FE-04, FE-05
type Class struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"not null" json:"name"` // e.g. "Grade 10A Natural"
	GradeLevel int            `gorm:"default:9" json:"grade_level"`
	Section    string         `gorm:"default:'A'" json:"section"`   // A, B, C
	Stream     string         `json:"stream"`                       // Natural Science, Social Science, or empty
	Status     string         `gorm:"default:Active" json:"status"` // Active, Inactive
	Year       int            `gorm:"not null" json:"year"`
	TeacherID  *uint          `json:"teacher_id"`
	Teacher    *Teacher       `gorm:"foreignKey:TeacherID;references:ID" json:"teacher,omitempty"`
	Students   []Student      `gorm:"foreignKey:ClassID" json:"students,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-" swaggertype:"string" example:"null"` // soft delete — omitted from API responses
}

// Subject represents a course/subject taught (e.g. "Mathematics")
type Subject struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"not null" json:"name"`
	Code       string         `gorm:"uniqueIndex;not null" json:"code"` // e.g. "MATH-G9"
	GradeLevel int            `gorm:"default:0" json:"grade_level"`     // 9–12, 0 = all grades
	Stream     string         `json:"stream"`
	Status     string         `gorm:"default:Active" json:"status"` // Active, Inactive
	TeacherID  *uint          `json:"teacher_id"`
	Teacher    *Teacher       `gorm:"foreignKey:TeacherID;references:ID" json:"teacher,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-" swaggertype:"string" example:"null"`
}

// Student links a User account to academic data
// Covers FE-04
// NOTE: GradeLevel is kept on the Student record for backward compatibility and performance.
// It MUST always match the Class.GradeLevel. Use ValidateGradeClassConsistency() to enforce.
type Student struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	ParentID        uint           `json:"parent_id"`
	User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClassID         *uint          `json:"class_id"`
	Class           *Class         `gorm:"foreignKey:ClassID;references:ID" json:"class,omitempty"`
	StudentCode     string         `gorm:"uniqueIndex;not null" json:"student_code"` // e.g. "STU-2024-001"
	ParentName      string         `json:"parent_name"`
	ParentEmail     string         `json:"parent_email"`
	ParentPhone     string         `json:"parent_phone"`
	DateOfBirth     time.Time      `json:"date_of_birth"`
	EnrolledAt      time.Time      `json:"enrolled_at"`
	Stream          string         `gorm:"not null;default:''" json:"stream"`        // Natural Science | Social Science (required from G9)
	GradeLevel      int            `gorm:"default:9" json:"grade_level"`             // current grade 9–12
	PromotionStatus string         `gorm:"default:'normal'" json:"promotion_status"` // normal | conditional | repeat
	AcademicYear    int            `gorm:"default:2025" json:"academic_year"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Teacher links a User account to teaching data
type Teacher struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TeacherCode   string         `gorm:"uniqueIndex;not null" json:"teacher_code"`
	Qualification string         `json:"qualification"`
	Department    string         `json:"department"`
	JoinedAt      time.Time      `json:"joined_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-" swaggertype:"string" example:"null"` // soft delete — omitted from API responses
}

// Enrollment links a Student to a Subject (FE-05)
type Enrollment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID uint      `gorm:"not null;uniqueIndex:idx_enrollment"`
	Student   Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	SubjectID uint      `gorm:"not null;uniqueIndex:idx_enrollment"`
	Subject   Subject   `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Attendance records one status per student per calendar day (homeroom / daily).
// SubjectID is optional legacy data; new records use NULL subject_id.
type Attendance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	Student   Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	SubjectID *uint     `json:"subject_id,omitempty"`
	Subject   *Subject  `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Date      time.Time `gorm:"not null" json:"date"`
	Status    string    `gorm:"not null" json:"status"` // "Present", "Absent", "Late"
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// Grade stores a student's score for a subject assessment (FE-09, FE-10, FE-11)
type Grade struct {
	ID uint `gorm:"primaryKey" json:"id"`

	StudentID uint    `gorm:"not null;uniqueIndex:uq_grade" json:"student_id"`
	Student   Student `gorm:"foreignKey:StudentID" json:"student,omitempty"`

	SubjectID uint    `gorm:"not null;uniqueIndex:uq_grade" json:"subject_id"`
	Subject   Subject `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`

	TeacherID uint `gorm:"not null" json:"teacher_id"`

	Score    float64 `gorm:"not null" json:"score"` // 0-100
	MaxScore float64 `gorm:"not null;default:100" json:"max_score"`

	Type string `gorm:"not null;uniqueIndex:uq_grade" json:"type"` // Midterm, Final, etc
	Semester string `gorm:"not null;uniqueIndex:uq_grade" json:"semester"` // Semester 1, Semester 2, Semester 3

	AcademicYear int `gorm:"not null;uniqueIndex:uq_grade" json:"academic_year"`

	Remarks   string    `json:"remarks"`
	CreatedAt time.Time `json:"created_at"`
}