package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

type PermanentDeleteInput struct {
	Password string `json:"password" binding:"required"`
}

// ListTrash returns soft-deleted records by entity type
func ListTrash(c *gin.Context) {
	entity := c.Query("entity") // students, teachers, classes, subjects, users
	offset, limit := parsePage(c)

	switch entity {
	case "students":
		var items []models.Student
		var total int64
		db := config.DB.Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Student{})
		db.Count(&total)
		db.Preload("User").Preload("Class").Offset(offset).Limit(limit).Find(&items)
		helpers.Success(c, http.StatusOK, "trash", gin.H{"total": total, "data": items, "entity": entity})
	case "teachers":
		var items []models.Teacher
		var total int64
		db := config.DB.Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Teacher{})
		db.Count(&total)
		db.Preload("User").Offset(offset).Limit(limit).Find(&items)
		helpers.Success(c, http.StatusOK, "trash", gin.H{"total": total, "data": items, "entity": entity})
	case "classes":
		var items []models.Class
		var total int64
		db := config.DB.Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Class{})
		db.Count(&total)
		db.Offset(offset).Limit(limit).Find(&items)
		helpers.Success(c, http.StatusOK, "trash", gin.H{"total": total, "data": items, "entity": entity})
	case "subjects":
		var items []models.Subject
		var total int64
		db := config.DB.Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Subject{})
		db.Count(&total)
		db.Offset(offset).Limit(limit).Find(&items)
		helpers.Success(c, http.StatusOK, "trash", gin.H{"total": total, "data": items, "entity": entity})
	case "users":
		var items []models.User
		var total int64
		db := config.DB.Unscoped().Where("deleted_at IS NOT NULL").Model(&models.User{})
		db.Count(&total)
		db.Offset(offset).Limit(limit).Find(&items)
		helpers.Success(c, http.StatusOK, "trash", gin.H{"total": total, "data": items, "entity": entity})
	default:
		helpers.Error(c, http.StatusBadRequest, "entity must be students, teachers, classes, subjects, or users")
	}
}

// RestoreTrash restores a soft-deleted record
func RestoreTrash(c *gin.Context) {
	entity := c.Param("entity")
	id := c.Param("id")

	switch entity {
	case "students":
		var s models.Student
		if err := config.DB.Unscoped().First(&s, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Model(&s).Update("deleted_at", nil)
		config.DB.Model(&models.User{}).Where("id = ?", s.UserID).Update("is_active", true)
		helpers.Success(c, http.StatusOK, "student restored", s)
	case "teachers":
		var t models.Teacher
		if err := config.DB.Unscoped().First(&t, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Model(&t).Update("deleted_at", nil)
		config.DB.Model(&models.User{}).Where("id = ?", t.UserID).Update("is_active", true)
		helpers.Success(c, http.StatusOK, "teacher restored", t)
	case "classes":
		var cl models.Class
		if err := config.DB.Unscoped().First(&cl, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Model(&cl).Update("deleted_at", nil)
		helpers.Success(c, http.StatusOK, "class restored", cl)
	case "subjects":
		var sub models.Subject
		if err := config.DB.Unscoped().First(&sub, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Model(&sub).Update("deleted_at", nil)
		helpers.Success(c, http.StatusOK, "subject restored", sub)
	case "users":
		var u models.User
		if err := config.DB.Unscoped().First(&u, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Model(&u).Update("deleted_at", nil)
		config.DB.Model(&u).Update("is_active", true)
		helpers.Success(c, http.StatusOK, "user restored", u)
	default:
		helpers.Error(c, http.StatusBadRequest, "invalid entity")
	}
}

// PermanentDelete permanently removes a soft-deleted record (password required)
func PermanentDelete(c *gin.Context) {
	entity := c.Param("entity")
	id, _ := strconv.Atoi(c.Param("id"))

	var input PermanentDeleteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetUint("userID")
	var admin models.User
	if err := config.DB.First(&admin, userID).Error; err != nil {
		helpers.Error(c, http.StatusUnauthorized, "user not found")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password)); err != nil {
		helpers.Error(c, http.StatusForbidden, "incorrect password")
		return
	}

	switch entity {
	case "students":
		var s models.Student
		if err := config.DB.Unscoped().First(&s, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Delete(&s)
		config.DB.Unscoped().Delete(&models.User{}, s.UserID)
	case "teachers":
		var t models.Teacher
		if err := config.DB.Unscoped().First(&t, id).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "not found")
			return
		}
		config.DB.Unscoped().Delete(&t)
		config.DB.Unscoped().Delete(&models.User{}, t.UserID)
	case "classes":
		config.DB.Unscoped().Delete(&models.Class{}, id)
	case "subjects":
		config.DB.Unscoped().Delete(&models.Subject{}, id)
	case "users":
		config.DB.Unscoped().Delete(&models.User{}, id)
	default:
		helpers.Error(c, http.StatusBadRequest, "invalid entity")
		return
	}
	helpers.Success(c, http.StatusOK, "permanently deleted", nil)
}
