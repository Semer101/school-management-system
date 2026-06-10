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
		"Wubishet", "Yohanis", "Zewde", "Aron", "Betekle",
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
		"Yeshi", "Zeneba", "Amarech", "Belay", "Chaltu",
	}
	lastNames = []string{
		"Bekele", "Girma", "Haile", "Alemu", "Tadesse",
		"Mengistu", "Worku", "Tesfaye", "Assefa", "Desta",
		"Kebede", "Hailu", "Gebremedhin", "Berhanu", "Fekadu",
		"Wondemu", "Abebe", "Lemma", "Tekle", "Aragaw",
		"Shiferaw", "Mulugeta", "Negash", "Asfaw", "Gebre",
		"Kassa", "Tamiru", "Wondalessie", "Gidey", "Beyene",
		"Ayalew", "Habte", "Yilma", "Dejene", "Tilahun",
		"Workineh", "Goshu", "Bizuayehu", "Abate", "Kibru",
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

// upsertUser creates or updates a user by email.
// Values must be passed explicitly — FirstOrCreate overwrites the struct with DB state.
func upsertUser(email, name, phone, password, role string) models.User {
	u := models.User{
		Name: name, Email: email, Password: hashPwd(password),
		Role: role, Phone: phone, IsActive: true,
	}
	config.DB.Where("email = ?", email).FirstOrCreate(&u)
	config.DB.Model(&u).Updates(map[string]any{
		"name": name, "phone": phone, "is_active": true,
	})
	config.DB.Where("email = ?", email).First(&u)
	return u
}

// pickName returns a stable name pair for a given index (idempotent across re-seeds).
func pickName(index int, female bool) (string, string) {
	firstNames := maleFirstNames
	if female {
		firstNames = femaleFirstNames
	}
	first := firstNames[index%len(firstNames)]
	last := lastNames[(index*7+3)%len(lastNames)]
	return first, last
}

func cleanupOrphanClasses(year int) {
	expectedClasses := expectedSeedClassNames(year)
	var orphanClassIDs []uint
	config.DB.Model(&models.Class{}).
		Where("year = ? AND name NOT IN ?", year, expectedClasses).
		Pluck("id", &orphanClassIDs)
	for _, cid := range orphanClassIDs {
		var studentCount int64
		config.DB.Model(&models.Student{}).Where("class_id = ?", cid).Count(&studentCount)
		if studentCount == 0 {
			config.DB.Unscoped().Delete(&models.Class{}, cid)
		}
	}

	// Legacy naming from older seeds (e.g. "Grade 9A")
	var legacyNamedIDs []uint
	config.DB.Model(&models.Class{}).Where("name LIKE ?", "Grade %").Pluck("id", &legacyNamedIDs)
	for _, cid := range legacyNamedIDs {
		var studentCount int64
		config.DB.Model(&models.Student{}).Where("class_id = ?", cid).Count(&studentCount)
		if studentCount == 0 {
			config.DB.Unscoped().Delete(&models.Class{}, cid)
		}
	}
}

func expectedSeedClassNames(year int) []string {
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
	names := make([]string, 0, len(classDefs))
	for _, cd := range classDefs {
		if cd.Stream != "" {
			names = append(names, fmt.Sprintf("%d%s %s", cd.Grade, cd.Section, cd.Stream))
		} else {
			names = append(names, fmt.Sprintf("%d%s", cd.Grade, cd.Section))
		}
	}
	_ = year // reserved for future multi-year seeds
	return names
}

func seedProductionData(studentCount int) {
	academicYear := 2025
	seedRng := rand.New(rand.NewSource(42))

	log.Println("Resetting previous seed records (grades, enrollments, transactions, attendances)...")
	config.DB.Exec("DELETE FROM grades WHERE academic_year = ?", academicYear)
	config.DB.Exec("DELETE FROM enrollments")
	config.DB.Exec("DELETE FROM transactions WHERE academic_year = ?", academicYear)
	config.DB.Exec("DELETE FROM attendances")

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
		admins = append(admins, upsertUser(d.Email, d.Name, d.Phone, "Admin@1234", models.RoleAdmin))
	}

	// ── 2. Teachers ───────────────────────────────────────
	var teachers []models.Teacher
	for i := 1; i <= sampleTeachers; i++ {
		firstName, lastName := pickName(i, i%2 == 0)
		name := fmt.Sprintf("%s %s", firstName, lastName)
		email := fmt.Sprintf("teacher%d@school.et", i)
		phone := fmt.Sprintf("0911100%03d", i)
		dept := departments[(i-1)%len(departments)]
		qual := qualifications[(i-1)%len(qualifications)]
		teacherCode := fmt.Sprintf("TCH-%04d", i)
		joinedAt := time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC)

		u := upsertUser(email, name, phone, "Teacher@1234", models.RoleTeacher)

		t := models.Teacher{
			UserID: u.ID, TeacherCode: teacherCode,
			Qualification: qual, Department: dept, JoinedAt: joinedAt,
		}
		config.DB.Where("user_id = ?", u.ID).FirstOrCreate(&t)
		config.DB.Model(&t).Updates(map[string]any{
			"teacher_code": teacherCode, "qualification": qual,
			"department": dept, "joined_at": joinedAt,
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
		firstName, lastName := pickName(i+100, i%2 == 0)
		name := fmt.Sprintf("%s %s", firstName, lastName)
		email := fmt.Sprintf("parent%d@school.et", i)
		phone := fmt.Sprintf("0944%06d", 100000+i)
		parents = append(parents, upsertUser(email, name, phone, "Parent@1234", models.RoleParent))
	}

	// ── 6. Students ───────────────────────────────────────
	var students []models.Student
	enrolledAt := time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= studentCount; i++ {
		isMale := i%2 == 0
		firstName, lastName := pickName(i+200, !isMale)
		name := fmt.Sprintf("%s %s", firstName, lastName)
		email := fmt.Sprintf("student%d@school.et", i)
		phone := fmt.Sprintf("0911500%03d", i)
		studentCode := fmt.Sprintf("STU-%d-%04d", academicYear, i)

		// Assign to a class (cycle through classes, with stable variety per index)
		classIdx := (i - 1) % len(classes)
		if i%3 == 0 {
			classIdx = (i * 7) % len(classes)
		}
		assignedClass := classes[classIdx]

		parentIdx := (i - 1) % len(parents)
		if i%7 == 0 {
			parentIdx = (i - 2) % len(parents)
			if parentIdx < 0 {
				parentIdx = 0
			}
		}

		u := upsertUser(email, name, phone, "Student@1234", models.RoleStudent)

		dobYear := academicYear - assignedClass.GradeLevel - 3
		dob := time.Date(dobYear, time.Month((i%12)+1), (i%28)+1, 0, 0, 0, 0, time.UTC)

		st := models.Student{
			UserID: u.ID, ParentID: parents[parentIdx].ID, ClassID: &assignedClass.ID,
			StudentCode: studentCode,
			ParentName:  parents[parentIdx].Name, ParentEmail: parents[parentIdx].Email,
			ParentPhone: parents[parentIdx].Phone,
			DateOfBirth: dob, Stream: assignedClass.Stream, GradeLevel: assignedClass.GradeLevel,
			PromotionStatus: models.PromotionNormal, AcademicYear: academicYear, EnrolledAt: enrolledAt,
		}
		config.DB.Where("user_id = ?", u.ID).FirstOrCreate(&st)
		config.DB.Model(&st).Updates(map[string]any{
			"student_code":     studentCode,
			"class_id":         assignedClass.ID,
			"grade_level":      assignedClass.GradeLevel,
			"stream":           assignedClass.Stream,
			"parent_id":        parents[parentIdx].ID,
			"parent_name":      parents[parentIdx].Name,
			"parent_email":     parents[parentIdx].Email,
			"parent_phone":     parents[parentIdx].Phone,
			"date_of_birth":    dob,
			"academic_year":    academicYear,
			"enrolled_at":      enrolledAt,
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
	var allAttendance []models.Attendance
	for _, st := range students {
		schoolDays := 0
		for day := 0; schoolDays < 30; day++ {
			d := attStart.AddDate(0, 0, day)
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			schoolDays++
			pick := (int(st.ID) + schoolDays) % len(attStatuses)
			allAttendance = append(allAttendance, models.Attendance{
				StudentID: st.ID, SubjectID: nil, Date: d, Status: attStatuses[pick],
			})
		}
	}
	if len(allAttendance) > 0 {
		log.Printf("Inserting %d attendance records in bulk...", len(allAttendance))
		config.DB.CreateInBatches(&allAttendance, 200)
	}

	// ── 9. Grades (Midterm + Final for each enrolled subject/semester) ─
	var allGrades []models.Grade
	for i, st := range students {
		var subs []models.Subject
		config.DB.Joins("JOIN enrollments ON enrollments.subject_id = subjects.id").
			Where("enrollments.student_id = ?", st.ID).Find(&subs)
		for _, sub := range subs {
			semsToSeed := []string{"Semester 1", "Semester 2", "Semester 3"}
			for _, sem := range semsToSeed {
				for _, gt := range []string{"Midterm", "Final"} {
					// Realistic grade distribution: bell curve around 72
					score := 72.0 + float64((st.ID+sub.ID)%25)
					// Add some randomness
					score += float64(seedRng.Intn(11) - 5) // ±5
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
					var teacherID uint
					if sub.TeacherID != nil {
						teacherID = *sub.TeacherID
					}
					allGrades = append(allGrades, models.Grade{
						StudentID: st.ID, SubjectID: sub.ID, TeacherID: teacherID,
						Score: score, MaxScore: 100, Type: gt, Semester: sem,
						AcademicYear: academicYear,
					})
				}
			}
		}
	}
	if len(allGrades) > 0 {
		log.Printf("Inserting %d grades in bulk...", len(allGrades))
		config.DB.CreateInBatches(&allGrades, 200)
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
	var allReceipts []models.NotificationReceipt
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
				allReceipts = append(allReceipts, models.NotificationReceipt{
					NotificationID: notif.ID, UserID: st.UserID, IsRead: i%2 == 0,
				})
			}
		}
	}
	if len(allReceipts) > 0 {
		log.Printf("Inserting %d notification receipts in bulk...", len(allReceipts))
		config.DB.CreateInBatches(&allReceipts, 200)
	}

	// ── 11. Finance (Transactions) ────────────────────────
	var allTransactions []models.Transaction
	for i, student := range students {
		// Semester 1 payment (everyone gets one)
		rid1 := fmt.Sprintf("ETH-CBE-%06d", 300000+i)
		txStatus1 := "Pending"
		if i%3 == 0 {
			txStatus1 = "Verified"
		}
		allTransactions = append(allTransactions, models.Transaction{
			StudentID: student.ID, Amount: 8500, ReceiptID: rid1, Type: "Tuition",
			Status: txStatus1, Description: "Semester 1 Tuition",
			CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 1",
		})

		// Semester 2 payment (subset to demonstrate overdue states)
		if i%2 == 0 {
			rid2 := fmt.Sprintf("ETH-CBE-%06d", 400000+i)
			txStatus2 := "Pending"
			if i%4 == 0 {
				txStatus2 = "Verified"
			}
			allTransactions = append(allTransactions, models.Transaction{
				StudentID: student.ID, Amount: 8500, ReceiptID: rid2, Type: "Tuition",
				Status: txStatus2, Description: "Semester 2 Tuition",
				CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 2",
			})
		}

		// Semester 3 payment (everyone gets one)
		rid3 := fmt.Sprintf("ETH-CBE-%06d", 500000+i)
		txStatus3 := "Pending"
		if i%3 == 1 {
			txStatus3 = "Verified"
		}
		allTransactions = append(allTransactions, models.Transaction{
			StudentID: student.ID, Amount: 8500, ReceiptID: rid3, Type: "Tuition",
			Status: txStatus3, Description: "Semester 3 Tuition",
			CreatedBy: student.UserID, AcademicYear: academicYear, Semester: "Semester 3",
		})
	}
	if len(allTransactions) > 0 {
		log.Printf("Inserting %d transactions in bulk...", len(allTransactions))
		config.DB.CreateInBatches(&allTransactions, 100)
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

	cleanupOrphanClasses(academicYear)

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

	cleanupOrphanClasses(2025)

	// Cap notifications (matches seed count)
	var notifIDs []uint
	config.DB.Model(&models.Notification{}).Order("id DESC").Pluck("id", &notifIDs)
	if len(notifIDs) > 8 {
		extra := notifIDs[8:]
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