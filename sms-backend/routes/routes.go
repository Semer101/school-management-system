package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"sms-backend/controllers"
	_ "sms-backend/docs"
	"sms-backend/middlewares"
	"sms-backend/models"
)

func SetupRoutes(r *gin.Engine) {

	// ── Global middleware ──────────────────────────────
	r.Use(middlewares.SecurityHeaders())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RequestLogger())
	r.Use(middlewares.RateLimitAPI())
	r.Use(gin.Recovery())

	// ── Swagger UI ────────────────────────────────────
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))
	r.StaticFile("/docs", "./static/swagger-ui.html")

	// ── Static pages ──────────────────────────────────
	r.StaticFile("/notifications", "./static/notifications.html")

	// ── Health check ─────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "SMS Backend"})
	})

	api := r.Group("/api")

	// ── Public routes — NO auth required ──────────────
	api.POST("/login", middlewares.RateLimitLogin(), controllers.Login)
	api.POST("/forgot-password", controllers.ForgotPassword)
	api.POST("/reset-password", controllers.ResetPasswordWithOTP)

	// token/refresh accepts both the sms_refresh HttpOnly cookie (browser)
	// and a refresh_token in the JSON body (API/mobile clients). No auth header needed.
	api.POST("/token/refresh", controllers.RefreshToken)

	// SSE stream is protected ONLY by SSEAuthMiddleware.
	// It must NOT be inside the auth group (which requires a valid access JWT),
	// because the whole point of the SSE token is to allow the stream to stay open
	// even after the short-lived access JWT expires. Putting it in the auth group
	// caused AuthMiddleware to reject the connection as soon as the access token
	// expired, making the SSE token mechanism non-functional for browser clients.
	api.GET("/notifications/stream",
		middlewares.SSEAuthMiddleware(),
		controllers.StreamNotifications,
	)

	// ── All routes below require a valid JWT ──────────
	auth := api.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.POST("/logout", controllers.Logout)

		// Profile — any authenticated user
		auth.GET("/me", controllers.GetMe)
		auth.PUT("/me/password", controllers.ChangePassword)
		auth.POST("/me/avatar", controllers.UploadAvatar)

		// Notifications — any authenticated user
		auth.GET("/notifications", controllers.GetMyNotifications)
		auth.PATCH("/notifications/:id/read", controllers.MarkAsRead)

		// Issue a short-lived SSE token. Browser clients call this first,
		// then open EventSource with ?sse_token=<token>. The full access JWT is
		// never passed in a URL (would be permanently logged by proxies).
		auth.POST("/notifications/sse-token", controllers.IssueSSEToken)

		// ════════════════════════════════════════════
		//  ADMIN ONLY
		// ════════════════════════════════════════════
		admin := auth.Group("/admin")
		admin.Use(middlewares.RoleMiddleware(models.RoleAdmin))
		{
			admin.POST("/register", controllers.Register)

			admin.GET("/admins", controllers.GetAdmins)
			admin.PUT("/admins/:id", controllers.UpdateAdmin)
			admin.DELETE("/admins/:id", controllers.ArchiveAdmin)
			admin.GET("/parents", controllers.GetParents)
			admin.PUT("/parents/:id", controllers.UpdateParent)
			admin.DELETE("/parents/:id", controllers.ArchiveParent)

			// Student management
			admin.POST("/students", controllers.CreateStudent)
			admin.GET("/students", controllers.GetStudents)
			admin.GET("/students/:id", controllers.GetStudent)
			admin.PUT("/students/:id", controllers.UpdateStudent)
			admin.DELETE("/students/:id", controllers.ArchiveStudent)

			// Teacher management
			admin.POST("/teachers", controllers.CreateTeacher)
			admin.GET("/teachers", controllers.GetTeachers)
			admin.GET("/teachers/:id", controllers.GetTeacher)
			admin.PUT("/teachers/:id", controllers.UpdateTeacher)
			admin.DELETE("/teachers/:id", controllers.ArchiveTeacher)

			// Class management
			// added PUT /classes/:id so homeroom teacher and year can be updated
			// without needing direct DB access.
			admin.POST("/classes", controllers.CreateClass)
			admin.GET("/classes", controllers.GetClasses)
			admin.PUT("/classes/:id", controllers.UpdateClass)
			admin.DELETE("/classes/:id", controllers.ArchiveClass)

			// Subject management
			// added PUT /subjects/:id so subject name, code, and teacher
			// can be updated — critical after archiving a teacher (prevents orphaned subjects).
			admin.POST("/subjects", controllers.CreateSubject)
			admin.GET("/subjects", controllers.GetSubjects)
			admin.PUT("/subjects/:id", controllers.UpdateSubject)
			admin.DELETE("/subjects/:id", controllers.ArchiveSubject)

			// Enrollment
			admin.POST("/enroll", controllers.EnrollStudent)
			// new endpoint — previously enrollments could only be created, never removed.
			admin.DELETE("/unenroll", controllers.UnenrollStudent)

			admin.GET("/finance/summary", controllers.GetAllTransactions)
			admin.PATCH("/finance/receipt/:id/verify", controllers.VerifyReceipt)
			admin.GET("/finance/payroll", controllers.GetPayrolls)
			admin.POST("/finance/payroll", controllers.CreatePayroll)
			admin.PATCH("/finance/payroll/:id/pay", controllers.MarkPayrollPaid)
			admin.GET("/finance/overdue", controllers.GetOverduePayments)
			admin.POST("/finance/remind", controllers.SendPaymentReminder)

			// Notifications
			admin.POST("/notify/broadcast", controllers.BroadcastAnnouncement)
			admin.POST("/notify/absences", controllers.NotifyParentsAbsentStudents)

			// Attendance dashboard
			admin.GET("/attendance/summary", controllers.GetAttendanceSummary)

			// Analytics & dashboard KPIs
			admin.GET("/analytics", controllers.GetAnalyticsSummary)
			admin.GET("/dashboard/kpis", controllers.GetDashboardKPIs)

			// Trash (soft-deleted records)
			admin.GET("/trash", controllers.ListTrash)
			admin.POST("/trash/:entity/:id/restore", controllers.RestoreTrash)
			admin.DELETE("/trash/:entity/:id/permanent", controllers.PermanentDelete)

			// Promotion & enrollment status
			admin.GET("/students/:id/enrollment-status", controllers.GetStudentEnrollmentStatus)
			admin.GET("/students/:id/promotion-preview", controllers.CheckPromotionPreview)
			admin.POST("/students/:id/promote", controllers.PromoteStudent)

			// Locker — admin read-only view for compliance/support
			admin.GET("/locker/student/:studentID", controllers.AdminGetLockerFiles)
		}

		// ════════════════════════════════════════════
		//  TEACHER ONLY
		// ════════════════════════════════════════════
		teacher := auth.Group("/academics")
		teacher.Use(middlewares.RoleMiddleware(models.RoleTeacher))
		{
			teacher.GET("/attendance/class/:classID", controllers.GetClassAttendance)

			teacher.POST("/grades/bulk", controllers.BulkGradeEntry)
			teacher.GET("/grades/subject/:subjectID", controllers.GetSubjectGrades)
		}

		// Shared academic reads — Teacher + Student + Admin
		// FIX #5: GetAttendancePercentage is now here (Teacher + Student + Admin) instead of
		// the Teacher-only group above.  The handler already contains an internal self-ownership
		// guard for the Student role, so students are correctly restricted to their own data.
		// Previously the handler was Teacher-only, making the Student guard permanently dead code.
		// Registering the same path in a separate Student group caused a Gin route conflict panic.
		shared := auth.Group("/academics")
		shared.Use(middlewares.RoleMiddleware(models.RoleTeacher, models.RoleStudent, models.RoleAdmin))
		{
			shared.POST("/attendance", controllers.RecordAttendance)
			shared.GET("/attendance/:studentID", controllers.GetAttendancePercentage)
			shared.GET("/grades/student/:studentID", controllers.GetStudentGrades)
			shared.GET("/reportcard/:studentID", controllers.GetReportCard)
			shared.GET("/reportcard/:studentID/pdf", controllers.DownloadReportCard)
		}

		// ════════════════════════════════════════════
		//  DIGITAL LOCKER
		// ════════════════════════════════════════════
		studentLocker := auth.Group("/locker")
		studentLocker.Use(middlewares.RoleMiddleware(models.RoleStudent))
		{
			studentLocker.POST("/upload", controllers.UploadLockerFile)
			studentLocker.GET("/my-files", controllers.GetMyLockerFiles)
			studentLocker.DELETE("/files/:fileID", controllers.DeleteLockerFile)
			studentLocker.PATCH("/files/:fileID/visibility", controllers.ToggleFileVisibility)
		}

		teacherLocker := auth.Group("/locker")
		teacherLocker.Use(middlewares.RoleMiddleware(models.RoleTeacher))
		{
			teacherLocker.GET("/student/:studentID/public", controllers.TeacherGetPublicFiles)
		}

		// ════════════════════════════════════════════
		//  PARENT ONLY
		// ════════════════════════════════════════════
		parent := auth.Group("/parent")
		parent.Use(middlewares.RoleMiddleware(models.RoleParent))
		{
			parent.GET("/children", controllers.GetMyChildren)

			parent.GET("/attendance/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetAttendancePercentage,
			)
			parent.GET("/grades/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetStudentGrades,
			)
			parent.GET("/reportcard/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetReportCard,
			)
			parent.GET("/reportcard/:studentID/pdf",
				middlewares.ParentOwnsStudent(),
				controllers.DownloadReportCard,
			)

			parent.POST("/finance/receipt", controllers.SubmitBankReceipt)
			parent.GET("/finance/transactions", controllers.GetMyTransactions)
		}

		// ════════════════════════════════════════════
		//  STUDENT FINANCE
		// ════════════════════════════════════════════
		studentFinance := auth.Group("/finance")
		studentFinance.Use(middlewares.RoleMiddleware(models.RoleStudent))
		{
			studentFinance.POST("/receipt", controllers.SubmitBankReceipt)
			studentFinance.GET("/transactions", controllers.GetMyTransactions)
		}
	}
}
