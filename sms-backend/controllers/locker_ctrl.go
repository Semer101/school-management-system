package controllers

import (
	"fmt"
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
	MaxFileSize = 5 * 1024 * 1024 // 5MB
	UploadDir   = "./uploads/locker"
)

var AllowedFileTypes = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
}

// UploadLockerFile godoc
// @Summary      Upload a file to the digital locker
// @Description  Students can upload files (max 5MB) of allowed types: pdf, jpg, jpeg, png, doc, docx, txt. Each file is stored under the student's locker directory.
// @Tags         locker
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file      formData  file    true  "File to upload (max 5MB)"
// @Param        category  formData  string  true  "Category: Certificate | Assignment | Portfolio | Material"
// @Success      201  {object}  helpers.APIResponse  "File uploaded"
// @Failure      400  {object}  helpers.APIResponse  "File missing, too large, or invalid type"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Students only"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/locker/upload [post]
func UploadLockerFile(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	if userRole != models.RoleStudent {
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
		helpers.Error(c, http.StatusBadRequest, "file size exceeds 5MB limit")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedFileTypes[ext] {
		helpers.Error(c, http.StatusBadRequest, "file type not allowed. Allowed: pdf, jpg, png, doc, docx, txt")
		return
	}

	safeFilename := filepath.Base(file.Filename)
	safeFilename = strings.ReplaceAll(safeFilename, "..", "")

	studentDir := fmt.Sprintf("%s/%d", UploadDir, student.ID)
	if err := os.MkdirAll(studentDir, 0750); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, safeFilename)
	filePath := fmt.Sprintf("%s/%s", studentDir, uniqueFilename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	category := c.PostForm("category")
	if category == "" {
		category = "General"
	}

	lockerFile := models.LockerFile{
		StudentID:  student.ID,
		FileName:   safeFilename,
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
		"uploaded_at": lockerFile.UploadedAt,
	})
}

// GetLockerFiles godoc
// @Summary      List files in a student's locker
// @Description  Students see all their own files. Teachers see only public files. Admins see all files.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path      int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "File list"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Cannot view another student's files"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/locker/student/{studentID} [get]
func GetLockerFiles(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")
	requestedStudentID := c.Param("studentID")

	if userRole == models.RoleStudent {
		var student models.Student
		if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if fmt.Sprintf("%d", student.ID) != requestedStudentID {
			helpers.Error(c, http.StatusForbidden, "you can only view your own files")
			return
		}
	}

	query := config.DB.Where("student_id = ?", requestedStudentID)
	if userRole == models.RoleTeacher {
		query = query.Where("is_public = true")
	}

	var files []models.LockerFile
	query.Find(&files)

	helpers.Success(c, http.StatusOK, "files fetched", files)
}

// DeleteLockerFile godoc
// @Summary      Delete a locker file
// @Description  Students can only delete their own files. Admins can delete any file. Removes from disk and database.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        fileID  path      int  true  "File ID"
// @Success      200  {object}  helpers.APIResponse  "File deleted"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Not your file"
// @Failure      404  {object}  helpers.APIResponse  "File not found"
// @Router       /api/locker/files/{fileID} [delete]
func DeleteLockerFile(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("role")
	fileID := c.Param("fileID")

	var file models.LockerFile
	if err := config.DB.First(&file, fileID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "file not found")
		return
	}

	if userRole == models.RoleStudent {
		var student models.Student
		config.DB.Where("user_id = ?", userID).First(&student)
		if file.StudentID != student.ID {
			helpers.Error(c, http.StatusForbidden, "you can only delete your own files")
			return
		}
	}

	os.Remove(file.FilePath)
	config.DB.Delete(&file)

	helpers.Success(c, http.StatusOK, "file deleted", nil)
}

// ToggleFileVisibility godoc
// @Summary      Toggle file visibility
// @Description  Toggles a locker file between private (teacher cannot see) and public (teacher can see). Student must own the file.
// @Tags         locker
// @Security     BearerAuth
// @Produce      json
// @Param        fileID  path      int  true  "File ID"
// @Success      200  {object}  helpers.APIResponse  "Visibility updated"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Not your file"
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
	config.DB.Where("user_id = ?", userID).First(&student)
	if file.StudentID != student.ID {
		helpers.Error(c, http.StatusForbidden, "you can only modify your own files")
		return
	}

	newVisibility := !file.IsPublic
	config.DB.Model(&file).Update("is_public", newVisibility)

	status := "private"
	if newVisibility {
		status = "visible to teachers"
	}

	helpers.Success(c, http.StatusOK, fmt.Sprintf("file is now %s", status), gin.H{
		"is_public": newVisibility,
	})
}
