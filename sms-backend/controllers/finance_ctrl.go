package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

type SubmitReceiptRequest struct {
	StudentID    uint    `json:"student_id"    binding:"required"             example:"1"`
	Amount       float64 `json:"amount"        binding:"required,min=1"        example:"5000.00"`
	ReceiptID    string  `json:"receipt_id"    binding:"required,min=5,max=50" example:"CBE-TXN-20250501-001"`
	Description  string  `json:"description"                                   example:"Semester 1 tuition fee"`
	AcademicYear int     `json:"academic_year"`
	Semester     string  `json:"semester"`
}

type VerifyReceiptRequest struct {
	Status  string `json:"status"  binding:"required,oneof=Verified Rejected" example:"Verified"`
	Remarks string `json:"remarks"                                             example:"Confirmed with bank"`
}

type CreatePayrollRequest struct {
	TeacherID uint    `json:"teacher_id" binding:"required"                    example:"2"`
	Amount    float64 `json:"amount"     binding:"required,min=1"               example:"15000.00"`
	Month     int     `json:"month"      binding:"required,min=1,max=12"        example:"5"`
	Year      int     `json:"year"       binding:"required,min=2000,max=2100"   example:"2025"`
}

func SubmitBankReceipt(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	if userRole != models.RoleStudent && userRole != models.RoleParent {
		helpers.Error(c, http.StatusForbidden, "only students and parents can submit receipts")
		return
	}

	var input SubmitReceiptRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var student models.Student
	if err := config.DB.First(&student, input.StudentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}

	if userRole == models.RoleStudent {
		if student.UserID != userID {
			helpers.Error(c, http.StatusForbidden, "you can only submit receipts for yourself")
			return
		}
	}

	if userRole == models.RoleParent {
		if student.ParentID != userID {
			helpers.Error(c, http.StatusForbidden, "this student is not linked to your account")
			return
		}
	}

	academicYear := input.AcademicYear
	if academicYear == 0 {
		academicYear = student.AcademicYear
	}

	// FIX #5: accept empty semester for non-semester fees; validate canonical labels.
	semester := strings.TrimSpace(input.Semester)
	switch semester {
	case "Semester 1", "Semester 2", "Semester 3":
		// OK
	case "":
		semester = ""
	default:
		helpers.Error(c, http.StatusBadRequest,
			"semester must be one of 'Semester 1', 'Semester 2', 'Semester 3', or omitted for non-semester fees")
		return
	}

	tx := models.Transaction{
		StudentID:    input.StudentID,
		Amount:       input.Amount,
		ReceiptID:    input.ReceiptID,
		Type:         "Tuition",
		Status:       "Pending",
		Description:  input.Description,
		CreatedBy:    userID,
		AcademicYear: academicYear,
		Semester:     semester,
	}

	err := config.DB.Transaction(func(db *gorm.DB) error {
		return db.Create(&tx).Error
	})

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			helpers.Error(c, http.StatusConflict, "receipt ID already submitted")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, "failed to submit receipt")
		return
	}
	helpers.Success(c, http.StatusCreated, "receipt submitted", tx)
}

func VerifyReceipt(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	adminID := c.GetUint("userID")
	txID := c.Param("id")

	var input VerifyReceiptRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var tx models.Transaction
	if err := config.DB.First(&tx, txID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "transaction not found")
		return
	}
	if tx.Status != "Pending" {
		helpers.Error(c, http.StatusBadRequest, "transaction already processed")
		return
	}

	now := time.Now()
	description := tx.Description
	if input.Remarks != "" {
		description = tx.Description + " | Admin note: " + input.Remarks
	}

	if err := config.DB.Model(&tx).Updates(map[string]any{
		"status":      input.Status,
		"verified_by": adminID,
		"verified_at": &now,
		"description": description,
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update transaction")
		return
	}

	config.DB.Preload("Student.User").First(&tx, tx.ID)
	helpers.Success(c, http.StatusOK, "transaction "+input.Status, tx)
}

func CreatePayroll(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	var input CreatePayrollRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var teacher models.Teacher
	if err := config.DB.First(&teacher, input.TeacherID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "teacher not found")
		return
	}

	var count int64
	config.DB.Model(&models.Payroll{}).
		Where("teacher_id = ? AND month = ? AND year = ?", input.TeacherID, input.Month, input.Year).
		Count(&count)
	if count > 0 {
		helpers.Error(c, http.StatusConflict, "payroll already exists for this teacher/month/year")
		return
	}

	payroll := models.Payroll{
		TeacherID: input.TeacherID,
		Amount:    input.Amount,
		Month:     input.Month,
		Year:      input.Year,
		Status:    "Pending",
	}

	if err := config.DB.Create(&payroll).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create payroll: "+err.Error())
		return
	}

	config.DB.Preload("Teacher.User").First(&payroll, payroll.ID)
	helpers.Success(c, http.StatusCreated, "payroll created", payroll)
}

func GetPayrolls(c *gin.Context) {
	var payrolls []models.Payroll
	db := config.DB.Model(&models.Payroll{})

	if month := c.Query("month"); month != "" {
		db = db.Where("payrolls.month = ?", month)
	}
	if year := c.Query("year"); year != "" {
		db = db.Where("payrolls.year = ?", year)
	}
	if dept := c.Query("department"); dept != "" {
		db = db.Joins("JOIN teachers ON payrolls.teacher_id = teachers.id").
			Where("teachers.department = ?", dept)
	}

	if err := db.Preload("Teacher.User").
		Order("payrolls.year DESC, payrolls.month DESC, payrolls.id DESC").
		Limit(200).
		Find(&payrolls).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch payroll")
		return
	}
	helpers.Success(c, http.StatusOK, "payroll fetched", payrolls)
}

func MarkPayrollPaid(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	id := c.Param("id")
	var payroll models.Payroll
	if err := config.DB.First(&payroll, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "payroll not found")
		return
	}
	if payroll.Status == "Paid" {
		helpers.Error(c, http.StatusBadRequest, "payroll already paid")
		return
	}

	now := time.Now()
	if err := config.DB.Model(&payroll).Updates(map[string]any{
		"status":  "Paid",
		"paid_at": &now,
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update payroll")
		return
	}

	config.DB.Preload("Teacher.User").First(&payroll, payroll.ID)
	helpers.Success(c, http.StatusOK, "payroll marked as paid", payroll)
}

func GetAllTransactions(c *gin.Context) {
	offset, limit := parsePage(c)
	var transactions []models.Transaction
	var total int64

	db := config.DB.Model(&models.Transaction{})

	if year := c.Query("academic_year"); year != "" {
		db = db.Where("transactions.academic_year = ?", year)
	}
	if sem := c.Query("semester"); sem != "" {
		db = db.Where("transactions.semester = ?", sem)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("transactions.status = ?", status)
	}
	if student := c.Query("student"); student != "" {
		like := "%" + student + "%"
		db = db.Joins("JOIN students ON transactions.student_id = students.id").
			Joins("JOIN users ON students.user_id = users.id").
			Where("users.name ILIKE ? OR students.student_code ILIKE ?", like, like)
	}

	db.Count(&total)

	if err := db.Preload("Student.User").
		Order("transactions.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	helpers.Success(c, http.StatusOK, "transactions fetched", gin.H{
		"total": total,
		"data":  transactions,
	})
}

func GetMyTransactions(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")
	offset, limit := parsePage(c)

	query := config.DB.Preload("Student.User").Order("created_at DESC")

	switch userRole {
	case models.RoleStudent:
		var student models.Student
		if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		query = query.Where("student_id = ?", student.ID)

	case models.RoleParent:
		var studentIDs []uint
		config.DB.Model(&models.Student{}).
			Where("parent_id = ?", userID).
			Pluck("id", &studentIDs)
		if len(studentIDs) == 0 {
			helpers.Success(c, http.StatusOK, "no linked students found", gin.H{"total": 0, "data": []any{}})
			return
		}
		query = query.Where("student_id IN ?", studentIDs)
	}

	var transactions []models.Transaction
	var total int64
	countQ := query.Session(&gorm.Session{})
	countQ.Model(&models.Transaction{}).Count(&total)
	if err := query.Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	helpers.Success(c, http.StatusOK, "transactions fetched", gin.H{
		"total": total,
		"data":  transactions,
	})
}

// GetOverduePayments — FIX #6: single-query replacement for the per-student N+1 loop.
func GetOverduePayments(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	type StudentPaymentStatus struct {
		StudentID    uint   `json:"student_id"`
		StudentName  string `json:"student_name"`
		StudentCode  string `json:"student_code"`
		ClassName    string `json:"class_name"`
		AcademicYear int    `json:"academic_year"`
		Semester1    string `json:"semester_1"`
		Semester2    string `json:"semester_2"`
		Semester3    string `json:"semester_3"`
	}

	query := `
		SELECT
			s.id AS student_id,
			COALESCE(u.name, '') AS student_name,
			s.student_code AS student_code,
			COALESCE(cl.name, 'Grade ' || s.grade_level::text) AS class_name,
			s.academic_year AS academic_year,
			CASE
				WHEN MAX(CASE WHEN t.semester = 'Semester 1' AND t.status = 'Verified' THEN 1 ELSE 0 END) = 1 THEN 'Paid'
				WHEN MAX(CASE WHEN t.semester = 'Semester 1' AND t.status = 'Pending'  THEN 1 ELSE 0 END) = 1 THEN 'Pending'
				ELSE 'Overdue'
			END AS semester_1,
			CASE
				WHEN MAX(CASE WHEN t.semester = 'Semester 2' AND t.status = 'Verified' THEN 1 ELSE 0 END) = 1 THEN 'Paid'
				WHEN MAX(CASE WHEN t.semester = 'Semester 2' AND t.status = 'Pending'  THEN 1 ELSE 0 END) = 1 THEN 'Pending'
				ELSE 'Overdue'
			END AS semester_2,
			CASE
				WHEN MAX(CASE WHEN t.semester = 'Semester 3' AND t.status = 'Verified' THEN 1 ELSE 0 END) = 1 THEN 'Paid'
				WHEN MAX(CASE WHEN t.semester = 'Semester 3' AND t.status = 'Pending'  THEN 1 ELSE 0 END) = 1 THEN 'Pending'
				ELSE 'Overdue'
			END AS semester_3
		FROM students s
		JOIN users u ON s.user_id = u.id AND u.is_active = true AND u.deleted_at IS NULL
		LEFT JOIN classes cl ON s.class_id = cl.id
		LEFT JOIN transactions t
		       ON t.student_id = s.id
		      AND t.academic_year = s.academic_year
		      AND t.type = 'Tuition'
		GROUP BY s.id, u.name, s.student_code, cl.name, s.grade_level, s.academic_year
		ORDER BY u.name
	`

	var result []StudentPaymentStatus
	if err := config.DB.Raw(query).Scan(&result).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to compute overdue payments: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "overdue payments fetched", result)
}

type PaymentReminderRequest struct {
	StudentID    uint   `json:"student_id" binding:"required"`
	AcademicYear int    `json:"academic_year" binding:"required"`
	Semester     string `json:"semester" binding:"required"`
}

func SendPaymentReminder(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	adminID := c.GetUint("userID")

	var input PaymentReminderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var student models.Student
	if err := config.DB.Preload("User").First(&student, input.StudentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	title := "Payment Reminder: Tuition Fee"
	body := fmt.Sprintf("Dear %s, this is a friendly reminder to settle the outstanding tuition fee for %s, Academic Year %d.",
		student.User.Name, input.Semester, input.AcademicYear)

	notification := models.Notification{
		Title:       title,
		Body:        body,
		TargetRoles: "Student,Parent",
		SenderID:    adminID,
	}
	if err := config.DB.Create(&notification).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create reminder notification")
		return
	}

	config.DB.Create(&models.NotificationReceipt{
		UserID:         student.UserID,
		NotificationID: notification.ID,
		IsRead:         false,
	})

	if student.ParentID > 0 {
		config.DB.Create(&models.NotificationReceipt{
			UserID:         student.ParentID,
			NotificationID: notification.ID,
			IsRead:         false,
		})
	}

	emails := []string{student.User.Email}
	if student.ParentEmail != "" {
		emails = append(emails, student.ParentEmail)
	}
	go helpers.SendBroadcast(emails, title, body)

	helpers.Success(c, http.StatusOK, "payment reminder sent successfully", nil)
}
