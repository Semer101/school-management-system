package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ReceiptMaxFileSize is the cap on a single receipt image (5 MB).
// Larger images are common for high-resolution phone photos of paper receipts.
const ReceiptMaxFileSize = 5 * 1024 * 1024

// ReceiptAllowedExt lists the image formats parents may upload.
var ReceiptAllowedExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

// ReceiptMagicBytes is a tiny lookup for each supported format. We use it
// after http.DetectContentType to be extra sure (e.g. WEBP is sometimes
// misidentified as "application/octet-stream" on some Go versions).
var ReceiptMagicBytes = map[string][]byte{
	".jpg":  {0xFF, 0xD8, 0xFF},
	".jpeg": {0xFF, 0xD8, 0xFF},
	".png":  {0x89, 0x50, 0x4E, 0x47},
	".webp": {0x52, 0x49, 0x46, 0x46}, // "RIFF" — WEBP is RIFF-wrapped
}

// getReceiptDir returns the configured (or default) directory for receipt images.
func getReceiptDir() string {
	if dir := os.Getenv("RECEIPT_DIR"); dir != "" {
		return dir
	}
	return "./uploads/receipts"
}

// randToken returns a short random hex string used to namespace uploaded files.
func randToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b)
}

// ══════════════════════════════════════════════════════
//  Upload — Parent only
// ══════════════════════════════════════════════════════

// UploadPaymentReceipt godoc
// @Summary      Upload a bank receipt image (Parent only)
// @Description  Parents upload a JPG/JPEG/PNG/WEBP image of a bank-issued
// @Description  payment receipt. The image is stored on disk and linked
// @Description  to a student. Admin can later view and approve/reject it.
// @Tags         finance
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        receipt     formData  file    true   "Receipt image (jpg/jpeg/png/webp, max 5 MB)"
// @Param        student_id  formData  int     true   "Student ID this payment is for"
// @Param        amount      formData  number  true   "Amount in ETB"
// @Param        receipt_id  formData  string  true   "Bank transaction / receipt ID"
// @Param        description formData  string  false  "Optional note"
// @Param        academic_year formData int    false  "Academic year (defaults to student's year)"
// @Param        semester    formData  string  false  "Semester 1/2/3 or empty"
// @Success      201  {object}  helpers.APIResponse  "Receipt uploaded"
// @Failure      400  {object}  helpers.APIResponse  "Missing/invalid form data or file"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — not the linked parent"
// @Router       /api/parent/finance/receipts [post]
func UploadPaymentReceipt(c *gin.Context) {
	userID := c.GetUint("userID")
	if c.GetString("role") != models.RoleParent {
		helpers.Error(c, http.StatusForbidden, "only parents can upload payment receipts")
		return
	}

	// ── 1. Parse + validate form fields ────────────────────────────────────
	studentIDStr := c.PostForm("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 64)
	if err != nil || studentID == 0 {
		helpers.Error(c, http.StatusBadRequest, "valid student_id is required")
		return
	}

	amountStr := c.PostForm("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		helpers.Error(c, http.StatusBadRequest, "valid positive amount is required")
		return
	}

	receiptID := strings.TrimSpace(c.PostForm("receipt_id"))
	if len(receiptID) < 5 || len(receiptID) > 50 {
		helpers.Error(c, http.StatusBadRequest, "receipt_id must be 5–50 characters")
		return
	}

	description := strings.TrimSpace(c.PostForm("description"))
	if description == "" {
		helpers.Error(c, http.StatusBadRequest, "description is required")
		return
	}
	academicYearStr := c.PostForm("academic_year")
	academicYear, _ := strconv.Atoi(academicYearStr)
	semester := strings.TrimSpace(c.PostForm("semester"))
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

	// ── 2. Ownership check: parent must own this student ────────────────────
	var student models.Student
	if err := config.DB.First(&student, studentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}
	if student.ParentID != userID {
		helpers.Error(c, http.StatusForbidden, "this student is not linked to your account")
		return
	}
	if academicYear == 0 {
		academicYear = student.AcademicYear
	}

	// ── 3. Validate uploaded file (size, extension, magic bytes) ───────────
	file, err := c.FormFile("receipt")
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "receipt image is required (form field: receipt)")
		return
	}
	if file.Size > ReceiptMaxFileSize {
		helpers.Error(c, http.StatusBadRequest,
			fmt.Sprintf("receipt image exceeds %d MB limit", ReceiptMaxFileSize/1024/1024))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	expectedMIME, extAllowed := ReceiptAllowedExt[ext]
	if !extAllowed {
		helpers.Error(c, http.StatusBadRequest,
			"unsupported file type — accepted: jpg, jpeg, png, webp")
		return
	}

	src, err := file.Open()
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}
	defer src.Close()

	// Read first 16 bytes — enough to cover the magic bytes for all supported formats.
	header := make([]byte, 16)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		helpers.Error(c, http.StatusInternalServerError, "failed to read file header")
		return
	}
	detectedMIME := http.DetectContentType(header[:n])
	detectedBase := strings.Split(detectedMIME, ";")[0]
	detectedBase = strings.TrimSpace(detectedBase)

	// Two checks, defence-in-depth:
	//   (a) Go's MIME sniffer agrees with the declared extension
	//   (b) The leading magic bytes match the extension
	mimeOK := detectedBase == expectedMIME
	magicBytes, _ := ReceiptMagicBytes[ext]
	magicOK := len(magicBytes) <= n && string(header[:len(magicBytes)]) == string(magicBytes)
	if !mimeOK && !magicOK {
		helpers.Error(c, http.StatusBadRequest,
			fmt.Sprintf("file content does not look like a %s image (detected: %s)", ext, detectedBase))
		return
	}

	// ── 4. Persist the image to disk ────────────────────────────────────────
	parentID := userID
	receiptDir := getReceiptDir()
	if err := os.MkdirAll(receiptDir, 0750); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to prepare receipt directory")
		return
	}

	// Filename: parent-<id>__<rand8>__<timestamp>.<ext>
	// Keeps uploads flat and avoids leaking the original (possibly PII-bearing) name.
	safeFilename := fmt.Sprintf("parent-%d__%s__%d%s",
		parentID, randToken(4), time.Now().Unix(), ext)
	fullPath := filepath.Join(receiptDir, safeFilename)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save receipt image")
		return
	}

	// Scan file for viruses using ClamAV
	if !helpers.ScanFileSimple(fullPath) {
		os.Remove(fullPath)
		helpers.Error(c, http.StatusBadRequest, "receipt image contains a virus or malicious content and was rejected")
		return
	}

	// Relative URL served by the public /uploads static mount in main.go.
	// We strip the leading "./" or absolute path so the URL is consistent.
	relURL := "/uploads/receipts/" + safeFilename

	// ── 5. Persist the transaction row ─────────────────────────────────────
	parentIDCopy := parentID
	tx := models.Transaction{
		StudentID:       uint(studentID),
		ParentID:        &parentIDCopy,
		Amount:          amount,
		ReceiptID:       receiptID,
		ReceiptImageURL: relURL,
		Type:            "Tuition",
		Status:          "Pending",
		Description:     description,
		CreatedBy:       parentID,
		AcademicYear:    academicYear,
		Semester:        semester,
	}

	if err := config.DB.Transaction(func(db *gorm.DB) error {
		return db.Create(&tx).Error
	}); err != nil {
		// Roll back the file we just wrote.
		_ = os.Remove(fullPath)
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			helpers.Error(c, http.StatusConflict, "receipt ID already submitted")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, "failed to save receipt record")
		return
	}

	// Return the full row so the parent UI can show the new transaction.
	config.DB.Preload("Student.User").First(&tx, tx.ID)
	helpers.Success(c, http.StatusCreated, "receipt uploaded — pending admin review", tx)
}

// ══════════════════════════════════════════════════════
//  List pending receipts — Admin only
// ══════════════════════════════════════════════════════

// ListPendingReceipts godoc
// @Summary      List receipts pending admin review
// @Description  Returns all transactions with status=Pending, newest first.
// @Description  Supports filtering by status, academic year, semester, and
// @Description  a free-text search across the bank receipt ID.
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        status        query  string  false  "Pending | Verified | Rejected (default Pending)"
// @Param        academic_year query  int     false  "Filter by academic year"
// @Param        semester      query  string  false  "Filter by semester"
// @Param        search        query  string  false  "Free-text on receipt_id"
// @Success      200  {object}  helpers.APIResponse
// @Router       /api/admin/finance/receipts [get]
func ListPendingReceipts(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	status := c.Query("status")
	if status == "" {
		status = "Pending"
	}
	switch status {
	case "Pending", "Verified", "Rejected":
		// OK
	default:
		helpers.Error(c, http.StatusBadRequest, "status must be Pending, Verified, or Rejected")
		return
	}

	db := config.DB.Model(&models.Transaction{}).
		Where("transactions.type = ?", "Tuition").
		Where("transactions.status = ?", status)

	if year := c.Query("academic_year"); year != "" {
		db = db.Where("transactions.academic_year = ?", year)
	}
	if sem := c.Query("semester"); sem != "" {
		db = db.Where("transactions.semester = ?", sem)
	}
	if q := strings.TrimSpace(c.Query("search")); q != "" {
		like := "%" + q + "%"
		db = db.Where("transactions.receipt_id ILIKE ?", like)
	}

	var receipts []models.Transaction
	if err := db.
		Preload("Student.User").
		Preload("Student.Parent").
		Order("transactions.created_at DESC").
		Limit(500).
		Find(&receipts).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch receipts: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "receipts fetched", receipts)
}

// ══════════════════════════════════════════════════════
//  Approve receipt — Admin only
// ══════════════════════════════════════════════════════

// ApproveReceipt godoc
// @Summary      Approve a parent-submitted receipt
// @Description  Marks a Pending receipt as Verified. Idempotent on
// @Description  already-verified receipts (returns 200, no-op).
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Transaction ID"
// @Success      200  {object}  helpers.APIResponse
// @Router       /api/admin/finance/receipts/{id}/approve [patch]
func ApproveReceipt(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	adminID := c.GetUint("userID")
	id := c.Param("id")

	var tx models.Transaction
	if err := config.DB.First(&tx, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "receipt not found")
		return
	}
	if tx.Status == "Verified" {
		helpers.Success(c, http.StatusOK, "receipt already verified", tx)
		return
	}
	if tx.Status == "Rejected" {
		helpers.Error(c, http.StatusBadRequest, "receipt was previously rejected; create a new submission to retry")
		return
	}

	now := time.Now()
	if err := config.DB.Model(&tx).Updates(map[string]any{
		"status":          "Verified",
		"verified_by":     adminID,
		"verified_at":     &now,
		"rejection_notes": "", // clear any notes from a prior provisional reject
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to approve receipt")
		return
	}

	config.DB.Preload("Student.User").First(&tx, tx.ID)
	helpers.Success(c, http.StatusOK, "receipt approved", tx)
}

// ══════════════════════════════════════════════════════
//  Reject receipt — Admin only
// ══════════════════════════════════════════════════════

type RejectReceiptRequest struct {
	Notes string `json:"notes" binding:"required,min=3,max=500" example:"Amount does not match bank slip"`
}

// RejectReceipt godoc
// @Summary      Reject a parent-submitted receipt with notes
// @Description  Marks a Pending receipt as Rejected and stores the admin's
// @Description  rejection notes so the parent can see the reason.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path   int                    true   "Transaction ID"
// @Param        body  body   RejectReceiptRequest   true   "Rejection reason (3–500 chars)"
// @Success      200   {object}  helpers.APIResponse
// @Router       /api/admin/finance/receipts/{id}/reject [patch]
func RejectReceipt(c *gin.Context) {
	if c.GetString("role") != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "admin only")
		return
	}

	adminID := c.GetUint("userID")
	id := c.Param("id")

	var input RejectReceiptRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var tx models.Transaction
	if err := config.DB.First(&tx, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "receipt not found")
		return
	}
	if tx.Status == "Verified" {
		helpers.Error(c, http.StatusBadRequest, "receipt is already verified and cannot be rejected")
		return
	}
	if tx.Status == "Rejected" {
		helpers.Error(c, http.StatusBadRequest, "receipt is already rejected")
		return
	}

	now := time.Now()
	if err := config.DB.Model(&tx).Updates(map[string]any{
		"status":          "Rejected",
		"verified_by":     adminID,
		"verified_at":     &now,
		"rejection_notes": strings.TrimSpace(input.Notes),
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to reject receipt")
		return
	}

	config.DB.Preload("Student.User").First(&tx, tx.ID)
	helpers.Success(c, http.StatusOK, "receipt rejected", tx)
}

// ══════════════════════════════════════════════════════
//  Get one receipt — Admin only (or the uploader)
// ══════════════════════════════════════════════════════

// GetReceipt godoc
// @Summary      Get full details of a single receipt
// @Tags         finance
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Transaction ID"
// @Success      200  {object}  helpers.APIResponse
// @Router       /api/finance/receipts/{id} [get]
func GetReceipt(c *gin.Context) {
	userID := c.GetUint("userID")
	role := c.GetString("role")
	id := c.Param("id")

	var tx models.Transaction
	if err := config.DB.
		Preload("Student.User").
		Preload("Student.Parent").
		First(&tx, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "receipt not found")
		return
	}

	// Admins can read any receipt. Parents can read their own.
	if role != models.RoleAdmin {
		isOwner := false
		if tx.ParentID != nil && *tx.ParentID == userID {
			isOwner = true
		}
		if tx.CreatedBy == userID {
			isOwner = true
		}
		if !isOwner {
			helpers.Error(c, http.StatusForbidden, "you can only view your own receipts")
			return
		}
	}

	helpers.Success(c, http.StatusOK, "receipt fetched", tx)
}
