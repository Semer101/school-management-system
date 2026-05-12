package controllers

import (
	"fmt"
	"net/http"
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

type AttendanceInput struct {
	StudentID uint   `json:"student_id" binding:"required"                        example:"1"`
	SubjectID uint   `json:"subject_id" binding:"required"                        example:"2"`
	Date      string `json:"date"       binding:"required"                        example:"2025-05-01"`
	Status    string `json:"status"     binding:"required,oneof=Present Absent Late" example:"Present"`
	Notes     string `json:"notes"                                                example:"Arrived 5 mins late"`
}

type GradeEntry struct {
	StudentID uint    `json:"student_id" binding:"required"          example:"1"`
	Score     float64 `json:"score"      binding:"required,min=0,max=100" example:"87.5"`
	Remarks   string  `json:"remarks"                                example:"Good improvement"`
}

type BulkGradeInput struct {
	SubjectID    uint         `json:"subject_id"    binding:"required"                              example:"3"`
	Type         string       `json:"type"          binding:"required,oneof=Midterm Final Quiz Assignment" example:"Midterm"`
	Term         string       `json:"term"          binding:"required,oneof=Term1 Term2 Term3"      example:"Term1"`
	AcademicYear int          `json:"academic_year" binding:"required"                              example:"2025"`
	MaxScore     float64      `json:"max_score"     binding:"required"                              example:"100"`
	Grades       []GradeEntry `json:"grades"        binding:"required,min=1"`
}

type CreateClassInput struct {
	Name      string `json:"name"       binding:"required" example:"Grade 10A"`
	Year      int    `json:"year"       binding:"required" example:"2025"`
	TeacherID uint   `json:"teacher_id"                    example:"1"`
}

type CreateSubjectInput struct {
	Name      string `json:"name"       binding:"required" example:"Mathematics"`
	Code      string `json:"code"       binding:"required" example:"MATH-101"`
	TeacherID uint   `json:"teacher_id"                    example:"1"`
}

type EnrollStudentInput struct {
	StudentID uint `json:"student_id" binding:"required" example:"1"`
	SubjectID uint `json:"subject_id" binding:"required" example:"2"`
}

// ══════════════════════════════════════════════════════
//
//	STUDENTS
//
// ══════════════════════════════════════════════════════

// CreateStudent godoc
// @Summary      Create a student
// @Description  Creates a new student account with a linked user login. Admin only.
// @Tags         students
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateStudentInput  true  "Student data"
// @Success      201   {object}  helpers.APIResponse   "Student created"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      409   {object}  helpers.APIResponse   "Email or student code already exists"
// @Router       /api/admin/students [post]
func CreateStudent(c *gin.Context) {
	var input CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var parent models.User
	if err := config.DB.First(&parent, input.ParentID).Error; err != nil {
		helpers.Error(c, http.StatusBadRequest, "invalid parent_id")
		return
	}

	var count int64
	config.DB.Unscoped().
		Model(&models.User{}).
		Where("email = ?", input.Email).
		Count(&count)

	if count > 0 {
		helpers.Error(c, http.StatusConflict, "email already registered")
		return
	}

	var codeCount int64
	config.DB.
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
// @Summary      List all students
// @Description  Returns all students with their user and class information. Admin only.
// @Tags         students
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "List of students"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/students [get]
func GetStudents(c *gin.Context) {
	var students []models.Student
	result := config.DB.Preload("User").Preload("Class").Find(&students)
	if result.Error != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to fetch students")
		return
	}
	helpers.Success(c, http.StatusOK, "students fetched", students)
}

// GetStudent godoc
// @Summary      Get a student by ID
// @Description  Returns a single student record. Admin only.
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
// @Summary      Update a student
// @Description  Updates class, parent info, or date of birth for a student. Admin only.
// @Tags         students
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "Student ID"
// @Param        body  body      UpdateStudentInput  true  "Fields to update"
// @Success      200   {object}  helpers.APIResponse   "Student updated"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      404   {object}  helpers.APIResponse   "Student not found"
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
		helpers.Error(c, http.StatusBadRequest, "no valid fields provided for update")
		return
	}

	if err := config.DB.Model(&student).Updates(updates).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to update student")
		return
	}

	helpers.Success(c, http.StatusOK, "student updated", student)
}

// ArchiveStudent godoc
// @Summary      Archive (soft-delete) a student
// @Description  Deactivates the student's user account and soft-deletes the student record. Admin only.
// @Tags         students
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Student ID"
// @Success      200  {object}  helpers.APIResponse  "Student archived"
// @Failure      404  {object}  helpers.APIResponse  "Student not found"
// @Router       /api/admin/students/{id} [delete]
func ArchiveStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", student.UserID).
			Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Delete(&student).Error
	})

	if txErr != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to archive student: "+txErr.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "student archived", nil)
}

// ══════════════════════════════════════════════════════
//
//	TEACHERS
//
// ══════════════════════════════════════════════════════

// CreateTeacher godoc
// @Summary      Create a teacher
// @Description  Creates a new teacher account with a linked user login. Admin only.
// @Tags         teachers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateTeacherInput  true  "Teacher data"
// @Success      201   {object}  helpers.APIResponse   "Teacher created"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      409   {object}  helpers.APIResponse   "Email or teacher code already exists"
// @Router       /api/admin/teachers [post]
func CreateTeacher(c *gin.Context) {
	var input CreateTeacherInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var count int64
	config.DB.Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		helpers.Error(c, http.StatusConflict, "email already registered")
		return
	}

	var codeCount int64
	config.DB.Model(&models.Teacher{}).
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

	err = config.DB.Transaction(func(tx *gorm.DB) error {
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

		if err := tx.Create(&teacher).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create teacher: "+err.Error())
		return
	}

	config.DB.Preload("User").First(&teacher, teacher.ID)

	helpers.Success(c, http.StatusCreated, "teacher created", gin.H{
		"user_id":    user.ID,
		"teacher_id": teacher.ID,
		"name":       user.Name,
		"code":       teacher.TeacherCode,
	})
}

// GetTeachers godoc
// @Summary      List all teachers
// @Description  Returns all teachers with their user information. Admin only.
// @Tags         teachers
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "List of teachers"
// @Failure      401  {object}  helpers.APIResponse  "Unauthorized"
// @Failure      403  {object}  helpers.APIResponse  "Forbidden — Admin only"
// @Router       /api/admin/teachers [get]
func GetTeachers(c *gin.Context) {
	var teachers []models.Teacher
	config.DB.Preload("User").Find(&teachers)
	helpers.Success(c, http.StatusOK, "teachers fetched", teachers)
}

// ══════════════════════════════════════════════════════
//
//	CLASSES & SUBJECTS
//
// ══════════════════════════════════════════════════════

// CreateClass godoc
// @Summary      Create a class
// @Description  Creates a new class and optionally assigns a teacher. Admin only.
// @Tags         classes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateClassInput   true  "Class data"
// @Success      201   {object}  helpers.APIResponse  "Class created"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Router       /api/admin/classes [post]
func CreateClass(c *gin.Context) {
	var input CreateClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	class := models.Class{Name: input.Name, Year: input.Year, TeacherID: input.TeacherID}
	config.DB.Create(&class)
	helpers.Success(c, http.StatusCreated, "class created", class)
}

// GetClasses godoc
// @Summary      List all classes
// @Description  Returns all classes with teacher and student info. Admin only.
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

// CreateSubject godoc
// @Summary      Create a subject
// @Description  Creates a new subject and optionally assigns a teacher. Admin only.
// @Tags         subjects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      CreateSubjectInput  true  "Subject data"
// @Success      201   {object}  helpers.APIResponse   "Subject created"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      409   {object}  helpers.APIResponse   "Subject code already exists"
// @Router       /api/admin/subjects [post]
func CreateSubject(c *gin.Context) {
	var input CreateSubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	subject := models.Subject{Name: input.Name, Code: input.Code, TeacherID: input.TeacherID}
	if err := config.DB.Create(&subject).Error; err != nil {
		helpers.Error(c, http.StatusConflict, "subject code already exists")
		return
	}
	helpers.Success(c, http.StatusCreated, "subject created", subject)
}

// GetSubjects godoc
// @Summary      List all subjects
// @Description  Returns all subjects with their assigned teacher. Admin only.
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

// ══════════════════════════════════════════════════════
//
//	ENROLLMENT
//
// ══════════════════════════════════════════════════════

// EnrollStudent godoc
// @Summary      Enroll student in a subject
// @Description  Enrolls a student into a subject. Prevents duplicate enrollments. Admin only.
// @Tags         enrollment
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      EnrollStudentInput  true  "Enrollment data"
// @Success      201   {object}  helpers.APIResponse   "Student enrolled"
// @Failure      400   {object}  helpers.APIResponse   "Validation error"
// @Failure      409   {object}  helpers.APIResponse   "Already enrolled"
// @Router       /api/admin/enroll [post]
func EnrollStudent(c *gin.Context) {
	var input EnrollStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
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
		helpers.Error(c, http.StatusConflict, "student already enrolled in this subject")
		return
	}
	helpers.Success(c, http.StatusCreated, "student enrolled", enrollment)
}

// ══════════════════════════════════════════════════════
//
//	ATTENDANCE
//
// ══════════════════════════════════════════════════════

// RecordAttendance godoc
// @Summary      Record attendance
// @Description  Records or updates attendance for a student in a subject on a given date. Teacher only.
// @Tags         attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      AttendanceInput    true  "Attendance data"
// @Success      201   {object}  helpers.APIResponse  "Attendance recorded"
// @Success      200   {object}  helpers.APIResponse  "Attendance updated (if record exists)"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Router       /api/academics/attendance [post]
func RecordAttendance(c *gin.Context) {
	var input AttendanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
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
// @Summary      Get class attendance for a date
// @Description  Returns all attendance records for a class on a specific date. Teacher only.
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        classID  path      int     true   "Class ID"
// @Param        date     query     string  true   "Date in YYYY-MM-DD format"
// @Success      200      {object}  helpers.APIResponse  "Attendance records"
// @Failure      400      {object}  helpers.APIResponse  "Missing or invalid date"
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
// @Description  Returns overall and per-subject attendance stats for a student. Accessible by Teacher, Student (own), Admin, Parent (own child).
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path      int  true  "Student ID"
// @Success      200        {object}  helpers.APIResponse  "Attendance summary"
// @Router       /api/academics/attendance/{studentID} [get]
func GetAttendancePercentage(c *gin.Context) {
	studentID := c.Param("studentID")
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
// @Summary      School-wide attendance summary
// @Description  Returns attendance percentage for every student across all classes. Admin only.
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
	config.DB.Raw(`
		SELECT
			st.id as student_id,
			u.name as student_name,
			cl.name as class_name,
			COUNT(a.id) as total,
			SUM(CASE WHEN a.status = 'Absent' THEN 1 ELSE 0 END) as absences,
			ROUND(SUM(CASE WHEN a.status IN ('Present','Late') THEN 1 ELSE 0 END) * 100.0 / COUNT(a.id), 2) as percentage
		FROM attendances a
		JOIN students st ON a.student_id = st.id
		JOIN users u ON st.user_id = u.id
		JOIN classes cl ON st.class_id = cl.id
		GROUP BY st.id, u.name, cl.name
		ORDER BY percentage ASC
	`).Scan(&summary)

	helpers.Success(c, http.StatusOK, "attendance summary", summary)
}

// ══════════════════════════════════════════════════════
//
//	GRADES
//
// ══════════════════════════════════════════════════════

// BulkGradeEntry godoc
// @Summary      Bulk enter grades
// @Description  Saves grades for multiple students for a subject and term in one request. Teacher only.
// @Tags         grades
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      BulkGradeInput     true  "Grades data"
// @Success      201   {object}  helpers.APIResponse  "Grades recorded"
// @Failure      400   {object}  helpers.APIResponse  "Validation error"
// @Router       /api/academics/grades/bulk [post]
func BulkGradeEntry(c *gin.Context) {
	teacherID := c.GetUint("userID")
	var input BulkGradeInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var grades []models.Grade
	for _, entry := range input.Grades {
		grades = append(grades, models.Grade{
			StudentID:    entry.StudentID,
			SubjectID:    input.SubjectID,
			TeacherID:    teacherID,
			Score:        entry.Score,
			MaxScore:     input.MaxScore,
			Type:         input.Type,
			Term:         input.Term,
			AcademicYear: input.AcademicYear,
			Remarks:      entry.Remarks,
		})
	}

	if err := config.DB.CreateInBatches(&grades, 100).Error; err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save grades")
		return
	}

	helpers.Success(c, http.StatusCreated, "grades recorded", gin.H{"saved": len(grades)})
}

// GetSubjectGrades godoc
// @Summary      Get grades for a subject
// @Description  Returns all grades for a subject, optionally filtered by type and term. Teacher only.
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        subjectID  path      int     true   "Subject ID"
// @Param        type       query     string  false  "Grade type: Midterm | Final | Quiz | Assignment"
// @Param        term       query     string  false  "Term: Term1 | Term2 | Term3"
// @Success      200        {object}  helpers.APIResponse  "Grade list"
// @Router       /api/academics/grades/subject/{subjectID} [get]
func GetSubjectGrades(c *gin.Context) {
	subjectID := c.Param("subjectID")
	gradeType := c.Query("type")
	term := c.Query("term")

	query := config.DB.Preload("Student.User").Where("subject_id = ?", subjectID)
	if gradeType != "" {
		query = query.Where("type = ?", gradeType)
	}
	if term != "" {
		query = query.Where("term = ?", term)
	}

	var grades []models.Grade
	query.Find(&grades)
	helpers.Success(c, http.StatusOK, "subject grades", grades)
}

// GetStudentGrades godoc
// @Summary      Get a student's grades
// @Description  Returns all grades for a student, optionally filtered by term and academic year. Accessible by Teacher, Student (own), Admin.
// @Tags         grades
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path      int     true   "Student ID"
// @Param        term       query     string  false  "Term: Term1 | Term2 | Term3"
// @Param        year       query     int     false  "Academic year e.g. 2025"
// @Success      200        {object}  helpers.APIResponse  "Grade list"
// @Router       /api/academics/grades/student/{studentID} [get]
func GetStudentGrades(c *gin.Context) {
	studentID := c.Param("studentID")
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
//
//	REPORT CARD
//
// ══════════════════════════════════════════════════════

// GetReportCard godoc
// @Summary      Get report card
// @Description  Returns a full report card with per-subject averages and letter grades. Accessible by Teacher, Student (own), Admin, Parent (own child).
// @Tags         report-card
// @Security     BearerAuth
// @Produce      json
// @Param        studentID  path      int  true  "Student ID"
// @Success      200        {object}  helpers.APIResponse  "Report card"
// @Failure      404        {object}  helpers.APIResponse  "Student not found"
// @Router       /api/academics/reportcard/{studentID} [get]
func GetReportCard(c *gin.Context) {
	studentID := c.Param("studentID")

	var student models.Student
	if err := config.DB.Preload("User").Preload("Class").First(&student, studentID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "student not found")
		return
	}

	type SubjectReport struct {
		SubjectName string  `json:"subject_name"`
		Midterm     float64 `json:"midterm"`
		Final       float64 `json:"final"`
		Assignments float64 `json:"assignments_avg"`
		Quizzes     float64 `json:"quizzes_avg"`
		Overall     float64 `json:"overall"`
		LetterGrade string  `json:"letter_grade"`
	}

	var report []SubjectReport
	config.DB.Raw(`
		SELECT
			s.name as subject_name,
			COALESCE(AVG(CASE WHEN g.type = 'Midterm' THEN g.score END), 0) as midterm,
			COALESCE(AVG(CASE WHEN g.type = 'Final' THEN g.score END), 0) as final,
			COALESCE(AVG(CASE WHEN g.type = 'Assignment' THEN g.score END), 0) as assignments_avg,
			COALESCE(AVG(CASE WHEN g.type = 'Quiz' THEN g.score END), 0) as quizzes_avg,
			COALESCE(AVG(g.score), 0) as overall
		FROM grades g
		JOIN subjects s ON g.subject_id = s.id
		WHERE g.student_id = ?
		GROUP BY s.name
	`, studentID).Scan(&report)

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
		"subjects": report,
	})
}

// DownloadReportCard godoc
// @Summary      Download report card as PDF
// @Description  Generates and downloads a PDF report card for a student. Accessible by Teacher, Student (own), Admin, Parent (own child).
// @Tags         report-card
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        studentID  path  int  true  "Student ID"
// @Success      200  {file}    binary               "PDF file"
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

	type row struct {
		SubjectName string
		Overall     float64
	}
	var rows []row
	config.DB.Raw(`
		SELECT s.name as subject_name, COALESCE(AVG(g.score), 0) as overall
		FROM grades g
		JOIN subjects s ON g.subject_id = s.id
		WHERE g.student_id = ?
		GROUP BY s.name
	`, student.ID).Scan(&rows)

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

	filename := fmt.Sprintf("report_%s.pdf", student.StudentCode)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ══════════════════════════════════════════════════════
//
//	PARENT
//
// ══════════════════════════════════════════════════════

// GetMyChildren godoc
// @Summary      Get my children
// @Description  Returns all students linked to the authenticated parent's account. Parent only.
// @Tags         parent
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  helpers.APIResponse  "List of children"
// @Failure      404  {object}  helpers.APIResponse  "User not found"
// @Router       /api/parent/children [get]
func GetMyChildren(c *gin.Context) {
	userID := c.GetUint("userID")

	var parent models.User
	if err := config.DB.First(&parent, userID).Error; err != nil {
		helpers.Error(c, http.StatusNotFound, "user not found")
		return
	}

	students := []models.Student{}
	config.DB.
		Preload("User").
		Preload("Class").
		Where("parent_email = ?", parent.Email).
		Find(&students)

	helpers.Success(c, http.StatusOK, "children fetched", students)
}

// scoreToLetter converts a numeric score to a letter grade.
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
