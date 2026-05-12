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

// SetupRoutes wires ALL routes
func SetupRoutes(r *gin.Engine) {

	// ── Global middleware ──────────────────────────────
	r.Use(middlewares.SecurityHeaders())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RequestLogger())
	r.Use(middlewares.RateLimitAPI())
	r.Use(gin.Recovery()) // catches panics, returns 500

	// ── Swagger ────────────────────────────────────────
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.StaticFile("/docs", "./static/swagger-ui.html")

	// ── Health check ──────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "SMS Backend"})
	})

	api := r.Group("/api")

	// ── Public routes ──────────────────────────────────
	api.POST("/register", controllers.Register)
	api.POST("/login", middlewares.RateLimitLogin(), controllers.Login)

	// ── Protected routes ──────────────────────────────
	auth := api.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		// Any logged-in user
		auth.GET("/me", controllers.GetMe)
		auth.PUT("/me/password", controllers.ChangePassword)

		// ── Admin only ─────────────────────────────────
		admin := auth.Group("/admin")
		admin.Use(middlewares.RoleMiddleware(models.RoleAdmin))
		{
			// Student management (FE-04)
			admin.POST("/students", controllers.CreateStudent)
			admin.GET("/students", controllers.GetStudents)
			admin.GET("/students/:id", controllers.GetStudent)
			admin.PUT("/students/:id", controllers.UpdateStudent)
			admin.DELETE("/students/:id", controllers.ArchiveStudent)

			// Teacher management (FE-04)
			admin.POST("/teachers", controllers.CreateTeacher)
			admin.GET("/teachers", controllers.GetTeachers)

			// Class management
			admin.POST("/classes", controllers.CreateClass)
			admin.GET("/classes", controllers.GetClasses)

			// Subject management
			admin.POST("/subjects", controllers.CreateSubject)
			admin.GET("/subjects", controllers.GetSubjects)

			// Enrollment (FE-05)
			admin.POST("/enroll", controllers.EnrollStudent)

			// Finance admin actions
			admin.GET("/finance/summary", controllers.GetTransactions)
			admin.PATCH("/finance/receipt/:id/verify", controllers.VerifyReceipt)
			admin.POST("/finance/payroll", controllers.CreatePayroll)
			admin.PATCH("/finance/payroll/:id/pay", controllers.MarkPayrollPaid)

			// Notifications (FE-19)
			admin.POST("/notify/broadcast", controllers.BroadcastAnnouncement)
			admin.POST("/notify/absences", controllers.NotifyParentsAbsentStudents)

			// Attendance summary dashboard
			admin.GET("/attendance/summary", controllers.GetAttendanceSummary)
		}

		// ── Teacher routes ─────────────────────────────
		teacher := auth.Group("/academics")
		teacher.Use(middlewares.RoleMiddleware(models.RoleTeacher))
		{
			// Attendance (FE-06, FE-07)
			teacher.POST("/attendance", controllers.RecordAttendance)
			teacher.GET("/attendance/class/:classID", controllers.GetClassAttendance)
			teacher.GET("/attendance/:studentID", controllers.GetAttendancePercentage)

			// Grades (FE-09, FE-10, FE-11)
			teacher.POST("/grades/bulk", controllers.BulkGradeEntry)
			teacher.GET("/grades/subject/:subjectID", controllers.GetSubjectGrades)
		}

		// ── Shared: Teacher + Student + Admin can view grades/reports ──
		shared := auth.Group("/academics")
		shared.Use(middlewares.RoleMiddleware(models.RoleTeacher, models.RoleStudent, models.RoleAdmin))
		{
			shared.GET("/grades/student/:studentID", controllers.GetStudentGrades)
			shared.GET("/reportcard/:studentID", controllers.GetReportCard)
			shared.GET("/reportcard/:studentID/pdf", controllers.DownloadReportCard)
		}

		// ── Digital Locker (FE-12, FE-13, FE-14) ──────
		locker := auth.Group("/locker")
		locker.Use(middlewares.RoleMiddleware(models.RoleStudent, models.RoleTeacher, models.RoleAdmin))
		{
			locker.POST("/upload", controllers.UploadLockerFile)
			locker.GET("/student/:studentID", controllers.GetLockerFiles)
			locker.DELETE("/files/:fileID", controllers.DeleteLockerFile)
			locker.PATCH("/files/:fileID/visibility", controllers.ToggleFileVisibility)
		}

		// ── Parent routes (FE-03, FE-07, FE-11) ────────────
		parent := auth.Group("/parent")
		parent.Use(middlewares.RoleMiddleware(models.RoleParent))
		{
			// See which children are linked to this account
			parent.GET("/children", controllers.GetMyChildren)

			// View child's attendance (with ownership check)
			parent.GET("/attendance/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetAttendancePercentage,
			)

			// View child's grades
			parent.GET("/grades/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetStudentGrades,
			)

			// View child's report card
			parent.GET("/reportcard/:studentID",
				middlewares.ParentOwnsStudent(),
				controllers.GetReportCard,
			)

			// Download PDF report card
			parent.GET("/reportcard/:studentID/pdf",
				middlewares.ParentOwnsStudent(),
				controllers.DownloadReportCard,
			)

			parent.POST("/finance/receipt", controllers.SubmitBankReceipt)

			// View payment history
			parent.GET("/finance/transactions", controllers.GetTransactions)
		}

		// ── Finance (FE-20, FE-21) ─────────────────────
		// Student and Admin access.
		finance := auth.Group("/finance")
		finance.Use(middlewares.RoleMiddleware(models.RoleStudent, models.RoleAdmin))
		{
			finance.POST("/receipt", controllers.SubmitBankReceipt)
			finance.GET("/transactions", controllers.GetTransactions)
		}
	}
}
