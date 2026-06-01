package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// GetAnalyticsSummary returns aggregated stats for the analytics dashboard
func GetAnalyticsSummary(c *gin.Context) {
	var studentCount, teacherCount, parentCount int64
	config.DB.Model(&models.Student{}).Count(&studentCount)
	config.DB.Model(&models.Teacher{}).Count(&teacherCount)
	config.DB.Model(&models.User{}).Where("role = ?", models.RoleParent).Count(&parentCount)

	// Students by grade 9-12
	type gradeRow struct {
		Grade int   `json:"grade"`
		Count int64 `json:"count"`
	}
	var byGrade []gradeRow
	for g := 9; g <= 12; g++ {
		var c int64
		config.DB.Model(&models.Student{}).Where("grade_level = ?", g).Count(&c)
		byGrade = append(byGrade, gradeRow{Grade: g, Count: c})
	}

	// Students by stream — only relevant for grades 11-12 (9-10 are common)
	type streamRow struct {
		Stream string `json:"stream"`
		Count  int64  `json:"count"`
	}
	var byStream []streamRow
	config.DB.Model(&models.Student{}).Select("stream, count(*) as count").
		Where("stream != '' AND grade_level >= 11").Group("stream").Scan(&byStream)

	// Grade averages per subject
	type subjectAvg struct {
		SubjectName string  `json:"subject_name"`
		Average     float64 `json:"average"`
	}
	var gradeAvgs []subjectAvg
	config.DB.Model(&models.Grade{}).
		Select("subjects.name as subject_name, AVG(grades.score / grades.max_score * 100) as average").
		Joins("JOIN subjects ON subjects.id = grades.subject_id").
		Group("subjects.name").
		Scan(&gradeAvgs)

	// Attendance breakdown
	type attRow struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var attBreakdown []attRow
	config.DB.Model(&models.Attendance{}).Where("subject_id IS NULL").
		Select("status, count(*) as count").Group("status").Scan(&attBreakdown)

	// Finance summary
	var totalRevenue, pendingRevenue float64
	config.DB.Model(&models.Transaction{}).Where("status = ?", "Verified").Select("COALESCE(SUM(amount),0)").Scan(&totalRevenue)
	config.DB.Model(&models.Transaction{}).Where("status = ?", "Pending").Select("COALESCE(SUM(amount),0)").Scan(&pendingRevenue)

	var payrollPaid, payrollPending float64
	config.DB.Model(&models.Payroll{}).Where("status = ?", "Paid").Select("COALESCE(SUM(amount),0)").Scan(&payrollPaid)
	config.DB.Model(&models.Payroll{}).Where("status = ?", "Pending").Select("COALESCE(SUM(amount),0)").Scan(&payrollPending)
	// Promotion status distribution
	type promoRow struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var promoDist []promoRow
	config.DB.Model(&models.Student{}).Select("promotion_status as status, count(*) as count").Group("promotion_status").Scan(&promoDist)

	// Monthly attendance trend (last 6 months)
	type monthAtt struct {
		Month   string `json:"month"`
		Present int64  `json:"present"`
		Absent  int64  `json:"absent"`
	}
	var monthlyAtt []monthAtt
	config.DB.Raw(`
		SELECT TO_CHAR(date, 'Mon') as month,
			SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END) as present,
			SUM(CASE WHEN status = 'Absent' THEN 1 ELSE 0 END) as absent
		FROM attendances
		WHERE subject_id IS NULL AND date >= NOW() - INTERVAL '6 months'
		GROUP BY TO_CHAR(date, 'Mon'), EXTRACT(MONTH FROM date)
		ORDER BY EXTRACT(MONTH FROM date)
	`).Scan(&monthlyAtt)

	var notifCount int64
	config.DB.Model(&models.Notification{}).Count(&notifCount)

	helpers.Success(c, http.StatusOK, "analytics", gin.H{
		"kpis": gin.H{
			"students":        studentCount,
			"teachers":        teacherCount,
			"parents":         parentCount,
			"notifications":   notifCount,
			"revenue_etb":     totalRevenue,
			"pending_etb":     pendingRevenue,
			"payroll_paid":    payrollPaid,
			"payroll_pending": payrollPending,
		},
		"students_by_grade":      byGrade,
		"students_by_stream":     byStream,
		"grade_averages":         gradeAvgs,
		"attendance_breakdown":   attBreakdown,
		"monthly_attendance":     monthlyAtt,
		"promotion_distribution": promoDist,
	})
}

// GetDashboardKPIs returns quick stats for the dashboard
func GetDashboardKPIs(c *gin.Context) {
	var students, teachers, classes, subjects int64
	config.DB.Model(&models.Student{}).Count(&students)
	config.DB.Model(&models.Teacher{}).Count(&teachers)
	config.DB.Model(&models.Class{}).Count(&classes)
	config.DB.Model(&models.Subject{}).Count(&subjects)

	var presentToday, absentToday int64
	config.DB.Raw(`
		SELECT COUNT(DISTINCT student_id) FROM attendances
		WHERE subject_id IS NULL AND DATE(date) = CURRENT_DATE AND status = 'Present'
	`).Scan(&presentToday)
	config.DB.Raw(`
		SELECT COUNT(DISTINCT student_id) FROM attendances
		WHERE subject_id IS NULL AND DATE(date) = CURRENT_DATE AND status = 'Absent'
	`).Scan(&absentToday)

	var pendingTx int64
	config.DB.Model(&models.Transaction{}).Where("status = ?", "Pending").Count(&pendingTx)

	helpers.Success(c, http.StatusOK, "dashboard kpis", gin.H{
		"students": students, "teachers": teachers,
		"classes": classes, "subjects": subjects,
		"present_today": presentToday, "absent_today": absentToday,
		"pending_transactions": pendingTx,
	})
}
