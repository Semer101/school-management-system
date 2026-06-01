package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// autoEnrollStudentSubjects enrolls a student in all subjects for their stream and grade
func autoEnrollStudentSubjects(tx *gorm.DB, student *models.Student) error {
	if student.Stream == "" || student.GradeLevel < 9 {
		return nil
	}
	codes := models.SubjectCodesForStreamGrade(student.Stream, student.GradeLevel)
	var subjects []models.Subject
	if err := tx.Where("code IN ? AND (grade_level = ? OR grade_level = 0)", codes, student.GradeLevel).Find(&subjects).Error; err != nil {
		return err
	}
	for _, sub := range subjects {
		if sub.Stream != "" && sub.Stream != student.Stream {
			continue
		}
		var count int64
		tx.Model(&models.Enrollment{}).Where("student_id = ? AND subject_id = ?", student.ID, sub.ID).Count(&count)
		if count == 0 {
			if err := tx.Create(&models.Enrollment{StudentID: student.ID, SubjectID: sub.ID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// evaluatePromotion checks previous-year grades and returns promotion status + failed count
func evaluatePromotion(studentID uint, academicYear int) (status string, failed int, failedSubjectIDs []uint, err error) {
	var grades []models.Grade
	q := config.DB.Where("student_id = ?", studentID)
	if academicYear > 0 {
		q = q.Where("academic_year = ?", academicYear)
	}
	if err = q.Find(&grades).Error; err != nil {
		return "", 0, nil, err
	}
	// Average score per subject (best attempt)
	subjectScores := map[uint][]float64{}
	for _, g := range grades {
		pct := (g.Score / g.MaxScore) * 100
		subjectScores[g.SubjectID] = append(subjectScores[g.SubjectID], pct)
	}
	for subID, scores := range subjectScores {
		var sum float64
		for _, s := range scores {
			sum += s
		}
		avg := sum / float64(len(scores))
		if avg < 50 {
			failed++
			failedSubjectIDs = append(failedSubjectIDs, subID)
		}
	}
	switch {
	case failed >= 3:
		status = models.PromotionRepeat
	case failed >= 1:
		status = models.PromotionConditional
	default:
		status = models.PromotionNormal
	}
	return status, failed, failedSubjectIDs, nil
}

// PromoteStudent godoc — checks grades and promotes / auto-enrolls for next year
func PromoteStudent(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var student models.Student
	if err := config.DB.Preload("User").First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	prevYear := student.AcademicYear
	status, failed, failedSubjectIDs, err := evaluatePromotion(student.ID, prevYear)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to evaluate grades")
		return
	}

	if status == models.PromotionRepeat {
		txErr := config.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&student).Updates(map[string]any{
				"promotion_status": models.PromotionRepeat,
				"academic_year":    prevYear + 1,
			}).Error; err != nil {
				return err
			}
			student.PromotionStatus = models.PromotionRepeat
			student.AcademicYear = prevYear + 1
			// Repeat current grade: reset enrollments for current subjects
			tx.Where("student_id = ?", student.ID).Delete(&models.Enrollment{})
			return autoEnrollStudentSubjects(tx, &student)
		})
		if txErr != nil {
			helpers.Error(c, http.StatusInternalServerError, "repetition reset failed: "+txErr.Error())
			return
		}
		helpers.Success(c, http.StatusOK, fmt.Sprintf("%d subject(s) below 50%%. Student must repeat Grade %d.", failed, student.GradeLevel), gin.H{
			"student":          student,
			"promotion_status": status,
			"failed_subjects":  failed,
		})
		return
	}

	nextGrade := student.GradeLevel
	if status == models.PromotionNormal || status == models.PromotionConditional {
		if student.GradeLevel < 12 {
			nextGrade = student.GradeLevel + 1
		}
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"grade_level":      nextGrade,
			"promotion_status": status,
			"academic_year":    prevYear + 1,
		}
		if err := tx.Model(&student).Updates(updates).Error; err != nil {
			return err
		}
		student.GradeLevel = nextGrade
		student.AcademicYear = prevYear + 1
		student.PromotionStatus = status
		// Remove old enrollments and re-enroll for new grade
		tx.Where("student_id = ?", student.ID).Delete(&models.Enrollment{})
		if err := autoEnrollStudentSubjects(tx, &student); err != nil {
			return err
		}
		// Enroll in failed subjects as retakes
		for _, fsID := range failedSubjectIDs {
			if err := tx.Create(&models.Enrollment{StudentID: student.ID, SubjectID: fsID}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "promotion failed: "+txErr.Error())
		return
	}

	msg := "Promoted and enrolled in all stream subjects."
	if status == models.PromotionConditional {
		msg = fmt.Sprintf("Conditionally promoted (%d failed subject(s)). Retake classes assigned. Enrolled in new year subjects.", failed)
	}

	config.DB.Preload("User").Preload("Class").First(&student, student.ID)
	helpers.Success(c, http.StatusOK, msg, gin.H{
		"student":          student,
		"promotion_status": status,
		"failed_subjects":  failed,
	})
}

// GetStudentEnrollmentStatus returns per-subject enrollment for a student
func GetStudentEnrollmentStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	codes := models.SubjectCodesForStreamGrade(student.Stream, student.GradeLevel)
	var subjects []models.Subject
	config.DB.Where("code IN ?", codes).Find(&subjects)

	var enrolled []models.Enrollment
	config.DB.Where("student_id = ?", student.ID).Find(&enrolled)
	enrolledMap := map[uint]bool{}
	for _, e := range enrolled {
		enrolledMap[e.SubjectID] = true
	}

	type row struct {
		SubjectID   uint   `json:"subject_id"`
		SubjectName string `json:"subject_name"`
		SubjectCode string `json:"subject_code"`
		Enrolled    bool   `json:"enrolled"`
	}
	var rows []row
	for _, s := range subjects {
		if s.Stream != "" && s.Stream != student.Stream {
			continue
		}
		if s.GradeLevel != 0 && s.GradeLevel != student.GradeLevel {
			continue
		}
		rows = append(rows, row{
			SubjectID: s.ID, SubjectName: s.Name, SubjectCode: s.Code,
			Enrolled: enrolledMap[s.ID],
		})
	}
	helpers.Success(c, http.StatusOK, "enrollment status", rows)
}

// CheckPromotionPreview returns promotion eligibility without applying
func CheckPromotionPreview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}
	status, failed, _, err := evaluatePromotion(student.ID, student.AcademicYear)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "evaluation failed")
		return
	}
	helpers.Success(c, http.StatusOK, "promotion preview", gin.H{
		"promotion_status": status,
		"failed_subjects":  failed,
		"can_promote":      status != models.PromotionRepeat,
		"grade_level":      student.GradeLevel,
		"stream":           student.Stream,
	})
}
