//go:build ignore

package main

// Usage:
//   go run cmd/seed/main.go          — trim excess data, then seed production-quality data
//   go run cmd/seed/main.go -trim   — trim only (no re-seed)
//   go run cmd/seed/main.go -full   — generate 50 students with full records

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/models"
)

const (
	sampleStudents = 50
	sampleTeachers = 15
	sampleParents  = 25
	sampleAdmins   = 3
)

var (
	// Realistic Ethiopian first names
	maleFirstNames = []string{
		"Abinet", "Bemnet", "Cherenet", "Daniel", "Ephrem",
		"Fikru", "Girma", "Henok", "Isayas", "Jonas",
		"Kaleb", "Lalibela", "Mesfin", "Nahom", "Ogbamariam",
		"Paulos", "Qetsela", "Robel", "Samuel", "Tadesse",
		"Uthman", "Vesesub", "Wondafrash", "Yared", "Zerihun",
		"Abel", "Berhanu", "Dawit", "Elias", "Filimon",
		"Gebru", "Habtamu", "Ibrahim", "Jemberu", "Kassahun",
		"Lemi", "Mulugeta", "Nega", "Obang", "Petros",
		"Rahman", "Solomon", "Tewodros", "Ukullu", "Veta",
		"Wubishet", "Yohanis", "Zewde", "Aron", "Betekle"
	}
	femaleFirstNames = []string{
		"Abeba", "Birtukan", "Chaltu", "Danawork", "Eden",
		"Fikirte", "Gelila", "Hana", "Iman", "Jember",
		"Kidist", "Liya", "Meron", "Nardos", "Opal",
		"Rahel", "Sara", "Tigist", "Ubah", "Violet",
		"Woinshet", "Yabsira", "Zewditu", "Almaz", "Belaynesh",
		"Chimedes", "Desta", "Ejegayehu", "Frew", "Genet",
		"Hiwot", "Indermariat", "Jamila", "Kalid", "Lulseged",
		"Meseret", "Nebiyat", "Olivia", "Persia", "Rediat",
		"Selam", "Tirhas", "Umber", "Veronica", "Worknesh",
		"Yeshi", "Zeneba", "Amarech", "Belay", "Chaltu"
	}
	lastNames = []string{
		"Bekele", "Girma", "Haile", "Alemu", "Tadesse",
		"Mengistu", "Worku", "Tesfaye", "Assefa", "Desta",
		"Kebede", "Hailu", "Gebremedhin", "Berhanu", "Fekadu",
		"Wondemu", "Abebe", "Lemma", "Tekle", "Aragaw",
		"Shiferaw", "Mulugeta", "Negash", "Asfaw", "Gebre",
		"Kassa", "Tamiru", "Wondalessie", "Gidey", "Beyene",
		"Ayalew", "Habte", "Yilma", "Dejene", "Tilahun",
		"Workineh", "Goshu", "Bizuayehu", "Abate", "Kibru"
	}
	departments = []string{
		"Mathematics", "Physics", "Chemistry", "Biology",
		"English", "Amharic", "Social Studies", "Civics", "ICT",
	}
	qualifications = []string{
		"BEd Mathematics", "MSc Physics", "BEd Chemistry", "BSc Biology",
		"MEd English", "BA Amharic", "BEd Social Studies", "MA Civics",
		"BSc ICT", "MEd Mathematics", "PhD Chemistry", "MSc Biology",
	}
	streams = []string{
		models.StreamNatural, models.StreamNatural, models.StreamSocial,
	}
)

func main() {
	trimOnly := flag.Bool("trim", false, "only remove excess data, do not seed")
	fullMode := flag.Bool("full", false, "generate production-quality data (50 students)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file")
	}
	config.ConnectDB()

	trimExcessData()
	if *trimOnly {
		log.Println("Trim-only complete. Database cleaned to sample size.")
		os.Exit(0)
	}

	studentCount := sampleStudents
	if *fullMode {
		studentCount = 50
	}

	log.Printf("Seeding production-quality data: %d students, %d teachers, %d parents...",
		studentCount, sampleTeachers, sampleParents)
	seedProductionData(studentCount)
}

func hashPwd(pwd string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(h)
}

func randFromSlice(s []string) string {
	return s[rand.Intn(len(s))]
}

func seedProductionData(studentCount int) {
	academicYear := 2025
	rand.Seed(time.Now().UnixNano())

	// ── 1. Admins ─────────────────────────────────────────
	adminDefs := []struct {
		Name, Email, Phone string
	}{
		{"Dawit Bekele", "admin@school.et", "0911234567"},
		{"Selam Haile", "selam@school.et", "0922345678"},
		{"Tewodros Asfaw", "admin3@school.et", "0933456789"},
	}
	var admins []models.User
	for i := 0; i < sampleAdmins && i < len(adminDefs); i++ {
		d := adminDefs[i]
		u := models.User{
			Name: d.Name, Email: d.Email,
			Password: hashPwd("Admin@1234"), Role: models.RoleAdmin,
			Phone: d.Phone, IsActive: true,
		}
		config.DB.Where("email = ?", u.Email).FirstOrCreate(&u)
		config.DB.Model(&u).Updates(map[string]any{"name": d.Name, "phone": d.Phone, "is_active": true})
		config.DB.Where("email = ?", u.Email).First(&u)
		admins = append(admins, u)
	}

	// ── 2. Teachers ───────────────────────────────────────
	var teachers []models.Teacher
	for i := 1; i <= sampleTeachers; i++ {
		firstName := randFromSlice(maleFirstNames)
		lastName := randFromSlice(lastNames)
		email := fmt.Sprintf("teacher%d@school.et", i)
		dept := departments[(i-1)%len(departments)]
		qual := qualifications[(i-1)%len(qualifications)]

		u := models.User{
			Name: fmt.Sprintf("%s %s", firstName, lastName), Email: email,
			Password: hashPwd("Teacher@1234"), Role: models.RoleTeacher,
			Phone: fmt.Sprintf("0911100%03d", i), IsActive: true,
		}
		config.DB.Where("email = ?", u.Email).FirstOrCreate(&u)
		config.DB.Model(&u).Updates(map[string]any{
			"name": u.Name, "phone": u.Phone, "is_active": true,
		})

		t := models.Teacher{
			UserID: u.ID, TeacherCode: fmt.Sprintf("TCH-%04d", i),
			Qualification: qual, Department: dept,
			JoinedAt: time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC),
		}
		config.DB.Where("user_id = ?", u.ID).FirstOrCreate(&t)
		config.DB.Model(&t).Updates(map[string]any{
			"teacher_code": t.TeacherCode, "qualification": qual,
			"department": dept, "joined_at": t.JoinedAt,
		})
		config.DB.Where("user_id = ?", u.ID).First(&t)
		t.User = u
		teachers = append(teachers, t)
	}

	// ── 3. Classes ────────────────────────────────────────
	// Ethiopian curriculum: Grade 9–12, streams for 11–12
	classDefs := []struct {
		Grade           int
		Section, Stream string
	}{
		{9, "A", ""}, {9, "B", ""},
		{10, "A", ""}, {10, "B", ""},
		{11, "A", models.StreamNatural}, {11, "B", models.StreamNatural},
		{11, "C", models.StreamSocial},
		{12, "A", models.StreamNatural}, {12, "B", models.StreamSocial},
		{12, "C", models.StreamNatural},
	}
	var classes []models.Class
	for i, cd := range classDefs {
		var name string
		if cd.Stream != "" {
			name = fmt.Sprintf("%d%s %s", cd.Grade, cd.Section, cd.Stream)
		} else {
			name = fmt.Sprintf("%d%s", cd.Grade, cd.Section)
		}
		tID := teachers[i%len(teachers)].ID
		c := models.Class{
			Name: name, GradeLevel: cd.Grade, Section: cd.Section, Stream: cd.Stream,
			Status: "Active", Year: academicYear, TeacherID: &tID,
		}
		config.DB.Where("name = ? AND year = ?", name, academicYear).FirstOrCreate(&c)
		config.DB.Model(&c).Updates(map[string]any{"teacher_id": tID, "status": "Active"})
		config.DB.Where("name = ? AND year = ?", name, academicYear).First(&c)
		classes = append(classes, c)
	}

	// ── 4. Subjects (Ethiopian Grades 9–12 curriculum) ───
	var subjects []models.Subject
	for _, cur := range models.EthiopianGrades9to12Subjects() {
		for _, grade := range cur.Grades {
			if grade <= 10 && cur.Stream != "" {
				continue
			}
			tID := teachers[cur.TeacherIdx%len(teachers)].ID
			code := models.CurriculumSubjectCode(cur.Code, cur.Stream, grade)
			subjectName := cur.Name
			if cur.Stream != "" {
				subjectName = fmt.Sprintf("%s (%s)", cur.Name, cur.Stream)
			}
			sub := models.Subject{
				Name: subjectName, Code: code, GradeLevel: grade, Stream: cur.Stream,
				Status: "Active", TeacherID: &tID,
			}
			config.DB.Where("code = ?", code).FirstOrCreate(&sub)
			config.DB.Model(&sub).Updates(map[string]any{
				"name": subjectName, "grade_level": grade, "stream": cur.Stream,
				"status": "Active", "teacher_id": tID,
			})
			config.DB.First(&sub, sub.ID)
			subjects = append(subjects, sub)
		}
	}

	// ── 5. Parents ────────────────────────────────────────
	var parents []models.User
	for i := 1; i <= sampleParents; i++ {
		isMother := rand.Intn(2) == 0
		var firstName string
		if isMother {
			firstName = randFromSlice(femaleFirstNames)
		} else {
			firstName = randFromSlice(maleFirstNames)
		}
		lastName := randFromSlice(lastNames)
		email := fmt.Sprintf("parent%d@school.et", i)
		u := models.User{
			Name: fmt.Sprintf("%s %s", firstName, lastName), Email: email,
			Password: hashPwd("Parent@1234"), Role: models.RoleParent,
			Phone: fmt.Sprintf("0944%06d", 100000+i), IsActive: true,
		}
		config.DB.Where("email = ?", u.Email).FirstOrCreate(&u)
		config.DB.Model(&u).Updates(map[string]any{"name": u.Name, "is_active": true})
		config.DB.Where("email = ?", u.Email).First(&u)
		parents = append(parents, u)
	}

	// ── 6. Students ───────────────────────────────────────
	var students []models.Student
	for i := 1; i <= studentCount; i++ {
		isMale := rand.Intn(2) == 0
		var firstName string
		if isMale {
			firstName = randFromSlice(maleFirstNames)
		} else {
			firstName = randFromSlice(femaleFirstNames)
		}
		lastName := randFromSlice(lastNames)
		email := fmt.Sprintf("student%d@school.et", i)

		// Assign to a class (cycle through classes, weighted toward lower grades)
		classIdx := (i - 1) % len(classes)
		if rand.Float64() < 0.3 {
			classIdx = rand.Intn(len(classes))
		}
		assignedClass := classes[classIdx]

		parentIdx := (i - 1) % len(parents)
		if i%7 == 0 {
			// Some students share parents (siblings)
			parentIdx = (i - 2) % len(parents)
			if parentIdx < 0 {
				parentIdx = 0
			}
		}

		u := models.User{
			Name: fmt.Sprintf("%s %s", firstName, lastName), Email: email,
			Password: hashPwd("Student@1234"), Role: models.RoleStudent,
			Phone: fmt.Sprintf("0911500%03d", i), IsActive: true,
		}
		config.DB.Where("email = ?", u.Email).FirstOrCreate(&u)
		config.DB.Model(&u).Updates(map[string]any{"name": u.Name, "is_active": true})

		// Generate realistic date of birth based on grade level
		dobYear := academicYear - assignedClass.GradeLevel - 3
		dob := time.Date(dobYear, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC)

		st := models.Student{
			UserID: u.ID, ParentID: parents[parentIdx].ID, ClassID: &assignedClass.ID,
			StudentCode: fmt.Sprintf("STU-%d-%04d", academicYear, i),
			ParentName:  parents[parentIdx].Name, ParentEmail: parents[parentIdx].Email,
			ParentPhone: parents[parentIdx].Phone,
			DateOfBirth:  dob,
			Stream:       assignedClass.Stream, GradeLevel: assignedClass.GradeLevel,
			PromotionStatus: models.PromotionNormal,
			AcademicYear:    academicYear,
			EnrolledAt:      time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC),
		}
		config.DB.Where("user_id = ?", u.ID).FirstOrCreate(&st)
		config.DB.Model(&st).Updates(map[string]any{
			"class_id":         assignedClass.ID,
			"grade_level":      assignedClass.GradeLevel,
			"stream":           assignedClass.Stream,
			"parent_id":        parents[parentIdx].ID,
			"parent_name":      parents[parentIdx].Name,
			"parent_email":     parents[parentIdx].Email,
			"parent_phone":     parents[parentIdx].Phone,
			"academic_year":    academicYear,
			"enrolled_at":      time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC),
			"promotion_status": models.PromotionNormal,
		})
		config.DB.Where("user_id = ?", u.ID).First(&st)
		students = append(students, st)
	}

	// ── 7. Enrollments (students → grade-level subjects) ─
	for _, st := range students {
		config.DB.Where("student_id = ?", st.ID).Delete(&models.Enrollment{})
		for _, sub := range subjects {
			if sub.GradeLevel != st.GradeLevel {
				continue
			}
			if sub.Stream != "" && sub.Stream != st.Stream {
				continue
			}
			config.DB.Create(&models.Enrollment{StudentID: st.ID, SubjectID: sub.ID})
		}
	}

	// ── 8. Attendance (30 school days, realistic patterns) ─
	// ~80% Present, ~13% Late, ~7% Absent — varied per student
	attStatuses := []string{
		"Present", "Present", "Present", "Present", "Present",
		"Present", "Present", "Present", "Present", "Present",
		"Present", "Present", "Late", "Late",
		"Absent",
	}
	attStart := time.Now().AddDate(0, 0, -42)
	for _, st := range students {
		schoolDays := 0
		for day := 0; schoolDays < 30; day++ {
			d := attStart.AddDate(0, 0, day)
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			schoolDays++
			var cnt int64
			config.DB.Model(&models.Attendance{}).
				Where("student_id = ? AND subject_id IS NULL AND DATE(date) = DATE(?)", st.ID, d).Count(&cnt)
			if cnt == 0 {
				pick := (int(st.ID) + schoolDays) % len(attStatuses)
				config.DB.Create(&models.Attendance{
					StudentID: st.ID, SubjectID: nil, Date: d, Status: attStatuses[pick],
				})
			}
		}
	}

	// ── 9. Grades (Midterm + Final for each enrolled subject/semester) ─
	for i, st := range students {
		var subs []models.Subject
		config.DB.Joins("JOIN enrollments ON enrollments.subject_id = subjects.id").
			Where("enrollments.student_id = ?", st.ID).Find(&subs)
		for _, sub := range subs {
			semsToSeed := []string{"Semester 1", "Semester 2"}
			if i <= 2 {
				semsToSeed = append(semsToSeed, "Semester 3")
			}
			for _, sem := range semsToSeed {
				for _, gt := range []string{"Midterm", "Final"} {
					// Realistic grade distribution: bell curve around 72
					score := 72.0 + float64((st.ID+sub.ID)%25)
					// Add some randomness
					score += float64(rand.Intn(11) - 5) // ±5
					if score < 20 {
						score = 20
					}
					if score > 100 {
						score = 100
					}
					// Edge cases for demonstration
					if i == 0 {
						score = 40.0 // Abinet - Repeat
					} else if i == 1 && len(subs) > 0 && sub.ID == subs[0].ID {
						score = 42.0 // Bemnet - Conditional
					} else if i == 3 && gt == "Final" {
						score = 42.0
					}
					var gc int64
					config.DB.Model(&models.Grade{}).Where(
						"student_id = ? AND subject_id = ? AND grade_type = ? AND semester = ? AND academic_year = ?",
						st.ID, sub.ID, gt, sem, academicYear,
					).Count(&gc)
					if gc == 0 {
						var teacherID uint
						if sub.TeacherID != nil {
							teacherID = *sub.TeacherID
						}
						config.DB.Create(&models.Grade{
							StudentID: st.ID, SubjectID: sub.ID, TeacherID: teacherID,
							Score: score, MaxScore: 100, Type: gt, Semester: sem,
							AcademicYear: academicYear,
						})
					}
				}
			}
		}
	}

	// ── 10. Notifications ─────────────────────────────────
	notifDefs := []struct{ Title, Body string }{
		{"Welcome 2025", "New academic year for Grades 9–12. Welcome back!"},
		{"Parent Meeting", "Friday 2pm — all parents invited to discuss student progress."},
		{"Fee Deadline", "Tuition fees for Semester 1 are due by end of month."},
		{"Exam Preparation", "Grade 12 national exam preparation week begins next Monday."},
		{"Sports Day", "Inter-class sports competition on Meskerem 20. All students welcome!"},
		{"Library Hours Extended", "Library will be open until 5pm during exam period."},
		{"Science Fair", "Annual science fair submissions due next week."},
		{"Teacher Training", "Professional development day — no classes on Friday."},
	}
	for i, nd := range notifDefs {
		var cnt int64
		config.DB.Model(&models.Notification{}).Where("title = ?", nd.Title).Count(&cnt)
		if cnt == 0 {
			notif := models.Notification{
				Title: nd.Title, Body: nd.Body, TargetRoles: "Student,Parent",
				SenderID: admins[0].ID,
			}
			config.DB.Create(&notif)
			for _, st := range students {
				var rc int64
				config.DB.Model(&models.NotificationReceipt{}).
					Where("notification_id = ? AND user_id = ?", notif.ID, st.UserID).Count(&rc)
				if rc == 0 {
					config.DB.Create(&models.NotificationReceipt{
						NotificationID: notif.ID, UserID: st.UserID, IsRead: i%2 == 0,
					})
				}
			}
		}
	}

	// ── 11. Finance (Transactions) ────────────────────────
	for i, student := range students {
		// Semester 1 payment (everyone gets one)
		rid1 := fmt.Sprintf("ETH-CBE-%06d", 300000+i)
		var tc1 int64
		config.DB.Model(&models.Transaction{}).Where("receipt_id = ?", rid1).Count(&tc1)
		if tc1 == 0 {
			txStatus := "Pending"
			if i%3 == 0 {
				txStatus = "Verified"
			}
			config.DB.Create(&models.Transaction{
				StudentID: student.ID, Amount: 8500, ReceiptID: rid1, Type: "Tuition",
				Status: txStatus, Description: "Semester 1 Tuition",
				CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 1",
			})
		}

		// Semester 2 payment (subset to demonstrate overdue states)
		if i%2 == 0 {
			rid2 := fmt.Sprintf("ETH-CBE-%06d", 400000+i)
			var tc2 int64
			config.DB.Model(&models.Transaction{}).Where("receipt_id = ?", rid2).Count(&tc2)
			if tc2 == 0 {
				txStatus := "Pending"
				if i%4 == 0 {
					txStatus = "Verified"
				}
				config.DB.Create(&models.Transaction{
					StudentID: student.ID, Amount: 8500, ReceiptID: rid2, Type: "Tuition",
					Status: txStatus, Description: "Semester 2 Tuition",
					CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 2",
				})
			}
		}

		// Semester 3 payment (only for first 10 students — top grades)
		if i < 10 {
			rid3 := fmt.Sprintf("ETH-CBE-%06d", 500000+i)
			var tc3 int64
			config.DB.Model(&models.Transaction{}).Where("receipt_id = ?", rid3).Count(&tc3)
			if tc3 == 0 {
				txStatus := "Pending"
				if i%2 == 0 {
					txStatus = "Verified"
				}
				config.DB.Create(&models.Transaction{
					StudentID: student.ID, Amount: 8500, ReceiptID: rid3, Type: "Tuition",
					Status: txStatus, Description: "Semester 3 Tuition",
					CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 3",
				})
			}
		}
	}

	// ── 12. Payroll (6 months per teacher) ────────────────
	for _, t := range teachers {
		for m := 1; m <= 6; m++ {
			var pc int64
			config.DB.Model(&models.Payroll{}).
				Where("teacher_id = ? AND month = ? AND year = ?", t.ID, m, academicYear).Count(&pc)
			if pc == 0 {
				paidAt := time.Now().AddDate(0, -(6 - m), 0)
				status := "Paid"
				if m >= 5 {
					status = "Pending"
					paidAt = time.Time{}
				}
				p := models.Payroll{
					TeacherID: t.ID, Amount: 12000 + float64(m*200),
					Month: m, Year: academicYear, Status: status,
				}
				if status == "Paid" {
					p.PaidAt = &paidAt
				}
				config.DB.Create(&p)
			}
		}
	}

	// ── 13. Locker files (2 per student) ──────────────────
	for _, student := range students {
		var lfc int64
		config.DB.Model(&models.LockerFile{}).Where("student_id = ?", student.ID).Count(&lfc)
		if lfc == 0 {
			config.DB.Create(&models.LockerFile{
				StudentID: student.ID,
				FileName:  fmt.Sprintf("Grade_%d_Portfolio.pdf", student.GradeLevel),
				FilePath:  fmt.Sprintf("./uploads/locker/portfolio_%d.pdf", student.ID),
				FileSize:  102400, FileType: "pdf", Category: "Portfolio",
				IsPublic:   true,
				UploadedAt: time.Now(),
			})
			config.DB.Create(&models.LockerFile{
				StudentID: student.ID,
				FileName:  "Community_Service_Certificate.jpg",
				FilePath:  fmt.Sprintf("./uploads/locker/cert_%d.jpg", student.ID),
				FileSize:  204800, FileType: "jpg", Category: "Certificate",
				IsPublic:   false,
				UploadedAt: time.Now(),
			})
		}
	}

	log.Println("=== Seed complete ===")
	log.Printf("  Admins:     %d", sampleAdmins)
	log.Printf("  Teachers:    %d", sampleTeachers)
	log.Printf("  Parents:     %d", sampleParents)
	log.Printf("  Students:    %d", studentCount)
	log.Printf("  Classes:     %d", len(classes))
	log.Printf("  Subjects:    %d", len(subjects))
	log.Println()
	log.Println("Credentials:")
	log.Println("  admin@school.et     / Admin@1234")
	log.Println("  teacher1@school.et  / Teacher@1234")
	log.Println("  student1@school.et  / Student@1234")
	log.Println("  parent1@school.et   / Parent@1234")
}

func trimExcessData() {
	log.Println("Trimming database (removing bulk/legacy seed rows)...")
	logTableCounts("before trim")

	// Hard-remove legacy grade 9/10 classes with stream
	var legacyClassIDs []uint
	config.DB.Unscoped().Model(&models.Class{}).
		Where("grade_level IN (9, 10) AND stream != ''").Pluck("id", &legacyClassIDs)
	if len(legacyClassIDs) > 0 {
		log.Printf("Removing %d legacy grade 9/10 stream classes", len(legacyClassIDs))
		config.DB.Where("class_id IN ?", legacyClassIDs).Delete(&models.Student{})
		config.DB.Unscoped().Where("id IN ?", legacyClassIDs).Delete(&models.Class{})
	}

	// Fix orphaned teacher references
	var fallbackTeacher uint
	if config.DB.Model(&models.Teacher{}).Order("id").Limit(1).Pluck("id", &fallbackTeacher); fallbackTeacher > 0 {
		config.DB.Exec("UPDATE classes SET teacher_id = ? WHERE teacher_id NOT IN (SELECT id FROM teachers)", fallbackTeacher)
		config.DB.Exec("UPDATE subjects SET teacher_id = ? WHERE teacher_id NOT IN (SELECT id FROM teachers)", fallbackTeacher)
	}

	// Clean orphaned rows
	config.DB.Exec(`
		DELETE FROM attendances WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM grades WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM enrollments WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM transactions WHERE student_id NOT IN (SELECT id FROM students);
	`)

	// Cap attendance per student
	var keptStudentIDs []uint
	config.DB.Model(&models.Student{}).Pluck("id", &keptStudentIDs)
	for _, sid := range keptStudentIDs {
		var excess []uint
		config.DB.Model(&models.Attendance{}).Where("student_id = ?", sid).
			Order("date ASC").Pluck("id", &excess)
		if len(excess) > 80 {
			config.DB.Where("id IN ?", excess[:len(excess)-80]).Delete(&models.Attendance{})
		}
	}

	// Global cap: attendance
	var attTotal int64
	config.DB.Model(&models.Attendance{}).Count(&attTotal)
	if attTotal > 800 {
		var dropIDs []uint
		config.DB.Model(&models.Attendance{}).Order("date ASC").Limit(int(attTotal - 800)).Pluck("id", &dropIDs)
		if len(dropIDs) > 0 {
			config.DB.Where("id IN ?", dropIDs).Delete(&models.Attendance{})
		}
	}

	// Remove soft-deleted rows
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Subject{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Class{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Student{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Teacher{})

	// Cap notifications
	var notifIDs []uint
	config.DB.Model(&models.Notification{}).Order("id DESC").Pluck("id", &notifIDs)
	if len(notifIDs) > 5 {
		extra := notifIDs[5:]
		config.DB.Where("notification_id IN ?", extra).Delete(&models.NotificationReceipt{})
		config.DB.Where("id IN ?", extra).Delete(&models.Notification{})
	}

	log.Println("Trim finished")
	logTableCounts("after trim")
}

func logTableCounts(label string) {
	var s, t, p, sub, att, g int64
	config.DB.Model(&models.Student{}).Count(&s)
	config.DB.Model(&models.Teacher{}).Count(&t)
	config.DB.Model(&models.User{}).Where("role = ?", models.RoleParent).Count(&p)
	config.DB.Model(&models.Subject{}).Count(&sub)
	config.DB.Model(&models.Attendance{}).Count(&att)
	config.DB.Model(&models.Grade{}).Count(&g)
	log.Printf("[%s] students=%d teachers=%d parents=%d subjects=%d attendance=%d grades=%d",
		label, s, t, p, sub, att, g)
}