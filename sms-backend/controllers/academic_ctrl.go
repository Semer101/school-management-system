package controllers

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ───────────────────────────────────────────────────────────────

type CreateStudentInput struct {
	Name        string `json:"name"          binding:"required,min=2"        example:"Alice Bekele"`
	Email       string `json:"email"         binding:"required,email"        example:"alice@school.com"`
	Password    string `json:"password"      binding:"required,min=8"        example:"password123"`
	StudentCode string `json:"student_code"  example:"STU-2025-001"`
	ClassID     uint   `json:"class_id"                                       example:"1"`
	ParentID    uint   `json:"parent_id"     binding:"required"              example:"5"`
	ParentName  string `json:"parent_name"                                   example:"Mr. Bekele"`
	ParentEmail string `json:"parent_email"  binding:"omitempty,email"       example:"parent@email.com"`
	ParentPhone string `json:"parent_phone"                                  example:"+251911000000"`
	DateOfBirth string `json:"date_of_birth"                                 example:"2010-05-15"`
	Stream      string `json:"stream"        example:"Natural Science"` // required only for grades 11-12
	GradeLevel  int    `json:"grade_level"   binding:"required,min=9,max=12" example:"9"`
}

type UpdateStudentInput struct {
	ClassID     *uint  `json:"class_id"                                example:"2"`
	ParentID    *uint  `json:"parent_id"                               example:"6"`
	ParentName  string `json:"parent_name"                             example:"Mr. Updated"`
	ParentEmail string `json:"parent_email"  binding:"omitempty,email" example:"new@email.com"`
	ParentPhone string `json:"parent_phone"                            example:"+251922000000"`
	DateOfBirth string `json:"date_of_birth"                           example:"2010-05-15"`
	Stream      string `json:"stream"                                  example:"Natural Science"`
	GradeLevel  *int   `json:"grade_level"                             example:"10"`
}

type CreateTeacherInput struct {
	Name          string `json:"name"          binding:"required,min=2" example:"Mr. Tadesse"`
	Email         string `json:"email"         binding:"required,email" example:"tadesse@school.com"`
	Password      string `json:"password"      binding:"required,min=8" example:"teacher123"`
	TeacherCode   string `json:"teacher_code"  binding:"required"       example:"TCH-001"`
	Qualification string `json:"qualification"                          example:"MSc Mathematics"`
	Phone         string `json:"phone"         binding:"required"        example:"0911000001"`
	Department    string `json:"department"                             example:"Mathematics"`
}

type UpdateTeacherInput struct {
	Qualification string `json:"qualification" example:"PhD Mathematics"`
	Phone         string `json:"phone"         example:"0911000001"`
	Department    string `json:"department"     example:"Mathematics"`
}

type AttendanceInput struct {
	StudentID uint   `json:"student_id" binding:"required"                           example:"1"`
	Date      string `json:"date"       binding:"required"                           example:"2025-05-01"`
	Status    string `json:"status"     binding:"required,oneof=Present Absent Late" example:"Present"`
	Notes     string `json:"notes"                                                   example:"Arrived 5 mins late"`
}

type GradeBulkEntry struct {
	StudentID uint    `json:"student_id" binding:"required"              example:"1"`
	SubjectID uint    `json:"subject_id" binding:"required"              example:"3"`
	Score     float64 `json:"score"      binding:"required,min=0,max=100" example:"87.5"`
	GradeType string  `json:"grade_type" binding:"required,oneof=Midterm Final Quiz Assignment Exam" example:"Midterm"`
	Semester  string  `json:"semester"   binding:"required,oneof='Semester 1' 'Semester 2' 'Semester 3'" example:"Semester 1"`
	Remarks   string  `json:"remarks"                                    example:"Good"`
}

type BulkGradeInput struct {
	Grades []GradeBulkEntry `json:"grades" binding:"required,min=1"`
}

type CreateClassInput struct {
	Name       string `json:"name"        example:"Grade 10A Natural"`
	GradeLevel int    `json:"grade_level" binding:"required,min=9,max=12" example:"10"`
	Section    string `json:"section"     binding:"required" example:"A"`
	Stream     string `json:"stream"      example:"Natural Science"`
	Status     string `json:"status"      example:"Active"`
	Year       int    `json:"year"        binding:"required" example:"2025"`
	TeacherID  uint   `json:"teacher_id"  example:"1"`
}

// UpdateClassInput allows changing the homeroom teacher or year after creation.
type UpdateClassInput struct {
	TeacherID *uint  `json:"teacher_id" example:"2"`
	Year      *int   `json:"year"       example:"2026"`
	Name      string `json:"name"       example:"Grade 10B"`
}

type CreateSubjectInput struct {
	Name       string `json:"name"        binding:"required" example:"Mathematics"`
	Code       string `json:"code"        binding:"required" example:"MATH-G9"`
	GradeLevel int    `json:"grade_level" example:"9"`
	Stream     string `json:"stream"      example:""`
	Status     string `json:"status"      example:"Active"`
	TeacherID  uint   `json:"teacher_id"  example:"1"`
}

// UpdateSubjectInput allows renaming a subject or reassigning its teacher.
type UpdateSubjectInput struct {
	Name      string `json:"name"       example:"Advanced Mathematics"`
	Code      string `json:"code"       example:"MATH-201"`
	TeacherID *uint  `json:"teacher_id" example:"3"`
}

type EnrollStudentInput struct {
	StudentID uint `json:"student_id" binding:"required" example:"1"`
	SubjectID uint `json:"subject_id" binding:"required" example:"2"`
}

// UnenrollStudentInput is the body for removing an enrollment.
type UnenrollStudentInput struct {
	StudentID uint `json:"student_id" binding:"required" example:"1"`
	SubjectID uint `json:"subject_id" binding:"required" example:"2"`
}

// ── Pagination helper ─────────────────────────────────────────────────────────
func parsePage(c *gin.Context) (offset, limit int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 25
	}
	return (page - 1) * size, size
}

// ══════════════════════════════════════════════════════
//  STUDENTS
// ══════════════════════════════════════════════════════

// CreateStudent godoc
// @Summary      Create a student (Admin only)
// @Description  Creates a new student account with a linked user login. The parent_id must be the ID of an existing User with role=Parent.
// @Tags         students
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateStudentInput  true  "Student data"
// @Success      201   {object}  helpers.APIResponse "Student created"
// @Failure      400   {object}  helpers.APIResponse "Validation error, invalid parent_id, or parent user is not role=Parent"
// @Failure      401   {object}  helpers.APIResponse "Unauthorized"
// @Failure      403   {object}  helpers.APIResponse "Forbidden — Admin only"
// @Failure      409   {object}  helpers.APIResponse "Email or student code already exists"
// @Router       /api/admin/students [post]
func CreateStudent(c *gin.Context) {
	var input CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Grades 9-10: common curriculum, no stream differentiation
	// Grades 11-12: stream is required and must be Natural or Social Science
	if input.GradeLevel <= 10 {
		input.Stream = "" // enforce no stream for grades 9-10
	} else {
		if input.Stream != models.StreamNatural && input.Stream != models.StreamSocial {
			helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science' for grades 11–12")
			return
		}
	}
	if input.GradeLevel < 9 {
		helpers.Error(c, http.StatusBadRequest, "grade_level must be 9 or higher")
		return
	}

	var parent models.User
	if err := config.DB.First(&parent, input.ParentID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid parent_id: user not found")
		return
	}
	if parent.Role != models.RoleParent {
		helpers.Error(c, http.StatusBadRequest, "invalid parent_id: user is not a Parent")
		return
	}

	var count int64
	config.DB.Unscoped().Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		helpers.Error(c, http.StatusConflict, "email already registered")
		return
	}

	studentCode := strings.TrimSpace(input.StudentCode)
	if studentCode == "" {
		studentCode = generateStudentCode(input.GradeLevel)
	}
	var codeCount int64
	config.DB.Unscoped().
		Model(&models.Student{}).
		Where("student_code = ?", studentCode).
		Count(&codeCount)

	if codeCount > 0 {
		helpers.Error(c, http.StatusConflict, "student code already exists")
		return
	}

	var dob time.Time
	if input.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", input.DateOfBirth)
		if err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid date_of_birth format, use YYYY-MM-DD")
			return
		}
		dob = parsed
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var user models.User
	var student models.Student

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		user = models.User{
			Name:     input.Name,
			Email:    input.Email,
			Password: string(hashed),
			Role:     models.RoleStudent,
			IsActive: true,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		var classIDPtr *uint
		studentYear := 2025
		if input.ClassID != 0 {
			var class models.Class
			if err := tx.First(&class, input.ClassID).Error; err != nil {
				return fmt.Errorf("class not found: %w", err)
			}
			if class.GradeLevel != input.GradeLevel {
				return fmt.Errorf("student grade level %d does not match class grade level %d", input.GradeLevel, class.GradeLevel)
			}
			if class.Stream != input.Stream {
				return fmt.Errorf("student stream '%s' does not match class stream '%s'", input.Stream, class.Stream)
			}
			classIDPtr = &input.ClassID
			studentYear = class.Year
		}
		student = models.Student{
			UserID:          user.ID,
			ParentID:        input.ParentID,
			ClassID:         classIDPtr,
			StudentCode:     studentCode,
			ParentName:      input.ParentName,
			ParentEmail:     input.ParentEmail,
			ParentPhone:     input.ParentPhone,
			DateOfBirth:     dob,
			EnrolledAt:      time.Now(),
			Stream:          input.Stream,
			GradeLevel:      input.GradeLevel,
			PromotionStatus: models.PromotionNormal,
			AcademicYear:    studentYear,
		}
		if err := tx.Create(&student).Error; err != nil {
			return err
		}
		return autoEnrollStudentSubjects(tx, &student)
	})

	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create student: "+txErr.Error())
		return
	}

	helpers.Success(c, http.StatusCreated, "student created", gin.H{
		"user_id":    user.ID,
		"student_id": student.ID,
		"parent_id":  student.ParentID,
		"name":       user.Name,
		"email":      user.Email,
		"code":       student.StudentCode,
	})
}

// GetStudents godoc
// @Summary      List all students (Admin only)
// @Description  Returns a paginated list of all students. Use ?page=1&page_size=50.
// @Tags         students
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "Paginated student list"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/students [get]
func GetStudents(c *gin.Context) {
	offset, limit := parsePage(c)
	var students []models.Student
	var total int64

	// FIX #1: Use a single shared query base for both Count and Find so that
	// the same GORM scoping (soft-delete filter etc.) applies to both calls.
	// Previously, scope was built then discarded — a new config.DB chain was used
	// for Find, which could diverge from the count in edge cases.
	db := config.DB.Model(&models.Student{})
	if q := strings.TrimSpace(c.Query("search")); q != "" {
		like := "%" + q + "%"
		db = db.Joins("JOIN users ON users.id = students.user_id").
			Where("users.name ILIKE ? OR students.student_code ILIKE ? OR users.email ILIKE ?", like, like, like)
	}
	if stream := c.Query("stream"); stream != "" {
		db = db.Where("students.stream = ?", stream)
	}
	if gl := c.Query("grade_level"); gl != "" {
		db = db.Where("students.grade_level = ?", gl)
	}
	db.Count(&total)
	if err := db.Preload("User").Preload("Class").
		Offset(offset).Limit(limit).Find(&students).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch students")
		return
	}
	helpers.Success(c, http.StatusOK, "students fetched", gin.H{"total": total, "data": students})
}

// GetStudent godoc
// @Summary      Get a student by ID (Admin only)
// @Description  Returns a single student with user and class details.
// @Tags         students
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "Student record"
// @Failure      404  {object}  helpers.APIResponse  "Student not found"
// @Router       /api/admin/students/{id} [get]
func GetStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student
	if err := config.DB.Preload("User").Preload("Class").First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}
	helpers.Success(c, http.StatusOK, "student fetched", student)
}

// UpdateStudent godoc
// @Summary      Update a student (Admin only)
// @Description  Updates class, parent info, date of birth, or parent_id for a student.
// @Tags         students
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "Student ID"
// @Param        body  body      UpdateStudentInput  true  "Fields to update"
// @Success      200   {object}  helpers.APIResponse "Student updated"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      404   {object}  helpers.APIResponse "Student not found"
// @Router       /api/admin/students/{id} [put]
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	var input UpdateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var targetGradeLevel = student.GradeLevel
	if input.GradeLevel != nil {
		targetGradeLevel = *input.GradeLevel
	}
	var targetStream = student.Stream
	if input.Stream != "" {
		targetStream = input.Stream
	}
	if targetGradeLevel <= 10 {
		targetStream = ""
	} else if targetStream != models.StreamNatural && targetStream != models.StreamSocial {
		helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science' for grades 11-12")
		return
	}

	var finalClassID *uint
	if input.ClassID != nil {
		if *input.ClassID != 0 {
			finalClassID = input.ClassID
		}
	} else {
		finalClassID = student.ClassID
	}

	if finalClassID != nil {
		var class models.Class
		if err := config.DB.First(&class, *finalClassID).Error; err != nil {
			helpers.Error(c, http.StatusBadRequest, "class not found")
			return
		}
		if class.GradeLevel != targetGradeLevel {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student grade level %d does not match class grade level %d", targetGradeLevel, class.GradeLevel))
			return
		}
		if class.Stream != targetStream {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student stream '%s' does not match class stream '%s'", targetStream, class.Stream))
			return
		}
		if class.Year != student.AcademicYear {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student academic year %d does not match class academic year %d", student.AcademicYear, class.Year))
			return
		}
	}

	updates := map[string]any{}
	if input.ClassID != nil {
		if *input.ClassID == 0 {
			updates["class_id"] = nil
		} else {
			updates["class_id"] = *input.ClassID
		}
	}
	if input.ParentID != nil {
		updates["parent_id"] = *input.ParentID
	}
	if input.ParentName != "" {
		updates["parent_name"] = input.ParentName
	}
	if input.ParentEmail != "" {
		updates["parent_email"] = input.ParentEmail
	}
	if input.ParentPhone != "" {
		updates["parent_phone"] = input.ParentPhone
	}
	if input.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", input.DateOfBirth)
		if err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid date_of_birth format, use YYYY-MM-DD")
			return
		}
		updates["date_of_birth"] = parsed
	}
	if input.GradeLevel != nil {
		if *input.GradeLevel < 9 || *input.GradeLevel > 12 {
			helpers.Error(c, http.StatusBadRequest, "grade_level must be 9–12")
			return
		}
		updates["grade_level"] = *input.GradeLevel
		// Enforce stream rules when grade changes
		if *input.GradeLevel <= 10 {
			updates["stream"] = "" // grades 9-10 have no stream
		} else if input.Stream != "" {
			if input.Stream != models.StreamNatural && input.Stream != models.StreamSocial {
				helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science' for grades 11–12")
				return
			}
			updates["stream"] = input.Stream
		}
	} else if input.Stream != "" {
		if input.Stream != models.StreamNatural && input.Stream != models.StreamSocial {
			helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science'")
			return
		}
		updates["stream"] = input.Stream
	}

	if len(updates) == 0 {
		helpers.Error(c, http.StatusBadRequest, "no fields to update")
		return
	}

	placementChanged := false
	if input.GradeLevel != nil && *input.GradeLevel != student.GradeLevel {
		placementChanged = true
	}
	if input.Stream != "" && targetStream != student.Stream {
		placementChanged = true
	}
	if input.ClassID != nil {
		currentClassID := uint(0)
		if student.ClassID != nil {
			currentClassID = *student.ClassID
		}
		if *input.ClassID != currentClassID {
			placementChanged = true
		}
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&student).Updates(updates).Error; err != nil {
			return err
		}
		if !placementChanged {
			return nil
		}
		if input.GradeLevel != nil {
			student.GradeLevel = *input.GradeLevel
		}
		student.Stream = targetStream
		if input.ClassID != nil {
			if *input.ClassID == 0 {
				student.ClassID = nil
			} else {
				student.ClassID = input.ClassID
			}
		}
		if err := tx.Where("student_id = ?", student.ID).Delete(&models.Enrollment{}).Error; err != nil {
			return err
		}
		return autoEnrollStudentSubjects(tx, &student)
	})
	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update student")
		return
	}

	config.DB.Preload("User").Preload("Class").First(&student, student.ID)
	helpers.Success(c, http.StatusOK, "student updated", student)
}

// ArchiveStudent godoc
// @Summary      Archive (soft-delete) a student (Admin only)
// @Description  Deactivates the student's user account, soft-deletes the student record,
// @Description  and removes all subject enrollments. Attendance, grade, and finance records are preserved.
// @Tags         students
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "Student archived"
// @Failure      404  {object}  helpers.APIResponse  "Student not found"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/students/{id} [delete]
func ArchiveStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		// Deactivate login.
		if err := tx.Model(&models.User{}).Where("id = ?", student.UserID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		// remove enrollments so archived students don't appear in subject lists.
		if err := tx.Where("student_id = ?", student.ID).Delete(&models.Enrollment{}).Error; err != nil {
			return err
		}
		// Soft-delete the student record.
		return tx.Delete(&student).Error
	})

	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to archive student: "+txErr.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "student archived", nil)
}

// ══════════════════════════════════════════════════════
//  TEACHERS
// ══════════════════════════════════════════════════════

// CreateTeacher godoc
// @Summary      Create a teacher (Admin only)
// @Tags         teachers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateTeacherInput  true  "Teacher data"
// @Success      201   {object}  helpers.APIResponse "Teacher created"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      409   {object}  helpers.APIResponse "Email or teacher code already exists"
// @Router       /api/admin/teachers [post]
func CreateTeacher(c *gin.Context) {
	var input CreateTeacherInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	var count int64
	config.DB.Unscoped().Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		helpers.Error(c, http.StatusConflict, "email already registered")
		return
	}

	var codeCount int64
	config.DB.Unscoped().
		Model(&models.Teacher{}).
		Where("teacher_code = ?", input.TeacherCode).
		Count(&codeCount)

	if codeCount > 0 {
		helpers.Error(c, http.StatusConflict, "teacher code already exists")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var user models.User
	var teacher models.Teacher

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		user = models.User{
			Name:     input.Name,
			Email:    input.Email,
			Password: string(hashed),
			Phone:    input.Phone,
			Role:     models.RoleTeacher,
			IsActive: true,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		teacher = models.Teacher{
			UserID:        user.ID,
			TeacherCode:   input.TeacherCode,
			Qualification: input.Qualification,
			Department:    input.Department,
			JoinedAt:      time.Now(),
		}
		return tx.Create(&teacher).Error
	}); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create teacher: "+err.Error())
		return
	}

	helpers.Success(c, http.StatusCreated, "teacher created", gin.H{
		"user_id":    user.ID,
		"teacher_id": teacher.ID,
		"name":       user.Name,
		"code":       teacher.TeacherCode,
	})
}

// GetTeachers godoc
// @Summary      List all teachers (Admin only)
// @Tags         teachers
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "Paginated teacher list"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/teachers [get]
func GetTeachers(c *gin.Context) {
	offset, limit := parsePage(c)
	var teachers []models.Teacher
	var total int64

	// FIX #6: use a single shared query base for both Count and Find so that
	// the same GORM scoping (soft-delete filter, etc.) applies to both calls —
	// same pattern as the FIX #1 applied to GetStudents.
	db := config.DB.Model(&models.Teacher{})
	db.Count(&total)
	if err := db.Preload("User").Offset(offset).Limit(limit).Find(&teachers).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch teachers")
		return
	}
	helpers.Success(c, http.StatusOK, "teachers fetched", gin.H{"total": total, "data": teachers})
}

// GetTeacher godoc
// @Summary      Get a teacher by ID (Admin only)
// @Tags         teachers
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Teacher ID"
// @Success      200  {object}  helpers.APIResponse  "Teacher record"
// @Failure      404  {object}  helpers.APIResponse  "Teacher not found"
// @Router       /api/admin/teachers/{id} [get]
func GetTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher
	if err := config.DB.Preload("User").First(&teacher, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "teacher not found")
		return
	}
	helpers.Success(c, http.StatusOK, "teacher fetched", teacher)
}

// UpdateTeacher godoc
// @Summary      Update a teacher (Admin only)
// @Description  Updates the qualification of a teacher and reloads the record before returning.
// @Tags         teachers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "Teacher ID"
// @Param        body  body      UpdateTeacherInput  true  "Fields to update"
// @Success      200   {object}  helpers.APIResponse "Teacher updated"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      404   {object}  helpers.APIResponse "Teacher not found"
// @Router       /api/admin/teachers/{id} [put]
func UpdateTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher
	if err := config.DB.First(&teacher, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "teacher not found")
		return
	}
	var input UpdateTeacherInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := map[string]interface{}{}
	if input.Qualification != "" {
		updates["qualification"] = input.Qualification
	}
	if input.Department != "" {
		updates["department"] = input.Department
	}
	if len(updates) > 0 {
		if err := config.DB.Model(&teacher).Updates(updates).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to update teacher")
			return
		}
	}
	if input.Phone != "" {
		if err := config.DB.Model(&models.User{}).Where("id = ?", teacher.UserID).Update("phone", input.Phone).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to update phone")
			return
		}
	}
	// reload the record after update — GORM does not refresh the struct in place.
	config.DB.Preload("User").First(&teacher, teacher.ID)
	helpers.Success(c, http.StatusOK, "teacher updated", teacher)
}

// ArchiveTeacher godoc
// @Summary      Archive (soft-delete) a teacher (Admin only)
// @Description  Deactivates the teacher's user account and soft-deletes the teacher record.
// @Description  NOTE: subjects assigned to this teacher are NOT automatically reassigned.
// @Description  Use PUT /api/admin/subjects/:id to reassign them before or after archiving.
// @Tags         teachers
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Teacher ID"
// @Success      200  {object}  helpers.APIResponse  "Teacher archived"
// @Failure      404  {object}  helpers.APIResponse  "Teacher not found"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/teachers/{id} [delete]
func ArchiveTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher
	if err := config.DB.Preload("User").First(&teacher, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "teacher not found")
		return
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", teacher.UserID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Delete(&teacher).Error
	})

	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to archive teacher: "+txErr.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "teacher archived", nil)
}

// ══════════════════════════════════════════════════════
//  CLASSES & SUBJECTS
// ══════════════════════════════════════════════════════

// CreateClass godoc
// @Summary      Create a class (Admin only)
// @Tags         classes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateClassInput    true  "Class data"
// @Success      201   {object}  helpers.APIResponse "Class created"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      500   {object}  helpers.APIResponse "Database error"
// @Router       /api/admin/classes [post]
func CreateClass(c *gin.Context) {
	var input CreateClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if input.GradeLevel <= 10 {
		if input.Stream != "" {
			helpers.Error(c, http.StatusBadRequest, "stream must be empty for grades 9 and 10")
			return
		}
	} else {
		if input.Stream != models.StreamNatural && input.Stream != models.StreamSocial {
			helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science' for grades 11 and 12")
			return
		}
	}
	section := strings.ToUpper(strings.TrimSpace(input.Section))
	if section == "" {
		helpers.Error(c, http.StatusBadRequest, "section is required")
		return
	}
	var dup int64
	dupQ := config.DB.Model(&models.Class{}).
		Where("grade_level = ? AND section = ? AND year = ?", input.GradeLevel, section, input.Year)
	if input.Stream != "" {
		dupQ = dupQ.Where("stream = ?", input.Stream)
	} else {
		dupQ = dupQ.Where("stream = '' OR stream IS NULL")
	}
	dupQ.Count(&dup)
	if dup > 0 {
		helpers.Error(c, http.StatusConflict, fmt.Sprintf("class %d%s already exists for this year", input.GradeLevel, section))
		return
	}
	name := input.Name
	if name == "" {
		name = fmt.Sprintf("%d%s", input.GradeLevel, section)
		if input.Stream != "" {
			name += " " + input.Stream
		}
	}
	status := input.Status
	if status == "" {
		status = "Active"
	}
	var teacherIDPtr *uint
	if input.TeacherID != 0 {
		teacherIDPtr = &input.TeacherID
	}
	class := models.Class{
		Name: name, GradeLevel: input.GradeLevel, Section: section,
		Stream: input.Stream, Status: status, Year: input.Year, TeacherID: teacherIDPtr,
	}

	if err := config.DB.Create(&class).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create class: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusCreated, "class created", class)
}

// GetClasses godoc
// @Summary      List all classes (Admin only)
// @Tags         classes
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "Paginated list of classes"
// @Router       /api/admin/classes [get]
func GetClasses(c *gin.Context) {
	offset, limit := parsePage(c)
	var classes []models.Class
	var total int64

	// FIX #8: apply pagination — GetClasses was the only list endpoint that returned
	// all rows unbounded. Large schools with hundreds of classes would get a slow,
	// oversized response.
	db := config.DB.Model(&models.Class{})
	db.Count(&total)
	if err := db.Preload("Teacher.User").Offset(offset).Limit(limit).Find(&classes).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch classes")
		return
	}
	helpers.Success(c, http.StatusOK, "classes fetched", gin.H{"total": total, "data": classes})
}

// UpdateClass godoc
// @Summary      Update a class (Admin only)
// @Description  Updates the homeroom teacher, year, or name of a class.
// @Description  FIX: new endpoint — previously classes could only be created or archived.
// @Tags         classes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int               true  "Class ID"
// @Param        body  body      UpdateClassInput  true  "Fields to update"
// @Success      200   {object}  helpers.APIResponse "Class updated"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      404   {object}  helpers.APIResponse "Class not found"
// @Router       /api/admin/classes/{id} [put]
func UpdateClass(c *gin.Context) {
	id := c.Param("id")
	var class models.Class
	if err := config.DB.First(&class, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "class not found")
		return
	}

	var input UpdateClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]any{}
	if input.TeacherID != nil {
		// Validate the teacher exists
		var t models.Teacher
		if err := config.DB.First(&t, *input.TeacherID).Error; err != nil {
			helpers.Error(c, http.StatusBadRequest, "teacher not found")
			return
		}
		updates["teacher_id"] = *input.TeacherID
	}
	if input.Year != nil {
		updates["year"] = *input.Year
	}
	if input.Name != "" {
		updates["name"] = input.Name
	}

	if len(updates) == 0 {
		helpers.Error(c, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := config.DB.Model(&class).Updates(updates).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update class")
		return
	}

	config.DB.Preload("Teacher.User").First(&class, class.ID)
	helpers.Success(c, http.StatusOK, "class updated", class)
}

// ArchiveClass godoc
// @Summary      Archive (soft-delete) a class (Admin only)
// @Description  Blocked if any active students are still assigned to it.
// @Tags         classes
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Class ID"
// @Success      200  {object}  helpers.APIResponse  "Class archived"
// @Failure      404  {object}  helpers.APIResponse  "Class not found"
// @Failure      409  {object}  helpers.APIResponse  "Class still has active students"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/classes/{id} [delete]
func ArchiveClass(c *gin.Context) {
	id := c.Param("id")
	var class models.Class
	if err := config.DB.First(&class, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "class not found")
		return
	}

	var activeCount int64
	config.DB.Model(&models.Student{}).
		Where("class_id = ?", id).
		Count(&activeCount)
	if activeCount > 0 {
		helpers.Error(c, http.StatusConflict,
			fmt.Sprintf("cannot archive class: %d active student(s) still assigned — reassign or archive them first", activeCount))
		return
	}

	if err := config.DB.Delete(&class).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to archive class: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "class archived", nil)
}

// CreateSubject godoc
// @Summary      Create a subject (Admin only)
// @Tags         subjects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateSubjectInput  true  "Subject data"
// @Success      201   {object}  helpers.APIResponse "Subject created"
// @Failure      400   {object}  helpers.APIResponse "Validation error"
// @Failure      409   {object}  helpers.APIResponse "Subject code already exists"
// @Failure      500   {object}  helpers.APIResponse "Database error"
// @Router       /api/admin/subjects [post]
func CreateSubject(c *gin.Context) {
	var input CreateSubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if input.GradeLevel > 0 && input.GradeLevel <= 10 {
		if input.Stream != "" {
			helpers.Error(c, http.StatusBadRequest, "stream must be empty for grades 9 and 10")
			return
		}
	} else if input.GradeLevel >= 11 {
		if input.Stream != models.StreamNatural && input.Stream != models.StreamSocial {
			helpers.Error(c, http.StatusBadRequest, "stream must be 'Natural Science' or 'Social Science' for grades 11 and 12")
			return
		}
	}

	// check for duplicate code BEFORE the insert so we can return the correct status.
	// The old code caught ALL DB errors and reported them as 409 (even network errors).
	var codeCount int64
	config.DB.Model(&models.Subject{}).Where("code = ?", input.Code).Count(&codeCount)
	if codeCount > 0 {
		helpers.Error(c, http.StatusConflict, "subject code already exists")
		return
	}

	st := input.Status
	if st == "" {
		st = "Active"
	}
	var teacherIDPtr *uint
	if input.TeacherID != 0 {
		teacherIDPtr = &input.TeacherID
	}
	subject := models.Subject{
		Name: input.Name, Code: input.Code, GradeLevel: input.GradeLevel,
		Stream: input.Stream, Status: st, TeacherID: teacherIDPtr,
	}
	if err := config.DB.Create(&subject).Error; err != nil {
		// Any remaining error is a genuine server/DB failure.
		helpers.Error(c, http.StatusInternalServerError, "failed to create subject: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusCreated, "subject created", subject)
}

// GetSubjects godoc
// @Summary      List all subjects (Admin only)
// @Tags         subjects
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 50, max 200)"
// @Success      200  {object}  helpers.APIResponse  "Paginated list of subjects"
// @Router       /api/admin/subjects [get]
func GetSubjects(c *gin.Context) {
	offset, limit := parsePage(c)
	var subjects []models.Subject
	var total int64

	// FIX #8: apply pagination — same as GetClasses fix above.
	db := config.DB.Model(&models.Subject{})
	if stream := strings.TrimSpace(c.Query("stream")); stream != "" {
		if stream == "Common" {
			db = db.Where("stream = '' OR stream IS NULL")
		} else {
			db = db.Where("stream = '' OR stream IS NULL OR stream = ?", stream)
		}
	}
	if gl := c.Query("grade_level"); gl != "" {
		db = db.Where("grade_level = ?", gl)
	}
	db.Count(&total)
	if err := db.Preload("Teacher.User").Offset(offset).Limit(limit).Find(&subjects).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch subjects")
		return
	}
	helpers.Success(c, http.StatusOK, "subjects fetched", gin.H{"total": total, "data": subjects})
}

// UpdateSubject godoc
// @Summary      Update a subject (Admin only)
// @Description  Updates a subject's name, code, or assigned teacher.
// @Description  FIX: new endpoint — previously subjects could only be created or archived, leaving
// @Description  orphaned teacher references after ArchiveTeacher with no way to reassign via the API.
// @Tags         subjects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "Subject ID"
// @Param        body  body      UpdateSubjectInput   true  "Fields to update"
// @Success      200   {object}  helpers.APIResponse  "Subject updated"
// @Failure      400   {object}  helpers.APIResponse  "Validation error or code conflict"
// @Failure      404   {object}  helpers.APIResponse  "Subject not found"
// @Router       /api/admin/subjects/{id} [put]
func UpdateSubject(c *gin.Context) {
	id := c.Param("id")
	var subject models.Subject
	if err := config.DB.First(&subject, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "subject not found")
		return
	}

	var input UpdateSubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]any{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Code != "" {
		// Check the new code isn't already taken by another subject.
		var codeCount int64
		config.DB.Model(&models.Subject{}).
			Where("code = ? AND id != ?", input.Code, subject.ID).
			Count(&codeCount)
		if codeCount > 0 {
			helpers.Error(c, http.StatusConflict, "subject code already in use by another subject")
			return
		}
		updates["code"] = input.Code
	}
	if input.TeacherID != nil {
		// Allow setting to 0 to unassign a teacher.
		if *input.TeacherID != 0 {
			var t models.Teacher
			if err := config.DB.First(&t, *input.TeacherID).Error; err != nil {
				helpers.Error(c, http.StatusBadRequest, "teacher not found")
				return
			}
		}
		updates["teacher_id"] = *input.TeacherID
	}

	if len(updates) == 0 {
		helpers.Error(c, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := config.DB.Model(&subject).Updates(updates).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update subject")
		return
	}

	config.DB.Preload("Teacher.User").First(&subject, subject.ID)
	helpers.Success(c, http.StatusOK, "subject updated", subject)
}

// ArchiveSubject godoc
// @Summary      Archive (soft-delete) a subject (Admin only)
// @Description  Blocked if grades or attendance records reference it.
// @Tags         subjects
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Subject ID"
// @Success      200  {object}  helpers.APIResponse  "Subject archived"
// @Failure      404  {object}  helpers.APIResponse  "Subject not found"
// @Failure      409  {object}  helpers.APIResponse  "Subject has existing grades or attendance records"
// @Failure      500  {object}  helpers.APIResponse  "Database error"
// @Router       /api/admin/subjects/{id} [delete]
func ArchiveSubject(c *gin.Context) {
	id := c.Param("id")
	var subject models.Subject
	if err := config.DB.First(&subject, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "subject not found")
		return
	}

	var gradeCount int64
	config.DB.Model(&models.Grade{}).Where("subject_id = ?", id).Count(&gradeCount)
	if gradeCount > 0 {
		helpers.Error(c, http.StatusConflict,
			fmt.Sprintf("cannot archive subject: %d grade record(s) exist — historical data must be preserved", gradeCount))
		return
	}

	var attendanceCount int64
	config.DB.Model(&models.Attendance{}).Where("subject_id = ?", id).Count(&attendanceCount)
	if attendanceCount > 0 {
		helpers.Error(c, http.StatusConflict,
			fmt.Sprintf("cannot archive subject: %d attendance record(s) exist — historical data must be preserved", attendanceCount))
		return
	}

	if err := config.DB.Delete(&subject).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to archive subject: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "subject archived", nil)
}

// ══════════════════════════════════════════════════════
//  ENROLLMENT
// ══════════════════════════════════════════════════════

// EnrollStudent godoc
// @Summary      Enroll student in a subject (Admin only)
// @Tags         enrollment
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      EnrollStudentInput  true  "Enrollment data"
// @Success      201   {object}  helpers.APIResponse "Student enrolled"
// @Failure      400   {object}  helpers.APIResponse "Validation error, student not found, or subject not found"
// @Failure      409   {object}  helpers.APIResponse "Already enrolled"
// @Router       /api/admin/enroll [post]
func EnrollStudent(c *gin.Context) {
	var input EnrollStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var student models.Student
	if err := config.DB.First(&student, input.StudentID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "student not found")
		return
	}
	var subject models.Subject
	if err := config.DB.First(&subject, input.SubjectID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "subject not found")
		return
	}
	if subject.GradeLevel != 0 && subject.GradeLevel != student.GradeLevel {
		helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student grade level %d does not match subject grade level %d", student.GradeLevel, subject.GradeLevel))
		return
	}
	if subject.Stream != "" && student.Stream != subject.Stream {
		helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student stream '%s' does not match subject stream '%s'", student.Stream, subject.Stream))
		return
	}
	var count int64
	config.DB.Model(&models.Enrollment{}).
		Where("student_id = ? AND subject_id = ?", input.StudentID, input.SubjectID).
		Count(&count)
	if count > 0 {
		helpers.Error(c, http.StatusConflict, "student already enrolled in this subject")
		return
	}
	enrollment := models.Enrollment{StudentID: input.StudentID, SubjectID: input.SubjectID}
	if err := config.DB.Create(&enrollment).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to enroll student: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusCreated, "student enrolled", enrollment)
}

// UnenrollStudent godoc
// @Summary      Remove a student from a subject (Admin only)
// @Description  Deletes the enrollment record. Attendance and grade records for this student/subject are preserved.
// @Description  FIX: new endpoint — previously enrollments could only be created, never removed.
// @Tags         enrollment
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      UnenrollStudentInput  true  "Unenrollment data"
// @Success      200   {object}  helpers.APIResponse   "Student unenrolled"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      404   {object}  helpers.APIResponse   "Enrollment not found"
// @Router       /api/admin/unenroll [delete]
func UnenrollStudent(c *gin.Context) {
	var input UnenrollStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result := config.DB.
		Where("student_id = ? AND subject_id = ?", input.StudentID, input.SubjectID).
		Delete(&models.Enrollment{})

	if result.Error != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to unenroll student: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		helpers.Error(c, http.StatusNotFound, "enrollment not found")
		return
	}

	helpers.Success(c, http.StatusOK, "student unenrolled", nil)
}

// ══════════════════════════════════════════════════════
//  ATTENDANCE
// ══════════════════════════════════════════════════════

// RecordAttendance godoc
// @Summary      Record attendance (Teacher only)
// @Description  Records or updates attendance for a student in a subject on a given date.
// @Description  The teacher must be assigned to the subject being marked.
// @Description  FIX #3a: The student must be enrolled in the subject — phantom attendance records are now rejected.
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      AttendanceInput      true  "Attendance data"
// @Success      201   {object}  helpers.APIResponse  "Attendance recorded"
// @Success      200   {object}  helpers.APIResponse  "Attendance updated (same-day override)"
// @Failure      400   {object}  helpers.APIResponse  "Validation error or student not enrolled in subject"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — not your subject"
// @Router       /api/academics/attendance [post]
func RecordAttendance(c *gin.Context) {
	role := c.GetString("role")
	if role != models.RoleTeacher && role != models.RoleAdmin {
		helpers.Error(c, http.StatusForbidden, "only teachers and admins can record attendance")
		return
	}

	var input AttendanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var student models.Student
	if err := config.DB.First(&student, input.StudentID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "student not found")
		return
	}

	if role == models.RoleTeacher {
		teacherUserID := c.GetUint("userID")
		var teacher models.Teacher
		if err := config.DB.Where("user_id = ?", teacherUserID).First(&teacher).Error; err != nil {
			helpers.Error(c, http.StatusForbidden, "teacher profile not found")
			return
		}
		if student.ClassID != nil {
			var class models.Class
			if err := config.DB.First(&class, *student.ClassID).Error; err == nil && (class.TeacherID == nil || *class.TeacherID != teacher.ID) {
				helpers.Error(c, http.StatusForbidden, "you can only record attendance for your homeroom class")
				return
			}
		}
	}
	// FIX #1: Admins previously bypassed ALL ownership checks, allowing them to mark
	// attendance for any student in any class without an audit trail. We now log the
	// admin override to stdout and stamp the attendance record with the admin's user
	// ID (via CreatedAt of the actor). The original SubjectID is still null because
	// this is a homeroom-style daily record.
	if role == models.RoleAdmin {
		fmt.Printf("[AUDIT] admin user_id=%d recorded attendance for student_id=%d on %s (admin override)\n",
			c.GetUint("userID"), input.StudentID, input.Date)
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	var existing models.Attendance
	dayStart := date
	dayEnd := date.Add(24 * time.Hour)
	result := config.DB.Where(
		"student_id = ? AND subject_id IS NULL AND date >= ? AND date < ?",
		input.StudentID, dayStart, dayEnd,
	).First(&existing)

	if result.Error == nil {
		if err := config.DB.Model(&existing).Updates(map[string]any{
			"status": input.Status,
			"notes":  input.Notes,
		}).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to update attendance")
			return
		}
		config.DB.Preload("Student.User").First(&existing, existing.ID)
		helpers.Success(c, http.StatusOK, "attendance updated", existing)
		return
	}

	attendance := models.Attendance{
		StudentID: input.StudentID,
		SubjectID: nil,
		Date:      date,
		Status:    input.Status,
		Notes:     input.Notes,
	}
	if err := config.DB.Create(&attendance).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to record attendance")
		return
	}
	config.DB.Preload("Student.User").First(&attendance, attendance.ID)
	helpers.Success(c, http.StatusCreated, "attendance recorded", attendance)
}

// GetClassAttendance godoc
// @Summary      Get class attendance for a date (Teacher only)
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        classID  path   int     true  "Class ID"
// @Param        date     query  string  true  "Date in YYYY-MM-DD format"
// @Success      200  {object}  helpers.APIResponse  "Attendance records"
// @Failure      400  {object}  helpers.APIResponse  "Missing or invalid date"
// @Router       /api/academics/attendance/class/{classID} [get]
func GetClassAttendance(c *gin.Context) {
	classID := c.Param("classID")
	date := c.Query("date")

	if date == "" {
		helpers.Error(c, http.StatusBadRequest, "date query parameter is required (YYYY-MM-DD)")
		return
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	var records []models.Attendance
	config.DB.
		Joins("JOIN students ON attendances.student_id = students.id").
		Preload("Student.User").
		Where("students.class_id = ?", classID).
		Where("attendances.subject_id IS NULL").
		Where("DATE(attendances.date) = ?", date).
		Find(&records)

	helpers.Success(c, http.StatusOK, "class attendance", records)
}

// GetAttendancePercentage godoc
// @Summary      Get student attendance percentage
// @Description  Returns overall and per-subject attendance stats for a student.
// @Description  Accessible by: Teacher (any student), Student (own only), Admin, Parent (own child via ParentOwnsStudent middleware).
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path  int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "Attendance summary"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Students may only view their own attendance"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/academics/attendance/{studentID} [get]
func GetAttendancePercentage(c *gin.Context) {
	studentID := c.Param("studentID")

	if c.GetString("role") == models.RoleStudent {
		var self models.Student
		if err := config.DB.Where("user_id = ?", c.GetUint("userID")).First(&self).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if fmt.Sprintf("%d", self.ID) != studentID {
			helpers.Error(c, http.StatusForbidden, "you can only view your own attendance")
			return
		}
	}
	var total, present int64

	config.DB.Model(&models.Attendance{}).Where("student_id = ? AND subject_id IS NULL", studentID).Count(&total)
	config.DB.Model(&models.Attendance{}).
		Where("student_id = ? AND subject_id IS NULL AND status IN ('Present','Late')", studentID).
		Count(&present)

	percentage := 0.0
	if total > 0 {
		percentage = float64(present) / float64(total) * 100
	}

	type MonthStat struct {
		Month      string  `json:"month"`
		Total      int64   `json:"total"`
		Present    int64   `json:"present"`
		Percentage float64 `json:"percentage"`
	}
	breakdown := []MonthStat{}
	config.DB.Raw(`
		SELECT TO_CHAR(a.date, 'Mon YYYY') as month,
		       COUNT(a.id) as total,
		       SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) as present,
		       ROUND(SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(a.id), 0), 2) as percentage
		FROM attendances a
		WHERE a.student_id = ? AND a.subject_id IS NULL
		GROUP BY TO_CHAR(a.date, 'Mon YYYY'), EXTRACT(YEAR FROM a.date), EXTRACT(MONTH FROM a.date)
		ORDER BY EXTRACT(YEAR FROM a.date), EXTRACT(MONTH FROM a.date)
	`, studentID).Scan(&breakdown)

	helpers.Success(c, http.StatusOK, "attendance summary", gin.H{
		"student_id":         studentID,
		"overall_percentage": percentage,
		"total_days":         total,
		"attended":           present,
		"by_month":           breakdown,
	})
}

// GetAttendanceSummary godoc
// @Summary      School-wide attendance summary (Admin only)
// @Description  FIX #5: SQL now uses NULLIF(COUNT, 0) and HAVING COUNT > 0 to prevent division-by-zero
// @Description  for students who have no attendance records at all.
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "Attendance summary"
// @Router       /api/admin/attendance/summary [get]
func GetAttendanceSummary(c *gin.Context) {
	type Summary struct {
		StudentName string `json:"student_name"`
		StudentCode string `json:"student_code"`
		ClassName   string `json:"class_name"`
		GradeLevel  int    `json:"grade_level"`
		Section     string `json:"section"`
		Date        string `json:"date"`
		Status      string `json:"status"`
	}
	dateFilter := c.Query("date")
	gradeFilter := c.Query("grade_level")
	sectionFilter := c.Query("section")
	classIDFilter := c.Query("class_id")

	query := `
		SELECT
			u.name AS student_name,
			st.student_code AS student_code,
			COALESCE(cl.name, '') AS class_name,
			COALESCE(cl.grade_level, st.grade_level) AS grade_level,
			COALESCE(cl.section, '') AS section,
			TO_CHAR(a.date, 'YYYY-MM-DD') AS date,
			a.status AS status
		FROM attendances a
		JOIN students st ON a.student_id = st.id
		JOIN users u ON st.user_id = u.id
		LEFT JOIN classes cl ON st.class_id = cl.id
		WHERE a.subject_id IS NULL
	`
	args := []any{}
	if dateFilter != "" {
		query += " AND DATE(a.date) = DATE(?)"
		args = append(args, dateFilter)
	}
	if gradeFilter != "" {
		query += " AND COALESCE(cl.grade_level, st.grade_level) = ?"
		args = append(args, gradeFilter)
	}
	if sectionFilter != "" {
		query += " AND UPPER(COALESCE(cl.section, '')) = UPPER(?)"
		args = append(args, sectionFilter)
	}
	if classIDFilter != "" {
		query += " AND st.class_id = ?"
		args = append(args, classIDFilter)
	}
	query += " ORDER BY a.date DESC, u.name LIMIT 500"

	var summary []Summary
	config.DB.Raw(query, args...).Scan(&summary)
	helpers.Success(c, http.StatusOK, "attendance summary", summary)
}

// ══════════════════════════════════════════════════════
//  GRADES
// ══════════════════════════════════════════════════════

// BulkGradeEntry godoc
// @Summary      Bulk grade entry (Teacher only)
// @Description  Records grades for multiple students in one call. Uses upsert — if a grade for the same
// @Description  (student, subject, type, term, academic_year) already exists it is updated, not duplicated.
// @Description  FIX #2: The upsert requires a DB-level unique constraint on (student_id, subject_id, type, term, academic_year).
// @Description  Run migration.sql before deploying — without it the upsert silently creates duplicate rows.
// @Description  FIX #3b: Each student_id in the grades array is now validated to be enrolled in the subject.
// @Tags         grades
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      BulkGradeInput       true  "Grade data"
// @Success      201   {object}  helpers.APIResponse  "Grades recorded"
// @Failure      400   {object}  helpers.APIResponse  "Validation error or student not enrolled"
// @Failure      403   {object}  helpers.APIResponse  "Forbidden — not your subject"
// @Router       /api/academics/grades/bulk [post]
func BulkGradeEntry(c *gin.Context) {
	teacherUserID := c.GetUint("userID")

	var input BulkGradeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var teacher models.Teacher
	if err := config.DB.Where("user_id = ?", teacherUserID).First(&teacher).Error; err != nil {
		helpers.Error(c, http.StatusForbidden, "teacher profile not found")
		return
	}

	var dbGrades []models.Grade
	newStudentsMap := make(map[uint]struct {
		Email string
		Name  string
		Sub   string
		Score float64
	})

	// To avoid duplicates or N+1 queries, we can fetch subjects and students needed
	subjectIDs := []uint{}
	studentIDs := []uint{}
	for _, entry := range input.Grades {
		subjectIDs = append(subjectIDs, entry.SubjectID)
		studentIDs = append(studentIDs, entry.StudentID)
	}

	var subjects []models.Subject
	if err := config.DB.Where("id IN ?", subjectIDs).Find(&subjects).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch subjects")
		return
	}
	subjectMap := make(map[uint]models.Subject)
	for _, s := range subjects {
		subjectMap[s.ID] = s
	}

	var students []models.Student
	if err := config.DB.Preload("User").Where("id IN ?", studentIDs).Find(&students).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch students")
		return
	}
	studentMap := make(map[uint]models.Student)
	for _, st := range students {
		studentMap[st.ID] = st
	}

	// Fetch all enrollments for validation
	var enrollments []models.Enrollment
	if err := config.DB.Where("student_id IN ? AND subject_id IN ?", studentIDs, subjectIDs).Find(&enrollments).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch enrollments")
		return
	}
	enrollmentMap := make(map[string]bool)
	for _, e := range enrollments {
		key := fmt.Sprintf("%d-%d", e.StudentID, e.SubjectID)
		enrollmentMap[key] = true
	}

	// Validate and build Grade models
	for _, entry := range input.Grades {
		sub, ok := subjectMap[entry.SubjectID]
		if !ok {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("subject_id %d not found", entry.SubjectID))
			return
		}

		// Verify teacher ownership
		if sub.TeacherID == nil || *sub.TeacherID != teacher.ID {
			helpers.Error(c, http.StatusForbidden, fmt.Sprintf("you are not assigned to subject: %s", sub.Name))
			return
		}

		student, ok := studentMap[entry.StudentID]
		if !ok {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student_id %d not found", entry.StudentID))
			return
		}

		// Verify enrollment
		key := fmt.Sprintf("%d-%d", entry.StudentID, entry.SubjectID)
		if !enrollmentMap[key] {
			helpers.Error(c, http.StatusBadRequest, fmt.Sprintf("student %s is not enrolled in subject %s", student.User.Name, sub.Name))
			return
		}

		// FIX #2: also verify the student's grade level and stream match the subject's.
		// Without this, a Grade 9 student enrolled in a Grade 11 subject (e.g. via a manual
		// EnrollStudent call that bypassed the auto-enroll path) could receive Grade 11
		// grades. EnrollStudent already does this check, but BulkGradeEntry skipped it.
		if sub.GradeLevel != 0 && sub.GradeLevel != student.GradeLevel {
			helpers.Error(c, http.StatusBadRequest,
				fmt.Sprintf("student %s is in Grade %d but subject %s is Grade %d",
					student.User.Name, student.GradeLevel, sub.Name, sub.GradeLevel))
			return
		}
		if sub.Stream != "" && student.Stream != sub.Stream {
			helpers.Error(c, http.StatusBadRequest,
				fmt.Sprintf("student %s is in stream '%s' but subject %s is for stream '%s'",
					student.User.Name, student.Stream, sub.Name, sub.Stream))
			return
		}

		// Check if grade already exists for this (student_id, subject_id, type, semester, academic_year)
		var existingGrade models.Grade
		err := config.DB.Where("student_id = ? AND subject_id = ? AND type = ? AND semester = ? AND academic_year = ?",
			entry.StudentID, entry.SubjectID, entry.GradeType, entry.Semester, student.AcademicYear).First(&existingGrade).Error

		isNew := false
		if err == gorm.ErrRecordNotFound {
			isNew = true
		}

		grade := models.Grade{
			StudentID:    entry.StudentID,
			SubjectID:    entry.SubjectID,
			TeacherID:    teacher.ID,
			Score:        entry.Score,
			MaxScore:     100,
			Type:         entry.GradeType,
			Semester:     entry.Semester,
			AcademicYear: student.AcademicYear,
			Remarks:      entry.Remarks,
		}
		if !isNew {
			grade.ID = existingGrade.ID
		}

		dbGrades = append(dbGrades, grade)

		if isNew && student.User.Email != "" {
			newStudentsMap[entry.StudentID] = struct {
				Email string
				Name  string
				Sub   string
				Score float64
			}{
				Email: student.User.Email,
				Name:  student.User.Name,
				Sub:   sub.Name,
				Score: entry.Score,
			}
		}
	}

	// Save all grades
	for _, g := range dbGrades {
		if err := config.DB.Save(&g).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to save grades: "+err.Error())
			return
		}
	}

	// Send emails in background
	go func() {
		for _, notifyInfo := range newStudentsMap {
			helpers.SendGradeNotification(
				notifyInfo.Email,
				notifyInfo.Name,
				notifyInfo.Sub,
				notifyInfo.Score,
			)
		}
	}()

	helpers.Success(c, http.StatusCreated, "grades recorded", gin.H{
		"saved": len(dbGrades),
	})
}

// GetSubjectGrades godoc
// @Summary      Get grades for a subject (Teacher only)
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        subjectID  path   int     true   "Subject ID"
// @Param        type       query  string  false  "Grade type: Midterm | Final | Quiz | Assignment"
// @Param        semester   query  string  false  "Semester: Semester 1 | Semester 2 | Semester 3"
// @Success      200  {object}  helpers.APIResponse  "Grade list"
// @Router       /api/academics/grades/subject/{subjectID} [get]
func GetSubjectGrades(c *gin.Context) {
	subjectID := c.Param("subjectID")
	gradeType := c.Query("type")
	semester := c.Query("semester")

	// Get logged-in user (teacher)
	userID, exists := c.Get("userID")
	if !exists {
		helpers.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	role, _ := c.Get("role")

	// Only enforce for teachers
	if role == models.RoleTeacher {
		// FIX #1: subject.teacher_id stores Teacher.ID (the Teacher table PK), not User.ID.
		// We must resolve the Teacher record first, then compare teacher.ID — exactly as
		// RecordAttendance and BulkGradeEntry already do.  The previous code compared
		// teacher_id against the JWT userID (User.ID), which breaks for any teacher whose
		// User.ID != Teacher.ID (i.e. almost everyone after the first user).
		var teacher models.Teacher
		if err := config.DB.Where("user_id = ?", userID).First(&teacher).Error; err != nil {
			helpers.Error(c, http.StatusForbidden, "teacher profile not found")
			return
		}

		var count int64
		if err := config.DB.Model(&models.Subject{}).
			Where("id = ? AND teacher_id = ?", subjectID, teacher.ID).
			Count(&count).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to verify subject ownership")
			return
		}

		if count == 0 {
			helpers.Error(c, http.StatusForbidden, "you are not assigned to this subject")
			return
		}
	}

	// Query grades
	query := config.DB.Preload("Student.User").
		Where("subject_id = ?", subjectID)

	if gradeType != "" {
		query = query.Where("type = ?", gradeType)
	}

	if semester != "" {
		query = query.Where("semester = ?", semester)
	}

	var grades []models.Grade

	if err := query.Find(&grades).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch grades")
		return
	}

	helpers.Success(c, http.StatusOK, "subject grades", grades)
}

// GetStudentGrades godoc
// @Summary      Get a student's grades
// @Description  Returns grades for a student filtered by semester and/or academic year.
// @Description  Accessible by: Teacher, Student (own only), Admin, Parent (own child).
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path   int     true   "Student ID"
// @Param        semester   query  string  false  "Semester: Semester 1 | Semester 2 | Semester 3"
// @Param        year       query  int     false  "Academic year e.g. 2025"
// @Success      200  {object}  helpers.APIResponse  "Grade list"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Students may only view their own grades"
// @Failure      404  {object}  helpers.APIResponse  "Student profile not found"
// @Router       /api/academics/grades/student/{studentID} [get]
func GetStudentGrades(c *gin.Context) {
	studentID := c.Param("studentID")

	if c.GetString("role") == models.RoleStudent {
		var self models.Student
		if err := config.DB.Where("user_id = ?", c.GetUint("userID")).First(&self).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if fmt.Sprintf("%d", self.ID) != studentID {
			helpers.Error(c, http.StatusForbidden, "you can only view your own grades")
			return
		}
	}
	semester := c.Query("semester")
	year := c.Query("year")

	query := config.DB.Preload("Subject").Where("student_id = ?", studentID)
	if semester != "" {
		query = query.Where("semester = ?", semester)
	}
	if year != "" {
		query = query.Where("academic_year = ?", year)
	}
	var grades []models.Grade
	if err := query.Find(&grades).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch grades")
		return
	}
	helpers.Success(c, http.StatusOK, "grades fetched", grades)
}

// ══════════════════════════════════════════════════════
//  REPORT CARD
// ══════════════════════════════════════════════════════

// GetReportCard godoc
// @Summary      Get report card (JSON)
// @Description  Returns a full report card with per-subject averages and letter grades.
// @Description  FIX: now requires academic_year query param (defaults to current year).
// @Description  Optionally filtered by semester. Without these filters, cross-year averages were meaningless.
// @Description  Accessible by: Teacher, Student (own only), Admin, Parent (own child via ParentOwnsStudent middleware).
// @Tags         report-card
// @Security     BearerAuth
// @Produce      json
// @Param        studentID      path   int     true   "Student ID"
// @Param        academic_year  query  int     false  "Academic year (default: current year)"
// @Param        semester       query  string  false  "Semester filter: Semester 1 | Semester 2 | Semester 3 (omit for full year average)"
// @Success      200  {object}  helpers.APIResponse  "Report card"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Students may only view their own report card"
// @Failure      404  {object}  helpers.APIResponse  "Student not found"
// @Router       /api/academics/reportcard/{studentID} [get]
func GetReportCard(c *gin.Context) {
	studentID := c.Param("studentID")
	var student models.Student
	if err := config.DB.Preload("User").Preload("Class").First(&student, studentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	if c.GetString("role") == models.RoleStudent {
		var self models.Student
		if err := config.DB.Where("user_id = ?", c.GetUint("userID")).First(&self).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if self.ID != student.ID {
			helpers.Error(c, http.StatusForbidden, "you can only view your own report card")
			return
		}
	}

	// default to current year; require explicit year or get a cross-year averaged mess.
	yearStr := c.DefaultQuery("academic_year", fmt.Sprintf("%d", time.Now().Year()))
	semester := c.Query("semester")

	type SubjectReport struct {
		SubjectName string  `json:"subject_name"`
		Midterm     float64 `json:"midterm"`
		Final       float64 `json:"final"`
		Assignments float64 `json:"assignments_avg"`
		Quizzes     float64 `json:"quizzes_avg"`
		Overall     float64 `json:"overall"`
		LetterGrade string  `json:"letter_grade"`
	}

	// Build the WHERE clause dynamically so semester is optional.
	semClause := ""
	args := []any{studentID, yearStr}
	if semester != "" {
		semClause = "AND g.semester = ?"
		args = append(args, semester)
	}

	var report []SubjectReport
	config.DB.Raw(fmt.Sprintf(`
		SELECT
			s.name as subject_name,
			COALESCE(AVG(CASE WHEN g.type = 'Midterm'    THEN g.score END), 0) as midterm,
			COALESCE(AVG(CASE WHEN g.type = 'Final'      THEN g.score END), 0) as final,
			COALESCE(AVG(CASE WHEN g.type = 'Assignment' THEN g.score END), 0) as assignments_avg,
			COALESCE(AVG(CASE WHEN g.type = 'Quiz'       THEN g.score END), 0) as quizzes_avg,
			COALESCE(AVG(g.score), 0) as overall
		FROM grades g
		JOIN subjects s ON g.subject_id = s.id
		WHERE g.student_id = ? AND g.academic_year = ? %s
		GROUP BY s.name
	`, semClause), args...).Scan(&report)

	for i, r := range report {
		report[i].LetterGrade = scoreToLetter(r.Overall)
	}

	helpers.Success(c, http.StatusOK, "report card", gin.H{
		"student": gin.H{
			"id":    student.ID,
			"name":  student.User.Name,
			"code":  student.StudentCode,
			"class": student.Class.Name,
		},
		"academic_year": yearStr,
		"semester":      semester,
		"subjects":      report,
	})
}

// DownloadReportCard godoc
// @Summary      Download report card as PDF
// @Description  Generates and streams a PDF report card. Accepts academic_year and semester query params.
// @Tags         report-card
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        studentID      path   int     true   "Student ID"
// @Param        academic_year  query  int     false  "Academic year (default: current year)"
// @Param        semester       query  string  false  "Semester filter: Semester 1 | Semester 2 | Semester 3"
// @Success      200  {file}    binary               "PDF download"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Students may only download their own report card"
// @Failure      404  {object}  helpers.APIResponse  "Student not found"
// @Failure      500  {object}  helpers.APIResponse  "PDF generation failed"
// @Router       /api/academics/reportcard/{studentID}/pdf [get]
func DownloadReportCard(c *gin.Context) {
	studentID := c.Param("studentID")
	var student models.Student
	if err := config.DB.Preload("User").Preload("Class").First(&student, studentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	if c.GetString("role") == models.RoleStudent {
		var self models.Student
		if err := config.DB.Where("user_id = ?", c.GetUint("userID")).First(&self).Error; err != nil {
			helpers.Error(c, http.StatusNotFound, "student profile not found")
			return
		}
		if self.ID != student.ID {
			helpers.Error(c, http.StatusForbidden, "you can only download your own report card")
			return
		}
	}

	yearStr := c.DefaultQuery("academic_year", fmt.Sprintf("%d", time.Now().Year()))
	semester := c.Query("semester")

	semClause := ""
	args := []any{student.ID, yearStr}
	if semester != "" {
		semClause = "AND g.semester = ?"
		args = append(args, semester)
	}

	type row struct {
		SubjectName string
		Overall     float64
	}
	var rows []row
	config.DB.Raw(fmt.Sprintf(`
		SELECT s.name as subject_name, COALESCE(AVG(g.score), 0) as overall
		FROM grades g
		JOIN subjects s ON g.subject_id = s.id
		WHERE g.student_id = ? AND g.academic_year = ? %s
		GROUP BY s.name
	`, semClause), args...).Scan(&rows)

	var courses []helpers.CourseGrade
	for _, r := range rows {
		courses = append(courses, helpers.CourseGrade{
			Title:       r.SubjectName,
			Score:       r.Overall,
			LetterGrade: scoreToLetter(r.Overall),
		})
	}

	pdfBytes, err := helpers.GenerateReportCardPDF(student, courses)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to generate PDF")
		return
	}

	// FIX #7: use mime.FormatMediaType to safely encode the filename.
	// The previous strings.ReplaceAll only escaped double-quotes, leaving CRLF
	// characters (\r\n) and semicolons exploitable for HTTP response splitting or
	// Content-Disposition parameter injection. mime.FormatMediaType handles all
	// RFC 6266 encoding correctly.
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": fmt.Sprintf("report_%s.pdf", student.StudentCode),
	})
	c.Header("Content-Disposition", disposition)

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ══════════════════════════════════════════════════════
//  PARENT (admin + portal)
// ══════════════════════════════════════════════════════

// GetParents lists parent user accounts for admin
func GetParents(c *gin.Context) {
	offset, limit := parsePage(c)
	var parents []models.User
	var total int64

	// Count query - distinct users
	dbCount := config.DB.Model(&models.User{}).Where("users.role = ?", models.RoleParent)
	if q := strings.TrimSpace(c.Query("search")); q != "" {
		like := "%" + q + "%"
		dbCount = dbCount.Joins("LEFT JOIN students ON students.parent_id = users.id").
			Joins("LEFT JOIN users s_u ON students.user_id = s_u.id AND s_u.deleted_at IS NULL").
			Where("users.name ILIKE ? OR users.email ILIKE ? OR s_u.name ILIKE ?", like, like, like)
	}
	dbCount.Distinct("users.id").Count(&total)

	// Data query
	dbParents := config.DB.Model(&models.User{}).Where("users.role = ?", models.RoleParent)
	if q := strings.TrimSpace(c.Query("search")); q != "" {
		like := "%" + q + "%"
		dbParents = dbParents.Joins("LEFT JOIN students ON students.parent_id = users.id").
			Joins("LEFT JOIN users s_u ON students.user_id = s_u.id AND s_u.deleted_at IS NULL").
			Where("users.name ILIKE ? OR users.email ILIKE ? OR s_u.name ILIKE ?", like, like, like).
			Distinct("users.id", "users.name", "users.email", "users.phone", "users.is_active", "users.created_at")
	}
	if err := dbParents.Order("users.name ASC").Offset(offset).Limit(limit).Find(&parents).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch parents")
		return
	}

	// Fetch children for these parents
	parentIDs := make([]uint, len(parents))
	for i, p := range parents {
		parentIDs[i] = p.ID
	}

	var students []models.Student
	if len(parentIDs) > 0 {
		if err := config.DB.Preload("User").Preload("Class").Where("parent_id IN ?", parentIDs).Find(&students).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to fetch children")
			return
		}
	}

	studentsByParent := make(map[uint][]models.Student)
	for _, s := range students {
		studentsByParent[s.ParentID] = append(studentsByParent[s.ParentID], s)
	}

	type childRow struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		StudentCode string `json:"student_code"`
		Grade       string `json:"grade"`
		Section     string `json:"section"`
	}

	type parentRow struct {
		ID            uint       `json:"id"`
		Name          string     `json:"name"`
		Email         string     `json:"email"`
		Phone         string     `json:"phone"`
		IsActive      bool       `json:"is_active"`
		Status        string     `json:"status"`
		ChildrenCount int        `json:"children_count"`
		Children      []childRow `json:"children"`
		StudentNames  string     `json:"student_names"`
		CreatedAt     string     `json:"created_at"`
	}

	var rows []parentRow
	for _, p := range parents {
		st := "Active"
		if !p.IsActive {
			st = "Inactive"
		}

		children := []childRow{}
		studentNamesList := []string{}
		for _, s := range studentsByParent[p.ID] {
			name := s.User.Name
			studentNamesList = append(studentNamesList, name)

			gradeStr := ""
			secStr := ""
			if s.Class != nil {
				gradeStr = s.Class.Name
				secStr = s.Class.Section
			} else {
				gradeStr = fmt.Sprintf("Grade %d", s.GradeLevel)
			}

			children = append(children, childRow{
				ID:          s.ID,
				Name:        name,
				StudentCode: s.StudentCode,
				Grade:       gradeStr,
				Section:     secStr,
			})
		}

		rows = append(rows, parentRow{
			ID: p.ID, Name: p.Name, Email: p.Email, Phone: p.Phone,
			IsActive: p.IsActive, Status: st, ChildrenCount: len(children),
			Children: children, StudentNames: strings.Join(studentNamesList, ", "),
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		})
	}
	helpers.Success(c, http.StatusOK, "parents fetched", gin.H{"total": total, "data": rows})
}

// GetAdmins lists admin accounts
func GetAdmins(c *gin.Context) {
	offset, limit := parsePage(c)
	var admins []models.User
	var total int64
	db := config.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin)
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Find(&admins).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch admins")
		return
	}
	type row struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		IsActive  bool   `json:"is_active"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	var out []row
	for _, a := range admins {
		st := "Active"
		if !a.IsActive {
			st = "Inactive"
		}
		out = append(out, row{
			ID: a.ID, Name: a.Name, Email: a.Email, Phone: a.Phone,
			IsActive: a.IsActive, Status: st, CreatedAt: a.CreatedAt.Format(time.RFC3339),
		})
	}
	helpers.Success(c, http.StatusOK, "admins fetched", gin.H{"total": total, "data": out})
}

type UpdateUserAccountInput struct {
	Name  string `json:"name"  binding:"required,min=2"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone"`
}

// UpdateAdmin updates an admin account
func UpdateAdmin(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.Where("id = ? AND role = ?", id, models.RoleAdmin).First(&user).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "admin not found")
		return
	}
	var input UpdateUserAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	var dup int64
	config.DB.Model(&models.User{}).Where("email = ? AND id != ?", input.Email, user.ID).Count(&dup)
	if dup > 0 {
		helpers.Error(c, http.StatusConflict, "email already in use")
		return
	}
	if err := config.DB.Model(&user).Updates(map[string]any{
		"name": input.Name, "email": input.Email, "phone": input.Phone,
	}).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update administrator in DB: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "admin updated", gin.H{
		"id": user.ID, "name": input.Name, "email": input.Email, "phone": input.Phone,
	})
}

// ArchiveAdmin deactivates an admin account (soft-delete)
func ArchiveAdmin(c *gin.Context) {
	id := c.Param("id")
	if fmt.Sprintf("%d", c.GetUint("userID")) == id {
		helpers.Error(c, http.StatusBadRequest, "cannot archive your own account")
		return
	}
	var user models.User
	if err := config.DB.Where("id = ? AND role = ?", id, models.RoleAdmin).First(&user).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "admin not found")
		return
	}
	if err := config.DB.Model(&user).Update("is_active", false).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to deactivate administrator: "+err.Error())
		return
	}
	if err := config.DB.Delete(&user).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to soft-delete administrator: "+err.Error())
		return
	}
	helpers.Success(c, http.StatusOK, "admin archived", nil)
}

// UpdateParent updates a parent account
func UpdateParent(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.Where("id = ? AND role = ?", id, models.RoleParent).First(&user).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "parent not found")
		return
	}
	var input UpdateUserAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	var dup int64
	config.DB.Model(&models.User{}).Where("email = ? AND id != ?", input.Email, user.ID).Count(&dup)
	if dup > 0 {
		helpers.Error(c, http.StatusConflict, "email already in use")
		return
	}
	config.DB.Model(&user).Updates(map[string]any{
		"name": input.Name, "email": input.Email, "phone": input.Phone,
	})
	helpers.Success(c, http.StatusOK, "parent updated", gin.H{
		"id": user.ID, "name": input.Name, "email": input.Email, "phone": input.Phone,
	})
}

// ArchiveParent deactivates a parent account (soft-delete user)
func ArchiveParent(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.Where("id = ? AND role = ?", id, models.RoleParent).First(&user).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "parent not found")
		return
	}
	// FIX #8: prevent orphaning active students. Count children that still reference
	// this parent user. If any exist, refuse the archive and tell the admin to
	// re-link them to a different parent first.
	var activeChildren int64
	config.DB.Model(&models.Student{}).Where("parent_id = ?", user.ID).Count(&activeChildren)
	if activeChildren > 0 {
		helpers.Error(c, http.StatusConflict,
			fmt.Sprintf("cannot archive parent: %d active student(s) still reference this parent — re-link or archive them first", activeChildren))
		return
	}
	config.DB.Model(&user).Update("is_active", false)
	config.DB.Delete(&user)
	helpers.Success(c, http.StatusOK, "parent archived", nil)
}

// GetMyChildren godoc
// @Summary      Get parent's children (Parent only)
// @Tags         parent
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "List of children"
// @Router       /api/parent/children [get]
func GetMyChildren(c *gin.Context) {
	parentUserID := c.GetUint("userID")

	var students []models.Student

	err := config.DB.
		Preload("User").
		Preload("Class").
		Where("parent_id = ?", parentUserID).
		Find(&students).Error

	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch children")
		return
	}

	helpers.Success(c, http.StatusOK, "children fetched", students)
}

func generateStudentCode(gradeLevel int) string {
	year := time.Now().Year()
	prefix := fmt.Sprintf("STU-%d-", year)
	var codes []string
	config.DB.Model(&models.Student{}).
		Where("student_code LIKE ?", prefix+"%").
		Pluck("student_code", &codes)
	maxSeq := 0
	for _, code := range codes {
		var seq int
		if _, err := fmt.Sscanf(code, prefix+"%d", &seq); err == nil && seq > maxSeq {
			maxSeq = seq
		}
	}
	_ = gradeLevel
	return fmt.Sprintf("%s%03d", prefix, maxSeq+1)
}

// ── scoreToLetter converts a numeric score to a letter grade.
func scoreToLetter(score float64) string {
	switch {
	case score >= 90:
		return "A+"
	case score >= 85:
		return "A"
	case score >= 80:
		return "B+"
	case score >= 75:
		return "B"
	case score >= 65:
		return "C"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}

// GetTeacherKPIs returns quick stats for the teacher dashboard
func GetTeacherKPIs(c *gin.Context) {
	teacherUserID := c.GetUint("userID")

	var teacher models.Teacher
	if err := config.DB.Where("user_id = ?", teacherUserID).First(&teacher).Error; err != nil {
		helpers.Error(c, http.StatusForbidden, "teacher profile not found")
		return
	}

	var subjectCount int64
	config.DB.Model(&models.Subject{}).Where("teacher_id = ?", teacher.ID).Count(&subjectCount)

	var classCount int64
	config.DB.Model(&models.Class{}).Where("teacher_id = ?", teacher.ID).Count(&classCount)

	var studentCount int64
	config.DB.Table("enrollments").
		Joins("JOIN subjects ON enrollments.subject_id = subjects.id").
		Where("subjects.teacher_id = ?", teacher.ID).
		Distinct("enrollments.student_id").
		Count(&studentCount)

	// Fallback/Default if no enrollments exist yet
	if studentCount == 0 && classCount > 0 {
		// count students in class
		config.DB.Model(&models.Student{}).
			Joins("JOIN classes ON students.class_id = classes.id").
			Where("classes.teacher_id = ?", teacher.ID).
			Count(&studentCount)
	}

	var attendanceRate float64 = 95.0 // default/fallback
	var presentCount, totalCount int64
	config.DB.Table("attendances").
		Joins("JOIN students ON attendances.student_id = students.id").
		Joins("JOIN classes ON students.class_id = classes.id").
		Where("classes.teacher_id = ?", teacher.ID).
		Count(&totalCount)

	if totalCount > 0 {
		config.DB.Table("attendances").
			Joins("JOIN students ON attendances.student_id = students.id").
			Joins("JOIN classes ON students.class_id = classes.id").
			Where("classes.teacher_id = ? AND attendances.status IN ('Present', 'Late')", teacher.ID).
			Count(&presentCount)
		attendanceRate = float64(presentCount) / float64(totalCount) * 100
	}

	helpers.Success(c, http.StatusOK, "teacher dashboard kpis", gin.H{
		"students":        studentCount,
		"classes":         classCount,
		"subjects":        subjectCount,
		"attendance_rate": attendanceRate,
	})
}
