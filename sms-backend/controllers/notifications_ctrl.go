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

type BroadcastInput struct {
	Title string   `json:"title" binding:"required,min=5"  example:"End of Term Reminder"`
	Body  string   `json:"body"  binding:"required,min=10" example:"Please collect report cards by Friday."`
	Roles []string `json:"roles"                            example:"[\"Student\",\"Parent\"]"`
}

type NotifyAbsencesInput struct {
	Date string `json:"date" example:"2025-05-01"`
}

// ══════════════════════════════════════════════════════
//
//	BROADCAST ANNOUNCEMENT
//
// ══════════════════════════════════════════════════════

// BroadcastAnnouncement godoc
// @Summary      Broadcast an announcement
// @Description  Sends an email announcement to all active users, or to a specific set of roles. Admin only.
// @Tags         notifications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      BroadcastInput     true  "Announcement data"
// @Success      200   {object}  helpers.APIResponse  "Announcement sent"
// @Failure      400   {object}  helpers.APIResponse  "Validation error or invalid role"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/notify/broadcast [post]
func BroadcastAnnouncement(c *gin.Context) {
	var input BroadcastInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	validRoles := map[string]bool{
		models.RoleAdmin: true, models.RoleTeacher: true,
		models.RoleStudent: true, models.RoleParent: true,
	}
	for _, r := range input.Roles {
		if !validRoles[r] {
			helpers.Error(c, http.StatusBadRequest, "invalid role: "+r+". Valid: Admin, Teacher, Student, Parent")
			return
		}
	}

	query := config.DB.Model(&models.User{}).Where("is_active = true")
	if len(input.Roles) > 0 {
		query = query.Where("role IN ?", input.Roles)
	}

	var users []models.User
	query.Find(&users)

	var emails []string
	for _, u := range users {
		emails = append(emails, u.Email)
	}

	if len(emails) > 0 {
		helpers.SendBroadcast(emails, input.Title, input.Body)
	}

	helpers.Success(c, http.StatusOK, "announcement sent", gin.H{
		"recipients": len(emails),
		"title":      input.Title,
	})
}

// ══════════════════════════════════════════════════════
//
//	NOTIFY PARENTS OF ABSENT STUDENTS
//
// ══════════════════════════════════════════════════════

// NotifyParentsAbsentStudents godoc
// @Summary      Notify parents of absent students
// @Description  Looks up all absences for the given date (or today if omitted) and sends an email to each student's parent. Admin only.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        date  query     string  false  "Date to check absences for (YYYY-MM-DD). Defaults to today."
// @Success      200   {object}  helpers.APIResponse  "Notifications sent"
// @Failure      400   {object}  helpers.APIResponse  "Invalid date format"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/notify/absences [post]
func NotifyParentsAbsentStudents(c *gin.Context) {
	dateParam := c.Query("date")

	var date string
	if dateParam == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", dateParam); err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
		date = dateParam
	}

	type AbsentRecord struct {
		StudentName string
		ParentEmail string
		SubjectName string
		Date        string
	}

	var records []AbsentRecord
	config.DB.Raw(`
		SELECT
			u.name as student_name,
			st.parent_email,
			s.name as subject_name,
			a.date::date::text as date
		FROM attendances a
		JOIN students st ON a.student_id = st.id
		JOIN users u ON st.user_id = u.id
		JOIN subjects s ON a.subject_id = s.id
		WHERE a.status = 'Absent'
		AND DATE(a.date) = DATE(?)
		AND st.parent_email IS NOT NULL
		AND st.parent_email != ''
	`, date).Scan(&records)

	notified := 0
	for _, r := range records {
		helpers.SendAbsenceAlert(r.ParentEmail, r.StudentName, r.SubjectName, r.Date)
		notified++
	}

	helpers.Success(c, http.StatusOK, "parent notifications sent", gin.H{
		"notifications_sent": notified,
		"date":               date,
	})
}
