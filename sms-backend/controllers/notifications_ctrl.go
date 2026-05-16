package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ══════════════════════════════════════════════════════
//  SSE HUB
// ══════════════════════════════════════════════════════

type sseClient struct {
	userID uint
	role   string
	ch     chan string
}

type sseHub struct {
	mu      sync.RWMutex
	clients map[uint]*sseClient
}

var hub = &sseHub{
	clients: make(map[uint]*sseClient),
}

func (h *sseHub) register(c *sseClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[c.userID]; ok {
		close(old.ch)
	}
	h.clients[c.userID] = c
}

// FIX #2: unregister now accepts the client pointer and only removes/closes the
// entry if it still points to THIS client instance.
// The previous version accepted only a userID, so when a user reconnected:
//  1. register() closed the old client's channel (correct).
//  2. The old goroutine exited and its deferred unregister(userID) ran.
//  3. unregister found the *new* client in the map, closed its channel, and
//     deleted it — silently killing the brand-new SSE stream.
func (h *sseHub) unregister(c *sseClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current, ok := h.clients[c.userID]; ok && current == c {
		close(current.ch)
		delete(h.clients, c.userID)
	}
}

func (h *sseHub) push(payload string, targetRoles []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roleSet := make(map[string]bool, len(targetRoles))
	for _, r := range targetRoles {
		roleSet[r] = true
	}
	for _, c := range h.clients {
		if len(roleSet) == 0 || roleSet[c.role] {
			select {
			case c.ch <- payload:
			default:
			}
		}
	}
}

// ══════════════════════════════════════════════════════
//  INPUT TYPES
// ══════════════════════════════════════════════════════

type BroadcastInput struct {
	Title string   `json:"title" binding:"required,min=5"  example:"End of Term Reminder"`
	Body  string   `json:"body"  binding:"required,min=10" example:"Please collect report cards by Friday."`
	Roles []string `json:"roles" example:"[\"Student\",\"Parent\"]"`
}

type NotifyAbsencesInput struct {
	Date string `json:"date" example:"2025-05-01"`
}

// ══════════════════════════════════════════════════════
//  SSE STREAM
// ══════════════════════════════════════════════════════

// StreamNotifications godoc
// @Summary      SSE stream for live notifications
// @Description  Opens a Server-Sent Events stream. Receives a "notification" event for every broadcast targeting the caller's role.
// @Description  FIX: This endpoint now uses SSEAuthMiddleware. Call POST /api/notifications/sse-token first to get a
// @Description  short-lived 60-second token, then connect with ?sse_token=<token>.
// @Description  The full access JWT must NOT be passed in the URL — it would be permanently logged by every proxy.
// @Tags         notifications
// @Produce      text/event-stream
// @Param        sse_token  query  string  true  "Short-lived SSE token (from POST /api/notifications/sse-token)"
// @Success      200  {string}  string              "SSE stream open"
// @Failure      401  {object}  helpers.APIResponse "Invalid or expired SSE token"
// @Router       /api/notifications/stream [get]
func StreamNotifications(c *gin.Context) {
	// userID and role are injected by SSEAuthMiddleware (set in routes.go).
	userID := c.GetUint("userID")
	role := c.GetString("role")

	client := &sseClient{
		userID: userID,
		role:   role,
		ch:     make(chan string, 16),
	}
	hub.register(client)
	defer hub.unregister(client) // FIX #2: pass pointer so only THIS client is removed

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	for {
		select {
		case payload, ok := <-client.ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", payload)
			c.Writer.Flush()

		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

// ══════════════════════════════════════════════════════
//  BROADCAST ANNOUNCEMENT — Admin only
// ══════════════════════════════════════════════════════

// BroadcastAnnouncement godoc
// @Summary      Broadcast an announcement (Admin only)
// @Description  Saves the notification, creates receipts for all targeted users, pushes live via SSE, and sends emails.
// @Tags         notifications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      BroadcastInput       true  "Announcement data"
// @Success      200   {object}  helpers.APIResponse  "Announcement sent"
// @Failure      400   {object}  helpers.APIResponse  "Validation error or invalid role"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/notify/broadcast [post]
func BroadcastAnnouncement(c *gin.Context) {
	senderID := c.GetUint("userID")

	var input BroadcastInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	validRoles := map[string]bool{
		models.RoleAdmin:   true,
		models.RoleTeacher: true,
		models.RoleStudent: true,
		models.RoleParent:  true,
	}
	for _, r := range input.Roles {
		if !validRoles[r] {
			helpers.Error(c, http.StatusBadRequest,
				"invalid role: "+r+". Valid roles: Admin, Teacher, Student, Parent")
			return
		}
	}

	notification := models.Notification{
		Title:       input.Title,
		Body:        input.Body,
		TargetRoles: strings.Join(input.Roles, ","),
		SenderID:    senderID,
	}
	if err := config.DB.Create(&notification).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save notification")
		return
	}

	query := config.DB.Model(&models.User{}).Where("is_active = true")
	if len(input.Roles) > 0 {
		query = query.Where("role IN ?", input.Roles)
	}
	var users []models.User
	query.Find(&users)

	receipts := make([]models.NotificationReceipt, 0, len(users))
	emails := make([]string, 0, len(users))
	for _, u := range users {
		receipts = append(receipts, models.NotificationReceipt{
			UserID:         u.ID,
			NotificationID: notification.ID,
			IsRead:         false,
		})
		emails = append(emails, u.Email)
	}

	if len(receipts) > 0 {
		if err := config.DB.CreateInBatches(&receipts, 200).Error; err != nil {
			fmt.Printf("[WARN] Failed to create notification receipts: %v\n", err)
		}
	}

	type ssePayload struct {
		ID        uint      `json:"id"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		SenderID  uint      `json:"sender_id"`
		CreatedAt time.Time `json:"created_at"`
	}
	payloadBytes, _ := json.Marshal(ssePayload{
		ID:        notification.ID,
		Title:     notification.Title,
		Body:      notification.Body,
		SenderID:  notification.SenderID,
		CreatedAt: notification.CreatedAt,
	})
	hub.push(string(payloadBytes), input.Roles)

	if len(emails) > 0 {
		go helpers.SendBroadcast(emails, input.Title, input.Body)
	}

	helpers.Success(c, http.StatusOK, "announcement sent", gin.H{
		"recipients":      len(users),
		"notification_id": notification.ID,
		"title":           notification.Title,
	})
}

// ══════════════════════════════════════════════════════
//  GET MY NOTIFICATIONS
// ══════════════════════════════════════════════════════

type NotificationResponse struct {
	ReceiptID  uint       `json:"receipt_id"`
	ID         uint       `json:"id"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	SenderName string     `json:"sender_name"`
	IsRead     bool       `json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	ReceivedAt time.Time  `json:"received_at"`
}

// GetMyNotifications godoc
// @Summary      Get my notifications
// @Description  Returns all notifications targeted at the current user, most recent first.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "Notification list"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Router       /api/notifications [get]
func GetMyNotifications(c *gin.Context) {
	userID := c.GetUint("userID")

	type row struct {
		ReceiptID  uint
		NotifID    uint
		Title      string
		Body       string
		SenderName string
		IsRead     bool
		ReadAt     *time.Time
		ReceivedAt time.Time
	}

	var rows []row
	config.DB.Raw(`
		SELECT
			nr.id          AS receipt_id,
			n.id           AS notif_id,
			n.title,
			n.body,
			u.name         AS sender_name,
			nr.is_read,
			nr.read_at,
			n.created_at   AS received_at
		FROM notification_receipts nr
		JOIN notifications n ON nr.notification_id = n.id
		JOIN users         u ON n.sender_id        = u.id
		WHERE nr.user_id = ?
		ORDER BY n.created_at DESC
	`, userID).Scan(&rows)

	result := make([]NotificationResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, NotificationResponse{
			ReceiptID:  r.ReceiptID,
			ID:         r.NotifID,
			Title:      r.Title,
			Body:       r.Body,
			SenderName: r.SenderName,
			IsRead:     r.IsRead,
			ReadAt:     r.ReadAt,
			ReceivedAt: r.ReceivedAt,
		})
	}
	helpers.Success(c, http.StatusOK, "notifications fetched", result)
}

// ══════════════════════════════════════════════════════
//  MARK AS READ
// ══════════════════════════════════════════════════════

// MarkAsRead godoc
// @Summary      Mark a notification as read
// @Description  Marks the notification receipt as read for the current user.
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Receipt ID"
// @Success      200  {object}  helpers.APIResponse  "Marked as read"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — not your receipt"
// @Failure      404  {object}  helpers.APIResponse  "Receipt not found"
// @Router       /api/notifications/{id}/read [patch]
func MarkAsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	receiptID := c.Param("id")

	var receipt models.NotificationReceipt
	if err := config.DB.First(&receipt, receiptID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "notification receipt not found")
		return
	}
	if receipt.UserID != userID {
		helpers.Error(c, http.StatusForbidden, "this is not your notification")
		return
	}
	if receipt.IsRead {
		helpers.Success(c, http.StatusOK, "already marked as read", nil)
		return
	}

	now := time.Now()
	if err := config.DB.Model(&receipt).Updates(models.NotificationReceipt{
		IsRead: true,
		ReadAt: &now,
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to mark as read")
		return
	}
	helpers.Success(c, http.StatusOK, "notification marked as read", nil)
}

// ══════════════════════════════════════════════════════
//  NOTIFY PARENTS OF ABSENT STUDENTS — Admin only
// ══════════════════════════════════════════════════════

// NotifyParentsAbsentStudents godoc
// @Summary      Notify parents of absent students (Admin only)
// @Description  Looks up all absences for a date (defaults to today) and sends ONE aggregated email per parent.
// @Description  FIX #4: Parents with multiple absent children now receive correct per-child emails.
// @Description  Previously the parent bundle only stored the last child's name — if two children were absent,
// @Description  only one child's name appeared in the email. Now each child is tracked separately.
// @Tags         notifications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      NotifyAbsencesInput  false  "Date to check (defaults to today)"
// @Success      200   {object}  helpers.APIResponse  "Emails sent"
// @Failure      400   {object}  helpers.APIResponse  "Invalid date format"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/notify/absences [post]
func NotifyParentsAbsentStudents(c *gin.Context) {
	var input NotifyAbsencesInput
	c.ShouldBindJSON(&input)

	// Handle date
	var date string
	if input.Date == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", input.Date); err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
		date = input.Date
	}

	// DB result structure
	type AbsentRecord struct {
		StudentName string
		ParentEmail string
		SubjectName string
		Date        string
	}

	var records []AbsentRecord
	config.DB.Raw(`
		SELECT
			u.name AS student_name,
			COALESCE(NULLIF(pu.email, ''), NULLIF(st.parent_email, '')) AS parent_email,
			s.name AS subject_name,
			TO_CHAR(a.date, 'YYYY-MM-DD') AS date
		FROM attendances a
		JOIN students st ON a.student_id = st.id
		JOIN users u ON st.user_id = u.id
		JOIN subjects s ON a.subject_id = s.id
		LEFT JOIN users pu ON st.parent_id = pu.id AND pu.role = ?
		WHERE a.status = 'Absent'
		  AND DATE(a.date) = DATE(?)
		  AND (
				(st.parent_id IS NOT NULL AND st.parent_id != 0)
				OR (st.parent_email IS NOT NULL AND st.parent_email != '')
		  )
	`, models.RoleParent, date).Scan(&records)

	// parentEmail -> studentName -> subjects
	type childAbsence struct {
		subjects []string
	}

	parentMap := make(map[string]map[string]*childAbsence)

	for _, r := range records {
		if r.ParentEmail == "" {
			continue
		}

		if parentMap[r.ParentEmail] == nil {
			parentMap[r.ParentEmail] = make(map[string]*childAbsence)
		}

		if parentMap[r.ParentEmail][r.StudentName] == nil {
			parentMap[r.ParentEmail][r.StudentName] = &childAbsence{
				subjects: []string{},
			}
		}

		parentMap[r.ParentEmail][r.StudentName].subjects =
			append(parentMap[r.ParentEmail][r.StudentName].subjects, r.SubjectName)
	}

	// Send notifications
	parentsNotified := 0

	for email, children := range parentMap {

		// Build aggregated message parts
		var studentSummaries []string

		for studentName, data := range children {
			subjectList := strings.Join(data.subjects, ", ")
			studentSummaries = append(
				studentSummaries,
				studentName+": "+subjectList,
			)
		}

		fullSummary := strings.Join(studentSummaries, " | ")

		// Single email per parent (same helper call preserved)
		helpers.SendAbsenceAlertBulk(
			email,
			"Multiple Students",
			fullSummary,
			date,
		)

		parentsNotified++
	}

	helpers.Success(c, http.StatusOK, "parent notifications sent", gin.H{
		"parents_notified": parentsNotified,
		"date":             date,
	})
}
