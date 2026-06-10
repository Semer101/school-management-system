package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/docs"
	"sms-backend/models"
	"sms-backend/routes"
)

// @title           SMS Backend API
// @version         1.1
// @description     School Management System — Go + Gin + PostgreSQL
// @description
// @description     ## Authentication
// @description     This API uses JWT Bearer tokens. Login via `POST /api/login` to receive an access token (1 hour) and a refresh token (7 days).
// @description     Click the **Authorize** button above and enter: `Bearer <your_token>`
// @description
// @description     ## Roles
// @description     | Role    | Access                                              |
// @description     |---------|-----------------------------------------------------|
// @description     | Admin   | Full access to all endpoints                        |
// @description     | Teacher | Attendance, grades, locker (read public)             |
// @description     | Student | Own grades, report card, locker, finance             |
// @description     | Parent  | Children's data, report cards, finance               |

// @host school-management-system-70z3.onrender.com
// @BasePath /
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Type "Bearer" followed by a space and your JWT token. Example: **Bearer eyJhbGci...**
func main() {
	// ── Swagger host/scheme: override at runtime so it works on Render ──────
	// Render automatically sets RENDER_EXTERNAL_HOSTNAME (e.g. school-management-system-70z3.onrender.com)
	if host := os.Getenv("RENDER_EXTERNAL_HOSTNAME"); host != "" {
		docs.SwaggerInfo.Host = host
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		// Local development fallback
		docs.SwaggerInfo.Host = "localhost:8080"
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// ── Connect DB ───────────────────────────────────────────────────────────
	config.ConnectDB()

	if config.DB == nil {
		log.Fatal("Database not initialized")
	}

	// 🚨 RUN MIGRATIONS ONLY IN DEVELOPMENT
	if os.Getenv("ENV") != "production" && os.Getenv("SKIP_MIGRATE") != "true" {
		// Safe local development only
		// NEVER runs on Render
		log.Println("Running AutoMigrate (dev mode only)")

		err := config.DB.AutoMigrate(
			&models.User{},
			&models.Class{},
			&models.Subject{},
			&models.Student{},
			&models.Teacher{},
			&models.Enrollment{},
			&models.Attendance{},
			&models.Grade{},
			&models.LockerFile{},
			&models.Transaction{},
			&models.Payroll{},
			&models.Notification{},
			&models.NotificationReceipt{},
			&models.RefreshToken{},
			&models.PasswordResetOTP{},
		)

		if err != nil {
			log.Fatal("Migration failed: ", err)
		}
	} else if os.Getenv("SKIP_MIGRATE") == "true" {
		log.Println("Skipping AutoMigrate (SKIP_MIGRATE=true)")
	}

	log.Println("Database ready")

	// FIX #15: Respect UPLOAD_DIR env var — same default as locker_ctrl.go's getUploadDir().
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads/locker"
	}
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		log.Fatal("Failed to create uploads directory: ", err)
	}
	avatarDir := os.Getenv("AVATAR_DIR")
	if avatarDir == "" {
		avatarDir = "./uploads/avatars"
	}
	if err := os.MkdirAll(avatarDir, 0750); err != nil {
		log.Fatal("Failed to create avatars directory: ", err)
	}
	// Receipt uploads (parent payment receipts) live in their own subdir
	// so we can apply different size limits and access control than the locker.
	receiptDir := os.Getenv("RECEIPT_DIR")
	if receiptDir == "" {
		receiptDir = "./uploads/receipts"
	}
	if err := os.MkdirAll(receiptDir, 0750); err != nil {
		log.Fatal("Failed to create receipts directory: ", err)
	}


	r := gin.Default()
	r.Static("/uploads", "./uploads")
	r.SetTrustedProxies(nil)

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("SMS server running on :%s\n", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
