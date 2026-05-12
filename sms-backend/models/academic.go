package models

import (
	"time"

	"gorm.io/gorm"
)

// Class represents a school class/grade level (e.g. "Grade 10A")
// Covers FE-04, FE-05
type Class struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"` // e.g. "Grade 10A"
	Year      int       `gorm:"not null" json:"year"` // academic year e.g. 2024
	TeacherID uint      `json:"teacher_id"`           // homeroom teacher
	Teacher   User      `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Students  []Student `gorm:"foreignKey:ClassID" json:"students,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Subject represents a course/subject taught (e.g. "Mathematics")
type Subject struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"` // e.g. "MATH101"
	TeacherID uint      `json:"teacher_id"`
	Teacher   User      `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Student links a User account to academic data
// Covers FE-04
type Student struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	ParentID    uint           `json:"parent_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClassID     uint           `json:"class_id"`
	Class       Class          `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	StudentCode string         `gorm:"uniqueIndex;not null" json:"student_code"` // e.g. "STU-2024-001"
	ParentName  string         `json:"parent_name"`
	ParentEmail string         `json:"parent_email"`
	ParentPhone string         `json:"parent_phone"`
	DateOfBirth time.Time      `json:"date_of_birth"`
	EnrolledAt  time.Time      `json:"enrolled_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Teacher links a User account to teaching data
type Teacher struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TeacherCode   string    `gorm:"uniqueIndex;not null" json:"teacher_code"`
	Qualification string    `json:"qualification"`
	JoinedAt      time.Time `json:"joined_at"`
}

// Enrollment links a Student to a Subject (FE-05)
type Enrollment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	Student   Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	SubjectID uint      `gorm:"not null" json:"subject_id"`
	Subject   Subject   `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Attendance records daily attendance per student per subject (FE-06, FE-07)
type Attendance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	Student   Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	SubjectID uint      `gorm:"not null" json:"subject_id"`
	Subject   Subject   `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Date      time.Time `gorm:"not null" json:"date"`
	Status    string    `gorm:"not null" json:"status"` // "Present", "Absent", "Late"
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// Grade stores a student's score for a subject assessment (FE-09, FE-10, FE-11)
type Grade struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StudentID    uint      `gorm:"not null" json:"student_id"`
	Student      Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	SubjectID    uint      `gorm:"not null" json:"subject_id"`
	Subject      Subject   `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	TeacherID    uint      `gorm:"not null" json:"teacher_id"`
	Score        float64   `gorm:"not null" json:"score"` // 0-100
	MaxScore     float64   `gorm:"not null;default:100" json:"max_score"`
	Type         string    `gorm:"not null" json:"type"` // "Midterm", "Final", "Quiz", "Assignment"
	Term         string    `gorm:"not null" json:"term"` // "Term1", "Term2", "Term3"
	AcademicYear int       `gorm:"not null" json:"academic_year"`
	Remarks      string    `json:"remarks"`
	CreatedAt    time.Time `json:"created_at"`
}
