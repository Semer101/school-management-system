package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ───────────────────────────────────────────────────────────────

type SubmitReceiptRequest struct {
	StudentID   uint    `json:"student_id"  binding:"required"             example:"1"`
	Amount      float64 `json:"amount"      binding:"required,min=1"        example:"5000.00"`
	ReceiptID   string  `json:"receipt_id"  binding:"required,min=5,max=50" example:"CBE-TXN-20250501-001"`
	Description string  `json:"description"                                example:"Term 1 tuition fee"`
}

type VerifyReceiptRequest struct {
	Status  string `json:"status"  binding:"required,oneof=Verified Rejected" example:"Verified"`
	Remarks string `json:"remarks"                                             example:"Confirmed with bank"`
}

// Year field now has min=2000 to prevent year 0 / nonsense values being stored.
// Previously it was only binding:"required" which allowed 0 or 1 to pass validation.
type CreatePayrollRequest struct {
	TeacherID uint    `json:"teacher_id" binding:"required"                    example:"2"`
	Amount    float64 `json:"amount"     binding:"required,min=1"               example:"15000.00"`
	Month     int     `json:"month"      binding:"required,min=1,max=12"        example:"5"`
	Year      int     `json:"year"       binding:"required,min=2000,max=2100"   example:"2025"`
}

// ══════════════════════════════════════════════════════
//  SUBMIT BANK RECEIPT — Student or Parent only
// ══════════════════════════════════════════════════════

// SubmitBankReceipt godoc
// @Summary      Submit a bank receipt (Student or Parent only)
// @Description  Submits proof of payment via an Ethiopian bank receipt transaction ID.
// @Description  Student can only submit for themselves. Parent can submit for any of their linked children.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      SubmitReceiptRequest  true  "Receipt data"
// @Success      201   {object}  helpers.APIResponse   "Receipt submitted"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      403   {object}  helpers.APIResponse   "Forbidden — Student/Parent only, or submitting for wrong student"
// @Failure      409   {object}  helpers.APIResponse   "Receipt ID already exists"
// @Router       /api/finance/receipt [post]
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

	if userRole == models.RoleStudent {
		var student models.Student
		if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if student.ID != input.StudentID {
			helpers.Error(c, http.StatusForbidden, "you can only submit receipts for yourself")
			return
		}
	}

	if userRole == models.RoleParent {
		var count int64
		config.DB.Model(&models.Student{}).
			Where("id = ? AND parent_id = ?", input.StudentID, userID).
			Count(&count)
		if count == 0 {
			helpers.Error(c, http.StatusForbidden, "this student is not linked to your account")
			return
		}
	}

	tx := models.Transaction{
		StudentID:   input.StudentID,
		Amount:      input.Amount,
		ReceiptID:   input.ReceiptID,
		Type:        "Tuition",
		Status:      "Pending",
		Description: input.Description,
		CreatedBy:   userID,
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

// ══════════════════════════════════════════════════════
//  VERIFY RECEIPT — Admin only
// ══════════════════════════════════════════════════════

// VerifyReceipt godoc
// @Summary      Verify or reject a bank receipt (Admin only)
// @Description  Marks a pending receipt as Verified or Rejected. Can only be done once per receipt.
// @Description  FIX #6: Admin note is only appended to the description when remarks is non-empty.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  int                   true  "Transaction ID"
// @Param        body  body  VerifyReceiptRequest  true  "Verification data"
// @Success      200   {object}  helpers.APIResponse  "Receipt processed"
// @Failure      400   {object}  helpers.APIResponse  "Already processed or validation error"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      404   {object}  helpers.APIResponse  "Transaction not found"
// @Failure      500   {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/finance/receipt/{id}/verify [patch]
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

	// FIX #6: Only append the admin note when remarks is non-empty.
	// Previously, submitting with an empty remarks field still appended " | Admin note: "
	// to the description, polluting it with a trailing label and no content.
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

// ══════════════════════════════════════════════════════
//  PAYROLL — Admin only
// ══════════════════════════════════════════════════════

// CreatePayroll godoc
// @Summary      Create a payroll entry (Admin only)
// @Description  Creates a payroll record for a teacher for a specific month and year.
// @Description  FIX: Year is now validated with min=2000 — previously 0 or 1 were accepted.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreatePayrollRequest  true  "Payroll data"
// @Success      201   {object}  helpers.APIResponse  "Payroll created"
// @Failure      400   {object}  helpers.APIResponse  "Validation error (including invalid year) or teacher not found"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      409   {object}  helpers.APIResponse  "Payroll already exists for this month/year"
// @Router       /api/admin/finance/payroll [post]
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

// GetPayrolls lists payroll records (Admin only)
func GetPayrolls(c *gin.Context) {
	var payrolls []models.Payroll
	if err := config.DB.Preload("Teacher.User").
		Order("year DESC, month DESC, id DESC").
		Limit(200).
		Find(&payrolls).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch payroll")
		return
	}
	helpers.Success(c, http.StatusOK, "payroll fetched", payrolls)
}

// MarkPayrollPaid godoc
// @Summary      Mark payroll as paid (Admin only)
// @Description  Sets a payroll record's status to Paid and records the payment timestamp.
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Payroll ID"
// @Success      200  {object}  helpers.APIResponse  "Payroll updated"
// @Failure      400  {object}  helpers.APIResponse  "Already paid"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      404  {object}  helpers.APIResponse  "Payroll not found"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/finance/payroll/{id}/pay [patch]
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

// ══════════════════════════════════════════════════════
//  TRANSACTIONS
// ══════════════════════════════════════════════════════

// GetAllTransactions godoc
// @Summary      List all transactions (Admin only)
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "All transactions"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/finance/summary [get]
func GetAllTransactions(c *gin.Context) {
	offset, limit := parsePage(c)
	var transactions []models.Transaction
	var total int64

	config.DB.Model(&models.Transaction{}).Count(&total)
	config.DB.Preload("Student.User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions)

	helpers.Success(c, http.StatusOK, "transactions fetched", gin.H{
		"total": total,
		"data":  transactions,
	})
}

// GetMyTransactions godoc
// @Summary      List my transactions (Student or Parent)
// @Description  Students see their own transactions. Parents see transactions for all their linked children.
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "Transaction list"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/finance/transactions [get]
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
	// Clone the scoped query for count so WHERE clauses are preserved without Offset/Limit.
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
