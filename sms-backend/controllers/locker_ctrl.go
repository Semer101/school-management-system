package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5 MB
)

// FIX #15: UploadDir is resolved at runtime from the UPLOAD_DIR env variable so the
// binary can be started from any working directory (e.g. inside Docker) without
// misrouting uploaded files. Falls back to "./uploads/locker" for local dev.
func getUploadDir() string {
	if dir := os.Getenv("UPLOAD_DIR"); dir != "" {
		return dir
	}
	return "./uploads/locker"
}

// AllowedExtensions maps permitted file extensions to their expected MIME type prefixes.
// MIME type is now verified against actual file content (first 512 bytes),
// not just the filename extension. A file named shell.pdf with PHP content is rejected.
var AllowedExtensions = map[string]string{
	".pdf":  "application/pdf",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".txt":  "text/plain",
}

// ══════════════════════════════════════════════════════
//  UPLOAD — Student only
// ══════════════════════════════════════════════════════

// UploadLockerFile godoc
// @Summary      Upload a file to the locker (Student only)
// @Description  Students upload files (max 5 MB) to their private digital locker.
// @Description  Allowed types: pdf, jpg, jpeg, png, doc, docx, txt.
// @Description  FIX #14: Both the file extension AND the actual file content (MIME sniffing) are validated.
// @Description  A file with a renamed extension (e.g. shell.pdf containing PHP) is rejected.
// @Description  FIX #15: Upload directory is configurable via UPLOAD_DIR env variable (defaults to ./uploads/locker).
// @Tags         locker
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file      formData  file    true   "File to upload (max 5 MB)"
// @Param        category  formData  string  false  "Category: Certificate | Assignment | Portfolio | Material"
// @Success      201  {object}  helpers.APIResponse  "File uploaded"
// @Failure      400  {object}  helpers.APIResponse  "File missing, too large, invalid extension, or MIME type mismatch"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Student only"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/locker/upload [post]
func UploadLockerFile(c *gin.Context) {
	userID := c.GetUint("userID")
	if c.GetString("role") != models.RoleStudent {
		helpers.Error(c, http.StatusForbidden, "only students can upload to their locker")
		return
	}

	var student models.Student
	if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "file is required")
		return
	}
	if file.Size > MaxFileSize {
		helpers.Error(c, http.StatusBadRequest, "file size exceeds 5 MB limit")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	expectedMIMEPrefix, extAllowed := AllowedExtensions[ext]
	if !extAllowed {
		helpers.Error(c, http.StatusBadRequest, "file extension not allowed — accepted: pdf, jpg, png, doc, docx, txt")
		return
	}

	// read the first 512 bytes to detect the actual MIME type.
	// http.DetectContentType inspects the raw bytes, not the filename.
	src, err := file.Open()
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to read file")
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		helpers.Error(c, http.StatusInternalServerError, "failed to read file header")
		return
	}
	detectedMIME := http.DetectContentType(header[:n])

	// Normalise: http.DetectContentType may append charset or parameters.
	// We only care about the primary type/subtype prefix.
	detectedBase := strings.Split(detectedMIME, ";")[0]
	detectedBase = strings.TrimSpace(detectedBase)

	// For .doc/.docx, Go's MIME sniffer sees the ZIP/OLE magic bytes and returns
	// "application/zip" or "application/octet-stream". We allow these for office formats.
	mimeOK := false
	switch {
	case detectedBase == expectedMIMEPrefix:
		mimeOK = true
	case (ext == ".doc" || ext == ".docx") &&
		(detectedBase == "application/zip" ||
			detectedBase == "application/octet-stream" ||
			detectedBase == "application/x-cfb"):
		// Office documents are ZIP or OLE compound files — this is expected.
		mimeOK = true
	case ext == ".txt" && strings.HasPrefix(detectedBase, "text/"):
		// Plain text can be detected as text/html if it starts with "<" — still allow it.
		mimeOK = true
	}

	if !mimeOK {
		helpers.Error(c, http.StatusBadRequest,
			fmt.Sprintf("file content does not match extension .%s (detected: %s)", strings.TrimPrefix(ext, "."), detectedBase))
		return
	}

	category := c.PostForm("category")
	if category == "" {
		category = "General"
	}

	timestamp := time.Now().UnixNano()
	safeFilename := fmt.Sprintf("%d%s", timestamp, ext)
	displayName := filepath.Base(file.Filename)

	studentDir := fmt.Sprintf("%s/%d", getUploadDir(), student.ID)
	if err := os.MkdirAll(studentDir, 0750); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	filePath := fmt.Sprintf("%s/%s", studentDir, safeFilename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	lockerFile := models.LockerFile{
		StudentID:  student.ID,
		FileName:   displayName,
		FilePath:   filePath,
		FileSize:   file.Size,
		FileType:   strings.TrimPrefix(ext, "."),
		Category:   category,
		IsPublic:   false,
		UploadedAt: time.Now(),
	}

	if err := config.DB.Create(&lockerFile).Error; err != nil {
		os.Remove(filePath)
		helpers.Error(c, http.StatusInternalServerError, "failed to save file record")
		return
	}

	helpers.Success(c, http.StatusCreated, "file uploaded successfully", gin.H{
		"id":          lockerFile.ID,
		"file_name":   lockerFile.FileName,
		"file_size":   lockerFile.FileSize,
		"file_type":   lockerFile.FileType,
		"category":    lockerFile.Category,
		"is_public":   lockerFile.IsPublic,
		"uploaded_at": lockerFile.UploadedAt,
	})
}

// ══════════════════════════════════════════════════════
//  GET MY FILES — Student only
// ══════════════════════════════════════════════════════

// GetMyLockerFiles godoc
// @Summary      List my locker files (Student only)
// @Description  Returns all files in the authenticated student's own locker (both public and private).
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "File list"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/locker/my-files [get]
func GetMyLockerFiles(c *gin.Context) {
	userID := c.GetUint("userID")

	var student models.Student
	if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}

	var files []models.LockerFile
	config.DB.Where("student_id = ?", student.ID).Find(&files)
	helpers.Success(c, http.StatusOK, "files fetched", files)
}

// ══════════════════════════════════════════════════════
//  TEACHER READ-ONLY — public files of a specific student
// ══════════════════════════════════════════════════════

// TeacherGetPublicFiles godoc
// @Summary      List a student's public locker files (Teacher only)
// @Description  Returns only files the student has toggled to is_public=true.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path  int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "Public file list"
// @Router       /api/locker/student/{studentID}/public [get]
func TeacherGetPublicFiles(c *gin.Context) {
	studentID := c.Param("studentID")
	var files []models.LockerFile
	config.DB.
		Where("student_id = ? AND is_public = true", studentID).
		Find(&files)
	helpers.Success(c, http.StatusOK, "public files fetched", files)
}

// ══════════════════════════════════════════════════════
//  ADMIN READ-ONLY — all files of a specific student
// ══════════════════════════════════════════════════════

// AdminGetLockerFiles godoc
// @Summary      List all locker files for a student (Admin only)
// @Description  Returns all files (public and private) for compliance or support review.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path  int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "All files for student"
// @Router       /api/admin/locker/student/{studentID} [get]
func AdminGetLockerFiles(c *gin.Context) {
	studentID := c.Param("studentID")
	var files []models.LockerFile
	config.DB.Where("student_id = ?", studentID).Find(&files)
	helpers.Success(c, http.StatusOK, "files fetched", files)
}

// ══════════════════════════════════════════════════════
//  DELETE — Student only
// ══════════════════════════════════════════════════════

// DeleteLockerFile godoc
// @Summary      Delete a locker file (Student only)
// @Description  Removes the file from disk and the database. Student must own the file.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        fileID  path  int  true  "File ID"
// @Success      200  {object}  helpers.APIResponse  "File deleted"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — not your file"
// @Failure      404  {object}  helpers.APIResponse  "File not found"
// @Router       /api/locker/files/{fileID} [delete]
func DeleteLockerFile(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("fileID")

	var file models.LockerFile
	if err := config.DB.First(&file, fileID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "file not found")
		return
	}

	var student models.Student
	if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}
	if file.StudentID != student.ID {
		helpers.Error(c, http.StatusForbidden, "you can only delete your own files")
		return
	}

	os.Remove(file.FilePath)
	if err := config.DB.Delete(&file).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to delete file record")
		return
	}
	helpers.Success(c, http.StatusOK, "file deleted", nil)
}

// ══════════════════════════════════════════════════════
//  TOGGLE VISIBILITY — Student only
// ══════════════════════════════════════════════════════

// ToggleFileVisibility godoc
// @Summary      Toggle file visibility (Student only)
// @Description  Flips is_public on a locker file. When is_public=true, the student's teachers can read it.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        fileID  path  int  true  "File ID"
// @Success      200  {object}  helpers.APIResponse  "Visibility toggled"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — not your file"
// @Failure      404  {object}  helpers.APIResponse  "File not found"
// @Router       /api/locker/files/{fileID}/visibility [patch]
func ToggleFileVisibility(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("fileID")

	var file models.LockerFile
	if err := config.DB.First(&file, fileID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "file not found")
		return
	}

	var student models.Student
	if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student profile not found")
		return
	}

	if file.StudentID != student.ID {
		helpers.Error(c, http.StatusForbidden, "you can only modify your own files")
		return
	}

	newVisibility := !file.IsPublic
	if err := config.DB.Model(&file).Update("is_public", newVisibility).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update file visibility")
		return
	}

	status := "private"
	if newVisibility {
		status = "visible to teachers"
	}
	helpers.Success(c, http.StatusOK, fmt.Sprintf("file is now %s", status), gin.H{
		"file_id":   file.ID,
		"is_public": newVisibility,
	})
}
