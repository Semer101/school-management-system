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

func subjectsForGradeStream(db *gorm.DB, gradeLevel int, stream string) *gorm.DB {
	q := db.Where("grade_level = ? AND status = ?", gradeLevel, "Active")
	if gradeLevel >= 11 {
		return q.Where("stream = '' OR stream IS NULL OR stream = ?", stream)
	}
	return q.Where("stream = '' OR stream IS NULL")
}

// autoEnrollStudentSubjects enrolls a student in all active subjects for their grade/stream.
func autoEnrollStudentSubjects(tx *gorm.DB, student *models.Student) error {
	if student.GradeLevel < 9 {
		return nil
	}
	codes := models.SubjectCodesForStreamGrade(student.Stream, student.GradeLevel)
	var subjects []models.Subject
	if err := subjectsForGradeStream(tx, student.GradeLevel, student.Stream).
		Where("code IN ?", codes).
		Order("name ASC").
		Find(&subjects).Error; err != nil {
		return err
	}
	for _, sub := range subjects {
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

// validateStudentPromotion verifies that the student is eligible to be evaluated for promotion
//
// FIX #4: returns (isIneligible bool, err error) so callers can map the "not yet eligible" case to
// HTTP 409 instead of HTTP 400. Previously the only error path was 400, conflating "bad client
// request" with "student's grade data isn't complete yet".
func validateStudentPromotion(student *models.Student) (bool, error) {
	// 1. Checks current academic year and enrollment year.
	enrollYear := student.EnrolledAt.Year()
	if enrollYear > student.AcademicYear {
		return false, fmt.Errorf("promotion blocked: enrollment year (%d) cannot be after the current academic year (%d)", enrollYear, student.AcademicYear)
	}

	// 2. Fetch all enrolled subjects for this student
	var enrollments []models.Enrollment
	if err := config.DB.Where("student_id = ?", student.ID).Find(&enrollments).Error; err != nil {
		return false, fmt.Errorf("failed to retrieve student enrollments: %w", err)
	}

	if len(enrollments) == 0 {
		return false, fmt.Errorf("promotion blocked: student is not enrolled in any subjects for the current academic year")
	}

	// 3. Fetch all grades for this student and academic year
	var grades []models.Grade
	if err := config.DB.Where("student_id = ? AND academic_year = ?", student.ID, student.AcademicYear).Find(&grades).Error; err != nil {
		return false, fmt.Errorf("failed to retrieve grades: %w", err)
	}

	// Group student grades by subject, then by semester
	subjectSemesters := make(map[uint]map[string]bool)
	for _, enrollment := range enrollments {
		subjectSemesters[enrollment.SubjectID] = make(map[string]bool)
	}

	for _, g := range grades {
		if _, exists := subjectSemesters[g.SubjectID]; exists {
			subjectSemesters[g.SubjectID][g.Semester] = true
		}
	}

	// For each enrolled subject, check if Semester 1, Semester 2, and Semester 3 are completed.
	for _, enrollment := range enrollments {
		var subject models.Subject
		if err := config.DB.First(&subject, enrollment.SubjectID).Error; err != nil {
			return false, fmt.Errorf("failed to retrieve subject details for ID %d: %w", enrollment.SubjectID, err)
		}

		sems := subjectSemesters[enrollment.SubjectID]
		if !sems["Semester 1"] || !sems["Semester 2"] || !sems["Semester 3"] {
			var missing []string
			if !sems["Semester 1"] {
				missing = append(missing, "Semester 1")
			}
			if !sems["Semester 2"] {
				missing = append(missing, "Semester 2")
			}
			if !sems["Semester 3"] {
				missing = append(missing, "Semester 3")
			}
			// isIneligible = true so the caller can return 409 (not 400).
			return true, fmt.Errorf("promotion blocked: student is missing grades for %v in subject %s", missing, subject.Name)
		}
	}

	return false, nil
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

	// Group scores by subject, and then by semester
	// subjectID -> semester -> slice of percentage scores
	subjectSemScores := make(map[uint]map[string][]float64)

	for _, g := range grades {
		if subjectSemScores[g.SubjectID] == nil {
			subjectSemScores[g.SubjectID] = make(map[string][]float64)
		}
		pct := (g.Score / g.MaxScore) * 100
		subjectSemScores[g.SubjectID][g.Semester] = append(subjectSemScores[g.SubjectID][g.Semester], pct)
	}

	// For each subject, calculate Sem1, Sem2, and Sem3 averages, and then the overall average.
	for subID, sems := range subjectSemScores {
		var sem1Sum, sem2Sum, sem3Sum float64
		var sem1Count, sem2Count, sem3Count float64

		for _, s := range sems["Semester 1"] {
			sem1Sum += s
			sem1Count++
		}
		for _, s := range sems["Semester 2"] {
			sem2Sum += s
			sem2Count++
		}
		for _, s := range sems["Semester 3"] {
			sem3Sum += s
			sem3Count++
		}

		sem1Avg := 0.0
		if sem1Count > 0 {
			sem1Avg = sem1Sum / sem1Count
		}
		sem2Avg := 0.0
		if sem2Count > 0 {
			sem2Avg = sem2Sum / sem2Count
		}
		sem3Avg := 0.0
		if sem3Count > 0 {
			sem3Avg = sem3Sum / sem3Count
		}

		// Final score is the average of the completed semesters (Sem 1 + Sem 2 + Sem 3).
		var count float64
		var sum float64
		if sem1Count > 0 {
			sum += sem1Avg
			count++
		}
		if sem2Count > 0 {
			sum += sem2Avg
			count++
		}
		if sem3Count > 0 {
			sum += sem3Avg
			count++
		}

		avg := 0.0
		if count > 0 {
			avg = sum / count
		}

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

	// Run validation first
	if ineligible, err := validateStudentPromotion(&student); err != nil {
		if ineligible {
			// 409 Conflict: student isn't yet eligible (e.g. grades missing)
			helpers.Error(c, http.StatusConflict, err.Error())
		} else {
			helpers.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	prevYear := student.AcademicYear
	status, failed, failedSubjectIDs, err := evaluatePromotion(student.ID, prevYear)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to evaluate grades")
		return
	}

	if status == models.PromotionRepeat {
		// FIX #3: do NOT bump academic_year for a repeating student.
		// A repeat means "stay in the same year and re-attempt the failed subjects".
		// Bumping the year would (a) misrepresent the year they're in, (b) cause
		// report cards to be misattributed, and (c) invalidate the
		// enrollYear > academicYear guard in validateStudentPromotion.
		txErr := config.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&student).Updates(map[string]any{
				"promotion_status": models.PromotionRepeat,
				// academic_year intentionally NOT changed
			}).Error; err != nil {
				return err
			}
			student.PromotionStatus = models.PromotionRepeat
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
			// Skip duplicates if a retake subject is already in the new grade's
			// subject set (FIX: prevents unique-constraint failures for some curricula).
			var existing int64
			if err := tx.Model(&models.Enrollment{}).
				Where("student_id = ? AND subject_id = ?", student.ID, fsID).
				Count(&existing).Error; err != nil {
				return err
			}
			if existing == 0 {
				if err := tx.Create(&models.Enrollment{StudentID: student.ID, SubjectID: fsID}).Error; err != nil {
					return err
				}
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

	var subjects []models.Subject
	config.DB.Scopes(func(db *gorm.DB) *gorm.DB {
		return subjectsForGradeStream(db, student.GradeLevel, student.Stream)
	}).Order("name ASC").Find(&subjects)

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

	// Run validation first
	if ineligible, err := validateStudentPromotion(&student); err != nil {
		if ineligible {
			helpers.Error(c, http.StatusConflict, err.Error())
		} else {
			helpers.Error(c, http.StatusBadRequest, err.Error())
		}
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
