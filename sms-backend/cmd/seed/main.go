//go:build ignore

package main

// Run with: go run cmd/seed/main.go
// This is a SEPARATE binary from your main server

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env")
	}

	config.ConnectDB()
	log.Println("Starting database seed...")

	hashPwd := func(pwd string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pwd), 12)
		return string(h)
	}

	// ── 1. ADMINS ─────────────────────────────────────

	admins := []models.User{

		{Name: "Dawit Bekele", Email: "admin@school.et", Password: hashPwd("Admin@1234"), Role: models.RoleAdmin, IsActive: true},
		{Name: "Selam Haile", Email: "selam@school.et", Password: hashPwd("Admin@1234"), Role: models.RoleAdmin, IsActive: true},
	}
	for i := range admins {
		config.DB.Where("email = ?", admins[i].Email).FirstOrCreate(&admins[i])
	}
	log.Printf("Seeded %d admins", len(admins))

	// ── 2. TEACHERS ───────────────────────────────────

	teacherData := []struct {
		Name          string
		Email         string
		Code          string
		Qualification string
	}{

		{"Abebe Girma", "abebe.g@school.et", "TCH-001", "MSc Mathematics"},
		{"Tigist Alemu", "tigist.a@school.et", "TCH-002", "BSc Physics"},
		{"Yonas Tadesse", "yonas.t@school.et", "TCH-003", "BA English Literature"},
		{"Hiwot Mengistu", "hiwot.m@school.et", "TCH-004", "MSc Chemistry"},
		{"Biruk Kebede", "biruk.k@school.et", "TCH-005", "BSc Biology"},
		{"Meron Tesfaye", "meron.t@school.et", "TCH-006", "BA History"},
		{"Samuel Hailu", "samuel.h@school.et", "TCH-007", "MSc Computer Science"},
		{"Rahel Worku", "rahel.w@school.et", "TCH-008", "BSc Physical Education"},
	}

	var teachers []models.Teacher
	for _, td := range teacherData {
		user := models.User{Name: td.Name, Email: td.Email, Password: hashPwd("Teacher@1234"), Role: models.RoleTeacher, IsActive: true}
		config.DB.Where("email = ?", user.Email).FirstOrCreate(&user)

		teacher := models.Teacher{UserID: user.ID, TeacherCode: td.Code, Qualification: td.Qualification, JoinedAt: time.Now()}
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&teacher)
		teacher.User = user
		teachers = append(teachers, teacher)
	}
	log.Printf("Seeded %d teachers", len(teachers))

	// ── 3. CLASSES ────────────────────────────────────

	classData := []struct {
		Name      string
		TeacherID uint
	}{
		{"Grade 9A", teachers[0].ID},
		{"Grade 9B", teachers[1].ID},
		{"Grade 10A", teachers[2].ID},
		{"Grade 10B", teachers[3].ID},
		{"Grade 11A", teachers[4].ID},
	}

	var classes []models.Class
	for _, cd := range classData {
		class := models.Class{Name: cd.Name, Year: 2024, TeacherID: cd.TeacherID}
		config.DB.Where("name = ? AND year = ?", class.Name, class.Year).FirstOrCreate(&class)
		classes = append(classes, class)
	}
	log.Printf("Seeded %d classes", len(classes))

	// ── 4. SUBJECTS ───────────────────────────────────

	subjectData := []struct {
		Name      string
		Code      string
		TeacherID uint
	}{
		{"Mathematics", "MATH101", teachers[0].ID},
		{"Physics", "PHYS101", teachers[1].ID},
		{"English", "ENG101", teachers[2].ID},
		{"Chemistry", "CHEM101", teachers[3].ID},
		{"Biology", "BIO101", teachers[4].ID},
		{"History", "HIST101", teachers[5].ID},
		{"Computer Science", "CS101", teachers[6].ID},
		{"Physical Education", "PE101", teachers[7].ID},
	}

	var subjects []models.Subject
	for _, sd := range subjectData {
		subject := models.Subject{Name: sd.Name, Code: sd.Code, TeacherID: sd.TeacherID}
		config.DB.Where("code = ?", subject.Code).FirstOrCreate(&subject)
		subjects = append(subjects, subject)
	}
	log.Printf("Seeded %d subjects", len(subjects))

	// ── 5. STUDENTS ───────────────────────────────────

	studentNames := []string{
		"Abinet Tadesse", "Bemnet Girma", "Cherenet Haile", "Daniel Bekele", "Eyerusalem Alemu",
		"Fikir Mengistu", "Gelila Worku", "Henok Tesfaye", "Iman Kebede", "Jonas Hailu",
		"Kalekidan Girma", "Lidya Tadesse", "Mikias Haile", "Natnael Bekele", "Olivia Alemu",
		"Parsalem Worku", "Robel Mengistu", "Sara Tesfaye", "Tewodros Kebede", "Urael Hailu",
		"Vanessa Girma", "Winta Tadesse", "Xara Haile", "Yordanos Bekele", "Zara Alemu",
		"Abreham Worku", "Bezawit Mengistu", "Caleb Tesfaye", "Desta Kebede", "Eden Hailu",
		"Frehiwot Girma", "Girma Tadesse", "Hana Haile", "Ismael Bekele", "Jemila Alemu",
		"Kedir Worku", "Lemlem Mengistu", "Mahlet Tesfaye", "Negash Kebede", "Orkum Hailu",
		"Petros Girma", "Qemer Tadesse", "Rediet Haile", "Seble Bekele", "Tamrat Alemu",
		"Urge Worku", "Veronica Mengistu", "Wondwossen Tesfaye", "Yabsira Kebede", "Zinash Hailu",
	}

	var students []models.Student
	for i, name := range studentNames {
		email := fmt.Sprintf("student%d@school.et", i+1)
		code := fmt.Sprintf("STU-2024-%03d", i+1)
		classID := classes[i%len(classes)].ID

		parentName := "Parent of " + name
		parentEmail := fmt.Sprintf("parent%d@school.et", i+1)

		dob := time.Now().AddDate(-15-i%3, -i%12, 0)

		user := models.User{Name: name, Email: email, Password: hashPwd("Student@1234"), Role: models.RoleStudent, IsActive: true}
		config.DB.Where("email = ?", user.Email).FirstOrCreate(&user)

		student := models.Student{
			UserID: user.ID, StudentCode: code, ClassID: classID,
			ParentName: parentName, ParentEmail: parentEmail,
			ParentPhone: fmt.Sprintf("09%08d", i+10000000),
			DateOfBirth: dob, EnrolledAt: time.Now(),
		}
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&student)
		students = append(students, student)
	}
	log.Printf("Seeded %d students", len(students))

	// ── PARENTS — one per student ──────────────────────
	parentCount := 0
	for i, student := range students[:10] { // seed parents for first 10 students

		parentEmail := fmt.Sprintf("parent%d@school.et", i+1)

		// Parent User account
		parentUser := models.User{
			Name:     student.ParentName,
			Email:    parentEmail,
			Password: hashPwd("Parent@1234"),
			Role:     models.RoleParent,
			IsActive: true,
		}
		config.DB.Where("email = ?", parentEmail).FirstOrCreate(&parentUser)

		// Make sure student.parent_email matches the parent account email
		config.DB.Model(&student).Update("parent_email", parentEmail)
		parentCount++
	}
	log.Printf("Seeded %d parent accounts", parentCount)
	log.Println("Parent login: parent1@school.et / Parent@1234")

	// ── 6. ENROLLMENTS ────────────────────────────────

	enrollCount := 0
	for _, student := range students {
		// Each student enrolled in 5 subjects
		for j := 0; j < 5; j++ {
			subject := subjects[j%len(subjects)]
			var count int64
			config.DB.Model(&models.Enrollment{}).
				Where("student_id = ? AND subject_id = ?", student.ID, subject.ID).Count(&count)
			if count == 0 {
				config.DB.Create(&models.Enrollment{StudentID: student.ID, SubjectID: subject.ID})
				enrollCount++
			}
		}
	}
	log.Printf("Seeded %d enrollments", enrollCount)

	// ── 7. ATTENDANCE ─────────────────────────────────

	statuses := []string{"Present", "Present", "Present", "Present", "Absent", "Late"}
	attCount := 0
	startDate := time.Now().AddDate(0, -2, 0) // 2 months ago

	for _, student := range students[:10] { // seed for first 10 students to keep it fast
		for day := 0; day < 30; day++ {
			date := startDate.AddDate(0, 0, day)
			if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				continue
			}
			for _, subject := range subjects[:3] {
				status := statuses[(student.ID+uint(day)+subject.ID)%uint(len(statuses))]
				var count int64
				config.DB.Model(&models.Attendance{}).
					Where("student_id = ? AND subject_id = ? AND DATE(date) = DATE(?)",
						student.ID, subject.ID, date).Count(&count)
				if count == 0 {
					config.DB.Create(&models.Attendance{
						StudentID: student.ID, SubjectID: subject.ID,
						Date: date, Status: status,
					})
					attCount++
				}
			}
		}
	}
	log.Printf("Seeded %d attendance records", attCount)

	// ── 8. GRADES ─────────────────────────────────────

	gradeTypes := []string{"Midterm", "Final", "Quiz", "Assignment"}
	terms := []string{"Term1", "Term2"}
	gradeCount := 0

	for _, student := range students[:15] {
		for _, subject := range subjects[:4] {
			for _, term := range terms {
				for _, gType := range gradeTypes {
					score := 50.0 + float64((student.ID+subject.ID+uint(len(gType)))%50)
					var count int64
					config.DB.Model(&models.Grade{}).
						Where("student_id = ? AND subject_id = ? AND term = ? AND type = ?",
							student.ID, subject.ID, term, gType).Count(&count)
					if count == 0 {
						config.DB.Create(&models.Grade{
							StudentID: student.ID, SubjectID: subject.ID,
							TeacherID: subject.TeacherID, Score: score,
							MaxScore: 100, Type: gType, Term: term, AcademicYear: 2024,
						})
						gradeCount++
					}
				}
			}
		}
	}
	log.Printf("Seeded %d grade records", gradeCount)

	// ── 9. TRANSACTIONS ───────────────────────────────

	txCount := 0
	for i, student := range students[:20] {
		receiptID := fmt.Sprintf("ETH-BANK-%06d", 100000+i)
		var count int64
		config.DB.Model(&models.Transaction{}).Where("receipt_id = ?", receiptID).Count(&count)
		if count == 0 {
			status := "Pending"
			if i%3 == 0 {
				status = "Verified"
			}
			config.DB.Create(&models.Transaction{
				StudentID: student.ID, Amount: 5000.00,
				ReceiptID: receiptID, Type: "Tuition",
				Status: status, Description: "Semester 1 tuition fee",
				CreatedBy: student.UserID,
			})
			txCount++
		}
	}
	log.Printf("Seeded %d transactions", txCount)

	// ── 10. PAYROLL ───────────────────────────────────

	payCount := 0
	for _, teacher := range teachers {
		for month := 1; month <= 4; month++ {
			var count int64
			config.DB.Model(&models.Payroll{}).
				Where("teacher_id = ? AND month = ? AND year = ?", teacher.ID, month, 2024).Count(&count)
			if count == 0 {
				status := "Paid"
				if month == 4 {
					status = "Pending"
				}
				config.DB.Create(&models.Payroll{
					TeacherID: teacher.ID, Amount: 8500.00,
					Month: month, Year: 2024, Status: status,
				})
				payCount++
			}
		}
	}
	log.Printf("Seeded %d payroll records", payCount)

	log.Println("=== Seed complete! ===")
	log.Println("Admin login:   admin@school.et / Admin@1234")
	log.Println("Teacher login: abebe.g@school.et / Teacher@1234")
	log.Println("Student login: student1@school.et / Student@1234")
	log.Println("Parent login:  parent1@school.et / Parent@1234")
}
