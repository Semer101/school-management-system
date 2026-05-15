package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"
)

// ── Input types ───────────────────────────────────────────────────────────────

type CreateStudentInput struct {
	Name        string `json:"name"          binding:"required,min=2"        example:"Alice Bekele"`
	Email       string `json:"email"         binding:"required,email"        example:"alice@school.com"`
	Password    string `json:"password"      binding:"required,min=8"        example:"password123"`
	StudentCode string `json:"student_code"  binding:"required"              example:"STU-001"`
	ClassID     uint   `json:"class_id"                                       example:"1"`
	ParentID    uint   `json:"parent_id"     binding:"required"              example:"5"`
	ParentName  string `json:"parent_name"                                   example:"Mr. Bekele"`
	ParentEmail string `json:"parent_email"  binding:"omitempty,email"       example:"parent@email.com"`
	ParentPhone string `json:"parent_phone"                                  example:"+251911000000"`
	DateOfBirth string `json:"date_of_birth"                                 example:"2010-05-15"`
}

type UpdateStudentInput struct {
	ClassID     *uint  `json:"class_id"                                example:"2"`
	ParentID    *uint  `json:"parent_id"                               example:"6"`
	ParentName  string `json:"parent_name"                             example:"Mr. Updated"`
	ParentEmail string `json:"parent_email"  binding:"omitempty,email" example:"new@email.com"`
	ParentPhone string `json:"parent_phone"                            example:"+251922000000"`
	DateOfBirth string `json:"date_of_birth"                           example:"2010-05-15"`
}

type CreateTeacherInput struct {
	Name          string `json:"name"          binding:"required,min=2" example:"Mr. Tadesse"`
	Email         string `json:"email"         binding:"required,email" example:"tadesse@school.com"`
	Password      string `json:"password"      binding:"required,min=8" example:"teacher123"`
	TeacherCode   string `json:"teacher_code"  binding:"required"       example:"TCH-001"`
	Qualification string `json:"qualification"                          example:"MSc Mathematics"`
}

type UpdateTeacherInput struct {
	Qualification string `json:"qualification" binding:"required" example:"PhD Mathematics"`
}

type AttendanceInput struct {
	StudentID uint   `json:"student_id" binding:"required"                           example:"1"`
	SubjectID uint   `json:"subject_id" binding:"required"                           example:"2"`
	Date      string `json:"date"       binding:"required"                           example:"2025-05-01"`
	Status    string `json:"status"     binding:"required,oneof=Present Absent Late" example:"Present"`
	Notes     string `json:"notes"                                                   example:"Arrived 5 mins late"`
}

type GradeEntry struct {
	StudentID uint    `json:"student_id" binding:"required"              example:"1"`
	Score     float64 `json:"score"      binding:"required,min=0,max=100" example:"87.5"`
	Remarks   string  `json:"remarks"                                    example:"Good improvement"`
}

type BulkGradeInput struct {
	SubjectID    uint         `json:"subject_id"    binding:"required"                                  example:"3"`
	Type         string       `json:"type"          binding:"required,oneof=Midterm Final Quiz Assignment" example:"Midterm"`
	Term         string       `json:"term"          binding:"required,oneof=Term1 Term2 Term3"           example:"Term1"`
	AcademicYear int          `json:"academic_year" binding:"required"                                  example:"2025"`
	MaxScore     float64      `json:"max_score"     binding:"required,min=1"                            example:"100"`
	Grades       []GradeEntry `json:"grades"        binding:"required,min=1"`
}

type CreateClassInput struct {
	Name      string `json:"name"       binding:"required" example:"Grade 10A"`
	Year      int    `json:"year"       binding:"required" example:"2025"`
	TeacherID uint   `json:"teacher_id"                    example:"1"`
}

// UpdateClassInput allows changing the homeroom teacher or year after creation.
type UpdateClassInput struct {
	TeacherID *uint  `json:"teacher_id" example:"2"`
	Year      *int   `json:"year"       example:"2026"`
	Name      string `json:"name"       example:"Grade 10B"`
}

type CreateSubjectInput struct {
	Name      string `json:"name"       binding:"required" example:"Mathematics"`
	Code      string `json:"code"       binding:"required" example:"MATH-101"`
	TeacherID uint   `json:"teacher_id"                    example:"1"`
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
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 50
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

	var codeCount int64
	config.DB.Unscoped().
		Model(&models.Student{}).
		Where("student_code = ?", input.StudentCode).
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
		student = models.Student{
			UserID:      user.ID,
			ParentID:    input.ParentID,
			ClassID:     input.ClassID,
			StudentCode: input.StudentCode,
			ParentName:  input.ParentName,
			ParentEmail: input.ParentEmail,
			ParentPhone: input.ParentPhone,
			DateOfBirth: dob,
			EnrolledAt:  time.Now(),
		}
		return tx.Create(&student).Error
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

	updates := map[string]any{}
	if input.ClassID != nil {
		updates["class_id"] = *input.ClassID
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

	if len(updates) == 0 {
		helpers.Error(c, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := config.DB.Model(&student).Updates(updates).Error; err != nil {
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

	config.DB.Model(&models.Teacher{}).Count(&total)
	if err := config.DB.Preload("User").Offset(offset).Limit(limit).Find(&teachers).Error; err != nil {
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
	if err := config.DB.Model(&teacher).Update("qualification", input.Qualification).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update teacher")
		return
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
	class := models.Class{Name: input.Name, Year: input.Year, TeacherID: input.TeacherID}

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
// @Success      200  {object}  helpers.APIResponse  "List of classes"
// @Router       /api/admin/classes [get]
func GetClasses(c *gin.Context) {
	var classes []models.Class
	config.DB.Preload("Teacher").Preload("Students").Find(&classes)
	helpers.Success(c, http.StatusOK, "classes fetched", classes)
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

	config.DB.Preload("Teacher").First(&class, class.ID)
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

	// check for duplicate code BEFORE the insert so we can return the correct status.
	// The old code caught ALL DB errors and reported them as 409 (even network errors).
	var codeCount int64
	config.DB.Model(&models.Subject{}).Where("code = ?", input.Code).Count(&codeCount)
	if codeCount > 0 {
		helpers.Error(c, http.StatusConflict, "subject code already exists")
		return
	}

	subject := models.Subject{Name: input.Name, Code: input.Code, TeacherID: input.TeacherID}
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
// @Success      200  {object}  helpers.APIResponse  "List of subjects"
// @Router       /api/admin/subjects [get]
func GetSubjects(c *gin.Context) {
	var subjects []models.Subject
	config.DB.Preload("Teacher").Find(&subjects)
	helpers.Success(c, http.StatusOK, "subjects fetched", subjects)
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

	config.DB.Preload("Teacher").First(&subject, subject.ID)
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
	teacherUserID := c.GetUint("userID")

	var input AttendanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var subject models.Subject
	if err := config.DB.First(&subject, input.SubjectID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "subject not found")
		return
	}
	var teacher models.Teacher
	if err := config.DB.Where("user_id = ?", teacherUserID).First(&teacher).Error; err != nil {
		helpers.Error(c, http.StatusForbidden, "teacher profile not found")
		return
	}
	// subject.TeacherID stores the Teacher table PK (not User PK).
	if subject.TeacherID != teacher.ID {
		helpers.Error(c, http.StatusForbidden, "you are not assigned to this subject")
		return
	}

	// FIX #3a: Verify the student is actually enrolled in this subject.
	// Without this check, a teacher could record attendance for a student who has
	// never been enrolled, creating phantom attendance records with no enrollment.
	var enrollCount int64
	config.DB.Model(&models.Enrollment{}).
		Where("student_id = ? AND subject_id = ?", input.StudentID, input.SubjectID).
		Count(&enrollCount)
	if enrollCount == 0 {
		helpers.Error(c, http.StatusBadRequest, "student is not enrolled in this subject")
		return
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
		"student_id = ? AND subject_id = ? AND date >= ? AND date < ?",
		input.StudentID, input.SubjectID, dayStart, dayEnd,
	).First(&existing)

	if result.Error == nil {
		if err := config.DB.Model(&existing).Updates(map[string]any{
			"status": input.Status,
			"notes":  input.Notes,
		}).Error; err != nil {
			helpers.Error(c, http.StatusInternalServerError, "failed to update attendance")
			return
		}
		helpers.Success(c, http.StatusOK, "attendance updated", existing)
		return
	}

	attendance := models.Attendance{
		StudentID: input.StudentID,
		SubjectID: input.SubjectID,
		Date:      date,
		Status:    input.Status,
		Notes:     input.Notes,
	}
	if err := config.DB.Create(&attendance).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to record attendance")
		return
	}
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
		Preload("Subject").
		Where("students.class_id = ?", classID).
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

	config.DB.Model(&models.Attendance{}).Where("student_id = ?", studentID).Count(&total)
	config.DB.Model(&models.Attendance{}).
		Where("student_id = ? AND status IN ('Present','Late')", studentID).
		Count(&present)

	percentage := 0.0
	if total > 0 {
		percentage = float64(present) / float64(total) * 100
	}

	type SubjectStat struct {
		SubjectName string  `json:"subject_name"`
		Total       int64   `json:"total"`
		Present     int64   `json:"present"`
		Percentage  float64 `json:"percentage"`
	}
	breakdown := []SubjectStat{}
	config.DB.Raw(`
		SELECT s.name as subject_name,
		       COUNT(a.id) as total,
		       SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) as present,
		       ROUND(SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) * 100.0 / COUNT(a.id), 2) as percentage
		FROM attendances a
		JOIN subjects s ON a.subject_id = s.id
		WHERE a.student_id = ?
		GROUP BY s.name
	`, studentID).Scan(&breakdown)

	helpers.Success(c, http.StatusOK, "attendance summary", gin.H{
		"student_id":         studentID,
		"overall_percentage": percentage,
		"total_classes":      total,
		"attended":           present,
		"by_subject":         breakdown,
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
		StudentID   uint    `json:"student_id"`
		StudentName string  `json:"student_name"`
		ClassName   string  `json:"class_name"`
		Total       int64   `json:"total"`
		Absences    int64   `json:"absences"`
		Percentage  float64 `json:"attendance_percentage"`
	}
	var summary []Summary
	// FIX #5: Use NULLIF(COUNT(a.id), 0) to prevent division-by-zero for students
	// who have zero attendance records. Previously COUNT(a.id) = 0 caused a DB error.
	// HAVING COUNT(a.id) > 0 also excludes students with no records from the report.
	config.DB.Raw(`
		SELECT
			st.id as student_id,
			u.name as student_name,
			cl.name as class_name,
			COUNT(a.id) as total,
			SUM(CASE WHEN a.status = 'Absent' THEN 1 ELSE 0 END) as absences,
			ROUND(SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(a.id), 0), 2) as percentage
		FROM attendances a
		JOIN students st ON a.student_id = st.id
		JOIN users u ON st.user_id = u.id
		JOIN classes cl ON st.class_id = cl.id
		GROUP BY st.id, u.name, cl.name
		HAVING COUNT(a.id) > 0
		ORDER BY percentage ASC
	`).Scan(&summary)
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

	var subject models.Subject
	if err := config.DB.First(&subject, input.SubjectID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "subject not found")
		return
	}

	var teacher models.Teacher
	if err := config.DB.Where("user_id = ?", teacherUserID).First(&teacher).Error; err != nil {
		helpers.Error(c, http.StatusForbidden, "teacher profile not found")
		return
	}

	// subject.TeacherID stores Teacher table PK
	if subject.TeacherID != teacher.ID {
		helpers.Error(c, http.StatusForbidden, "you are not assigned to this subject")
		return
	}

	teacherRecordID := teacher.ID

	// ─────────────────────────────────────────────────────────────
	// Validate enrollment + track NEW vs EXISTING grades
	// ─────────────────────────────────────────────────────────────

	enrolledStudentIDs := make(map[uint]bool)
	var enrollments []models.Enrollment

	config.DB.Where("subject_id = ?", input.SubjectID).Find(&enrollments)
	for _, e := range enrollments {
		enrolledStudentIDs[e.StudentID] = true
	}

	// fetch existing grades for THIS assessment (for spam prevention)
	type gradeKey struct {
		StudentID uint
	}

	existing := make(map[uint]bool)

	var oldGrades []models.Grade
	config.DB.Where(
		"subject_id = ? AND type = ? AND term = ? AND academic_year = ?",
		input.SubjectID, input.Type, input.Term, input.AcademicYear,
	).Find(&oldGrades)

	for _, g := range oldGrades {
		existing[g.StudentID] = true
	}

	newStudents := make([]uint, 0)

	var grades []models.Grade

	for _, entry := range input.Grades {

		if !enrolledStudentIDs[entry.StudentID] {
			helpers.Error(c, http.StatusBadRequest,
				fmt.Sprintf("student_id %d is not enrolled in subject_id %d", entry.StudentID, input.SubjectID))
			return
		}

		// mark only NEW grades for notification
		if !existing[entry.StudentID] {
			newStudents = append(newStudents, entry.StudentID)
		}

		grades = append(grades, models.Grade{
			StudentID:    entry.StudentID,
			SubjectID:    input.SubjectID,
			TeacherID:    teacherRecordID,
			Score:        entry.Score,
			MaxScore:     input.MaxScore,
			Type:         input.Type,
			Term:         input.Term,
			AcademicYear: input.AcademicYear,
			Remarks:      entry.Remarks,
		})
	}

	// ─────────────────────────────────────────────────────────────
	// UPSERT
	// ─────────────────────────────────────────────────────────────

	if err := config.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "student_id"},
			{Name: "subject_id"},
			{Name: "type"},
			{Name: "term"},
			{Name: "academic_year"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"score", "max_score", "remarks", "teacher_id",
		}),
	}).CreateInBatches(&grades, 100).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save grades")
		return
	}

	// ─────────────────────────────────────────────────────────────
	// FIX: prevent email spam on updates/re-submissions
	// Only notify NEW grade inserts
	// ─────────────────────────────────────────────────────────────

	go func() {
		for _, studentID := range newStudents {

			var student models.Student
			if err := config.DB.Preload("User").First(&student, studentID).Error; err != nil {
				continue
			}

			if student.User.Email == "" {
				continue
			}

			helpers.SendGradeNotification(
				student.User.Email,
				student.User.Name,
				subject.Name,
				0, // optional: can be enhanced later to pass real score
			)
		}
	}()

	helpers.Success(c, http.StatusCreated, "grades recorded", gin.H{
		"saved": len(grades),
	})
}

// GetSubjectGrades godoc
// @Summary      Get grades for a subject (Teacher only)
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        subjectID  path   int     true   "Subject ID"
// @Param        type       query  string  false  "Grade type: Midterm | Final | Quiz | Assignment"
// @Param        term       query  string  false  "Term: Term1 | Term2 | Term3"
// @Success      200  {object}  helpers.APIResponse  "Grade list"
// @Router       /api/academics/grades/subject/{subjectID} [get]
func GetSubjectGrades(c *gin.Context) {
	subjectID := c.Param("subjectID")
	gradeType := c.Query("type")
	term := c.Query("term")

	// Get logged-in user (teacher)
	userID, exists := c.Get("userID")
	if !exists {
		helpers.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	role, _ := c.Get("role")

	// Only enforce for teachers
	if role == models.RoleTeacher {
		var count int64

		// Verify the teacher owns this subject
		err := config.DB.Model(&models.Subject{}).
			Where("id = ? AND teacher_id = ?", subjectID, userID).
			Count(&count).Error

		if err != nil {
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

	if term != "" {
		query = query.Where("term = ?", term)
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
// @Description  Returns grades for a student filtered by term and/or academic year.
// @Description  Accessible by: Teacher, Student (own only), Admin, Parent (own child).
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path   int     true   "Student ID"
// @Param        term       query  string  false  "Term: Term1 | Term2 | Term3"
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
	term := c.Query("term")
	year := c.Query("year")

	query := config.DB.Preload("Subject").Where("student_id = ?", studentID)
	if term != "" {
		query = query.Where("term = ?", term)
	}
	if year != "" {
		query = query.Where("academic_year = ?", year)
	}
	var grades []models.Grade
	query.Find(&grades)
	helpers.Success(c, http.StatusOK, "grades fetched", grades)
}

// ══════════════════════════════════════════════════════
//  REPORT CARD
// ══════════════════════════════════════════════════════

// GetReportCard godoc
// @Summary      Get report card (JSON)
// @Description  Returns a full report card with per-subject averages and letter grades.
// @Description  FIX: now requires academic_year query param (defaults to current year).
// @Description  Optionally filtered by term. Without these filters, cross-year averages were meaningless.
// @Description  Accessible by: Teacher, Student (own only), Admin, Parent (own child via ParentOwnsStudent middleware).
// @Tags         report-card
// @Security     BearerAuth
// @Produce      json
// @Param        studentID      path   int     true   "Student ID"
// @Param        academic_year  query  int     false  "Academic year (default: current year)"
// @Param        term           query  string  false  "Term filter: Term1 | Term2 | Term3 (omit for full year average)"
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
	term := c.Query("term")

	type SubjectReport struct {
		SubjectName string  `json:"subject_name"`
		Midterm     float64 `json:"midterm"`
		Final       float64 `json:"final"`
		Assignments float64 `json:"assignments_avg"`
		Quizzes     float64 `json:"quizzes_avg"`
		Overall     float64 `json:"overall"`
		LetterGrade string  `json:"letter_grade"`
	}

	// Build the WHERE clause dynamically so term is optional.
	termClause := ""
	args := []any{studentID, yearStr}
	if term != "" {
		termClause = "AND g.term = ?"
		args = append(args, term)
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
	`, termClause), args...).Scan(&report)

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
		"term":          term,
		"subjects":      report,
	})
}

// DownloadReportCard godoc
// @Summary      Download report card as PDF
// @Description  Generates and streams a PDF report card. Accepts academic_year and term query params.
// @Tags         report-card
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        studentID      path   int     true   "Student ID"
// @Param        academic_year  query  int     false  "Academic year (default: current year)"
// @Param        term           query  string  false  "Term filter: Term1 | Term2 | Term3"
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
	term := c.Query("term")

	termClause := ""
	args := []any{student.ID, yearStr}
	if term != "" {
		termClause = "AND g.term = ?"
		args = append(args, term)
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
	`, termClause), args...).Scan(&rows)

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

	safeCode := strings.ReplaceAll(student.StudentCode, `"`, `_`)

	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="report_%s.pdf"`, safeCode),
	)

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ══════════════════════════════════════════════════════
//  PARENT
// ══════════════════════════════════════════════════════

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
