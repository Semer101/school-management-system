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
	r.Use(middlewares.SecurityHeaders())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RequestLogger())
	r.Use(middlewares.RateLimitAPI())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))
	r.StaticFile("/docs", "./static/swagger-ui.html")
	r.StaticFile("/notifications", "./static/notifications.html")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "SMS Backend"})
	})

	api := r.Group("/api")

	api.POST("/login", middlewares.RateLimitLogin(), controllers.Login)
	api.POST("/forgot-password", controllers.ForgotPassword)
	api.POST("/reset-password", controllers.ResetPasswordWithOTP)
	api.POST("/token/refresh", controllers.RefreshToken)

	api.GET("/notifications/stream",
		middlewares.SSEAuthMiddleware(),
		controllers.StreamNotifications,
	)

	auth := api.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.POST("/logout", controllers.Logout)
		auth.GET("/me", controllers.GetMe)
		auth.PUT("/me/password", controllers.ChangePassword)
		auth.POST("/me/avatar", controllers.UploadAvatar)
		auth.GET("/notifications", controllers.GetMyNotifications)
		auth.PATCH("/notifications/:id/read", controllers.MarkAsRead)
		auth.POST("/notifications/sse-token", controllers.IssueSSEToken)

		// ADMIN ONLY
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
			admin.POST("/students", controllers.CreateStudent)
			admin.GET("/students", controllers.GetStudents)
			admin.GET("/students/:id", controllers.GetStudent)
			admin.PUT("/students/:id", controllers.UpdateStudent)
			admin.DELETE("/students/:id", controllers.ArchiveStudent)
			admin.POST("/teachers", controllers.CreateTeacher)
			admin.GET("/teachers", controllers.GetTeachers)
			admin.GET("/teachers/:id", controllers.GetTeacher)
			admin.PUT("/teachers/:id", controllers.UpdateTeacher)
			admin.DELETE("/teachers/:id", controllers.ArchiveTeacher)
			admin.POST("/classes", controllers.CreateClass)
			admin.GET("/classes", controllers.GetClasses)
			admin.PUT("/classes/:id", controllers.UpdateClass)
			admin.DELETE("/classes/:id", controllers.ArchiveClass)
			admin.POST("/subjects", controllers.CreateSubject)
			admin.GET("/subjects", controllers.GetSubjects)
			admin.PUT("/subjects/:id", controllers.UpdateSubject)
			admin.DELETE("/subjects/:id", controllers.ArchiveSubject)
			admin.POST("/enroll", controllers.EnrollStudent)
			admin.DELETE("/unenroll", controllers.UnenrollStudent)
			admin.GET("/finance/summary", controllers.GetAllTransactions)
			admin.PATCH("/finance/receipt/:id/verify", controllers.VerifyReceipt)
			admin.GET("/finance/payroll", controllers.GetPayrolls)
			admin.POST("/finance/payroll", controllers.CreatePayroll)
			admin.PATCH("/finance/payroll/:id/pay", controllers.MarkPayrollPaid)
			admin.GET("/finance/overdue", controllers.GetOverduePayments)
			admin.POST("/finance/remind", controllers.SendPaymentReminder)
			// Receipt moderation
			admin.GET("/finance/receipts", controllers.ListPendingReceipts)
			admin.PATCH("/finance/receipts/:id/approve", controllers.ApproveReceipt)
			admin.PATCH("/finance/receipts/:id/reject", controllers.RejectReceipt)
			admin.POST("/notify/broadcast", controllers.BroadcastAnnouncement)
			admin.POST("/notify/absences", controllers.NotifyParentsAbsentStudents)
			admin.GET("/attendance/summary", controllers.GetAttendanceSummary)
			admin.GET("/analytics", controllers.GetAnalyticsSummary)
			admin.GET("/dashboard/kpis", controllers.GetDashboardKPIs)
			admin.GET("/trash", controllers.ListTrash)
			admin.POST("/trash/:entity/:id/restore", controllers.RestoreTrash)
			admin.DELETE("/trash/:entity/:id/permanent", controllers.PermanentDelete)
			admin.GET("/students/:id/enrollment-status", controllers.GetStudentEnrollmentStatus)
			admin.GET("/students/:id/promotion-preview", controllers.CheckPromotionPreview)
			admin.POST("/students/:id/promote", controllers.PromoteStudent)
			admin.GET("/locker/student/:studentID", controllers.AdminGetLockerFiles)
		}

		// TEACHER ONLY
		teacher := auth.Group("/academics")
		teacher.Use(middlewares.RoleMiddleware(models.RoleTeacher))
		{
			teacher.GET("/attendance/class/:classID", controllers.GetClassAttendance)
			teacher.POST("/grades/bulk", controllers.BulkGradeEntry)
			teacher.GET("/grades/subject/:subjectID", controllers.GetSubjectGrades)
			teacher.GET("/dashboard/kpis", controllers.GetTeacherKPIs)
		}

		// SHARED
		shared := auth.Group("/academics")
		shared.Use(middlewares.RoleMiddleware(models.RoleTeacher, models.RoleStudent, models.RoleAdmin))
		{
			shared.POST("/attendance", controllers.RecordAttendance)
			shared.GET("/attendance/:studentID", controllers.GetAttendancePercentage)
			shared.GET("/grades/student/:studentID", controllers.GetStudentGrades)
			shared.GET("/reportcard/:studentID", controllers.GetReportCard)
			shared.GET("/reportcard/:studentID/pdf", controllers.DownloadReportCard)
		}

		// LOCKER
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

		// PARENT ONLY
		parent := auth.Group("/parent")
		parent.Use(middlewares.RoleMiddleware(models.RoleParent))
		{
			parent.GET("/children", controllers.GetMyChildren)
			parent.GET("/attendance/:studentID", middlewares.ParentOwnsStudent(), controllers.GetAttendancePercentage)
			parent.GET("/grades/:studentID", middlewares.ParentOwnsStudent(), controllers.GetStudentGrades)
			parent.GET("/reportcard/:studentID", middlewares.ParentOwnsStudent(), controllers.GetReportCard)
			parent.GET("/reportcard/:studentID/pdf", middlewares.ParentOwnsStudent(), controllers.DownloadReportCard)
			parent.POST("/finance/receipt", controllers.SubmitBankReceipt)
			parent.GET("/finance/transactions", controllers.GetMyTransactions)
			// Receipt image upload + viewing
			parent.POST("/finance/receipts", controllers.UploadPaymentReceipt)
			parent.GET("/finance/receipts/:id", controllers.GetReceipt)
		}
	}
}