package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ───────────────────────────────────────────────────────────────

type SubmitReceiptRequest struct {
	StudentID   uint    `json:"student_id"  binding:"required"           example:"1"`
	Amount      float64 `json:"amount"      binding:"required,min=1"      example:"5000.00"`
	ReceiptID   string  `json:"receipt_id"  binding:"required,min=5,max=50" example:"CBE-TXN-20250501"`
	Description string  `json:"description"                              example:"Term 1 tuition fee"`
}

type VerifyReceiptRequest struct {
	Status  string `json:"status"  binding:"required,oneof=Verified Rejected" example:"Verified"`
	Remarks string `json:"remarks"                                             example:"Confirmed with bank"`
}

type CreatePayrollRequest struct {
	TeacherID uint    `json:"teacher_id" binding:"required"           example:"2"`
	Amount    float64 `json:"amount"     binding:"required,min=1"      example:"15000.00"`
	Month     int     `json:"month"      binding:"required,min=1,max=12" example:"5"`
	Year      int     `json:"year"       binding:"required"            example:"2025"`
}

// ══════════════════════════════════════════════════════
//
//	SUBMIT BANK RECEIPT
//
// ══════════════════════════════════════════════════════

// SubmitBankReceipt godoc
// @Summary      Submit a bank receipt
// @Description  Student or Parent submits proof of payment. Duplicate receipt IDs are rejected.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      SubmitReceiptRequest  true  "Receipt data"
// @Success      201   {object}  helpers.APIResponse     "Receipt submitted"
// @Failure      400   {object}  helpers.APIResponse     "Validation error"
// @Failure      403   {object}  helpers.APIResponse     "Forbidden — Student/Parent only"
// @Failure      409   {object}  helpers.APIResponse     "Receipt ID already exists"
// @Router       /api/finance/receipt [post]
func SubmitBankReceipt(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	if userRole != models.RoleStudent && userRole != models.RoleParent {
		helpers.Error(c, http.StatusForbidden, "not allowed")
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
			helpers.Error(c, http.StatusForbidden, "you can only submit for yourself")
			return
		}
	}

	var count int64
	config.DB.Model(&models.Transaction{}).
		Where("receipt_id = ?", input.ReceiptID).
		Count(&count)

	if count > 0 {
		helpers.Error(c, http.StatusConflict, "receipt already exists")
		return
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

	config.DB.Create(&tx)

	helpers.Success(c, http.StatusCreated, "receipt submitted", tx)
}

// ══════════════════════════════════════════════════════
//
//	VERIFY RECEIPT
//
// ══════════════════════════════════════════════════════

// VerifyReceipt godoc
// @Summary      Verify or reject a bank receipt
// @Description  Admin marks a pending receipt as Verified or Rejected. Can only be done once per receipt.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "Transaction ID"
// @Param        body  body      VerifyReceiptRequest   true  "Verification data"
// @Success      200   {object}  helpers.APIResponse      "Receipt processed"
// @Failure      400   {object}  helpers.APIResponse      "Already processed or validation error"
// @Failure      403   {object}  helpers.APIResponse      "Forbidden — Admin only"
// @Failure      404   {object}  helpers.APIResponse      "Transaction not found"
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
		helpers.Error(c, http.StatusBadRequest, "already processed")
		return
	}

	now := time.Now()

	config.DB.Model(&tx).Updates(map[string]any{
		"status":      input.Status,
		"verified_by": adminID,
		"verified_at": &now,
		"description": tx.Description + " | Admin: " + input.Remarks,
	})

	config.DB.Preload("Student.User").First(&tx, tx.ID)

	helpers.Success(c, http.StatusOK, "transaction "+input.Status, tx)
}

// ══════════════════════════════════════════════════════
//
//	PAYROLL
//
// ══════════════════════════════════════════════════════

// CreatePayroll godoc
// @Summary      Create a payroll entry
// @Description  Creates a payroll record for a teacher for a specific month and year. Admin only. Prevents duplicates.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreatePayrollRequest  true  "Payroll data"
// @Success      201   {object}  helpers.APIResponse     "Payroll created"
// @Failure      400   {object}  helpers.APIResponse     "Validation error or teacher not found"
// @Failure      403   {object}  helpers.APIResponse     "Forbidden — Admin only"
// @Failure      409   {object}  helpers.APIResponse     "Payroll already exists for this month"
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
		Where("teacher_id = ? AND month = ? AND year = ?",
			input.TeacherID, input.Month, input.Year).
		Count(&count)

	if count > 0 {
		helpers.Error(c, http.StatusConflict, "payroll already exists")
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
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	config.DB.Preload("Teacher.User").First(&payroll, payroll.ID)

	helpers.Success(c, http.StatusCreated, "payroll created", payroll)
}

// MarkPayrollPaid godoc
// @Summary      Mark payroll as paid
// @Description  Sets a payroll record's status to Paid and records the payment timestamp. Admin only.
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Payroll ID"
// @Success      200  {object}  helpers.APIResponse  "Payroll updated"
// @Failure      400  {object}  helpers.APIResponse  "Already paid"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      404  {object}  helpers.APIResponse  "Payroll not found"
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
		helpers.Error(c, http.StatusBadRequest, "already paid")
		return
	}

	now := time.Now()

	config.DB.Model(&payroll).Updates(map[string]any{
		"status":  "Paid",
		"paid_at": &now,
	})

	config.DB.Preload("Teacher.User").First(&payroll, payroll.ID)

	helpers.Success(c, http.StatusOK, "payroll updated", payroll)
}

// ══════════════════════════════════════════════════════
//
//	TRANSACTIONS
//
// ══════════════════════════════════════════════════════

// GetTransactions godoc
// @Summary      List transactions
// @Description  Admin sees all transactions. Students see only their own. Parents see transactions for all their linked children.
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "Transaction list"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Failure      404  {object}  helpers.APIResponse  "Student/user profile not found"
// @Router       /api/finance/transactions [get]
func GetTransactions(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	query := config.DB.Preload("Student.User")

	switch userRole {
	case models.RoleStudent:
		var student models.Student
		if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		query = query.Where("student_id = ?", student.ID)
	case models.RoleParent:
		var parent models.User
		if err := config.DB.First(&parent, userID).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "user not found")
			return
		}
		var studentIDs []uint
		config.DB.Model(&models.Student{}).
			Where("parent_email = ?", parent.Email).
			Pluck("id", &studentIDs)
		query = query.Where("student_id IN ?", studentIDs)
	}

	var transactions []models.Transaction
	query.Order("created_at DESC").Find(&transactions)

	helpers.Success(c, http.StatusOK, "transactions fetched", transactions)
}
