//go:build ignore

package main

// Usage:
//   go run cmd/seed/main.go          — trim excess data, then seed ~10 per category
//   go run cmd/seed/main.go -trim   — trim only (no re-seed)

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"sms-backend/config"
	"sms-backend/models"
)

const sampleSize = 10

func main() {
	trimOnly := flag.Bool("trim", false, "only remove excess data, do not seed")
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

	log.Println("Seeding sample data (~10 per category)...")
	seedSampleData()
}

func seedSampleData() {
	hashPwd := func(pwd string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pwd), 12)
		return string(h)
	}
	academicYear := 2025

	admins := []models.User{
		{Name: "Dawit Bekele", Email: "admin@school.et", Password: hashPwd("Admin@1234"), Role: models.RoleAdmin, Phone: "0911234567", IsActive: true},
		{Name: "Selam Haile", Email: "selam@school.et", Password: hashPwd("Admin@1234"), Role: models.RoleAdmin, Phone: "0922345678", IsActive: true},
	}
	for i := range admins {
		config.DB.Where("email = ?", admins[i].Email).FirstOrCreate(&admins[i])
	}

	teacherNames := []string{"Abebe Girma", "Tigist Alemu", "Yonas Tadesse", "Hiwot Mengistu", "Biruk Kebede",
		"Meron Tesfaye", "Samuel Hailu", "Rahel Worku", "Kebede Assefa", "Hanna Desta"}
	teacherPhones := []string{
		"0911100001", "0911100002", "0911100003", "0911100004", "0911100005",
		"0911100006", "0911100007", "0911100008", "0911100009", "0911100010",
	}
	departments := []string{"Mathematics", "Physics", "Chemistry", "Biology", "English", "Amharic", "Social Studies", "Civics", "ICT", "Mathematics"}
	qualifications := []string{"BEd Mathematics", "MSc Physics", "BEd Chemistry", "BSc Biology", "MEd English", "BA Amharic", "BEd Social Studies", "MA Civics", "BSc ICT", "MEd Mathematics"}
	var teachers []models.Teacher
	for i, name := range teacherNames {
		email := fmt.Sprintf("teacher%d@school.et", i+1)
		user := models.User{Name: name, Email: email, Password: hashPwd("Teacher@1234"), Role: models.RoleTeacher, Phone: teacherPhones[i], IsActive: true}
		config.DB.Where("email = ?", email).FirstOrCreate(&user)
		t := models.Teacher{UserID: user.ID, TeacherCode: fmt.Sprintf("TCH-%03d", i+1), Qualification: qualifications[i%len(qualifications)], Department: departments[i%len(departments)], JoinedAt: time.Now()}
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&t)
		t.User = user
		teachers = append(teachers, t)
	}

	// Remove any legacy classes where grade 9 or 10 was incorrectly given a stream.
	// In the Ethiopian curriculum grades 9-10 have a common curriculum — no stream.
	config.DB.Unscoped().Where("grade_level IN (9, 10) AND stream != ''").Delete(&models.Class{})

	classDefs := []struct {
		Grade           int
		Section, Stream string
	}{
		// Grades 9 & 10: common curriculum — no stream
		{9, "A", ""}, {9, "B", ""},
		{10, "A", ""}, {10, "B", ""},
		// Grades 11 & 12: stream-based
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
		classes = append(classes, c)
	}

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
	var subjectCodes []string
	for _, sub := range subjects {
		subjectCodes = append(subjectCodes, sub.Code)
	}

	// Remove legacy subject rows that used stream-specific Grade 9/10 codes or
	// incomplete sample-only curriculum codes.
	var legacySubjectIDs []uint
	config.DB.Model(&models.Subject{}).Where("code NOT IN ?", subjectCodes).Pluck("id", &legacySubjectIDs)
	if len(legacySubjectIDs) > 0 {
		config.DB.Where("subject_id IN ?", legacySubjectIDs).Delete(&models.Enrollment{})
		config.DB.Where("subject_id IN ?", legacySubjectIDs).Delete(&models.Attendance{})
		config.DB.Where("subject_id IN ?", legacySubjectIDs).Delete(&models.Grade{})
		config.DB.Unscoped().Where("id IN ?", legacySubjectIDs).Delete(&models.Subject{})
	}

	var parents []models.User
	for i := 1; i <= sampleSize; i++ {
		email := fmt.Sprintf("parent%d@school.et", i)
		p := models.User{
			Name: fmt.Sprintf("Parent %d", i), Email: email,
			Password: hashPwd("Parent@1234"), Role: models.RoleParent,
			Phone: fmt.Sprintf("0944%06d", 100000+i), IsActive: true,
		}
		config.DB.Where("email = ?", email).FirstOrCreate(&p)
		config.DB.Model(&p).Updates(map[string]any{
			"name": fmt.Sprintf("Parent %d", i),
		})
		parents = append(parents, p)
	}

	studentNames := []string{
		"Abinet Tadesse", "Bemnet Girma", "Cherenet Haile", "Daniel Bekele", "Eyerusalem Alemu",
		"Fikir Mengistu", "Gelila Worku", "Henok Tesfaye", "Iman Kebede", "Jonas Hailu",
	}
	var students []models.Student
	for i, name := range studentNames {
		assignedClass := classes[i%len(classes)]
		classID := assignedClass.ID
		grade := assignedClass.GradeLevel
		stream := assignedClass.Stream

		email := fmt.Sprintf("student%d@school.et", i+1)
		user := models.User{Name: name, Email: email, Password: hashPwd("Student@1234"), Role: models.RoleStudent, IsActive: true, Phone: fmt.Sprintf("0911500%03d", i+1)}
		config.DB.Where("email = ?", email).FirstOrCreate(&user)
		parentIdx := i
		if i == 1 {
			parentIdx = 0 // Student 2 also gets Parent 1
		} else if i == 3 {
			parentIdx = 2 // Student 4 also gets Parent 3
		}

		st := models.Student{
			UserID: user.ID, ParentID: parents[parentIdx].ID, ClassID: &classID,
			StudentCode: fmt.Sprintf("STU-%d-%03d", academicYear, i+1),
			ParentName:  parents[parentIdx].Name, ParentEmail: parents[parentIdx].Email,
			ParentPhone: parents[parentIdx].Phone,
			DateOfBirth: time.Now().AddDate(-15, 0, 0),
			Stream:      stream, GradeLevel: grade, PromotionStatus: models.PromotionNormal,
			AcademicYear: academicYear, EnrolledAt: time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC),
		}
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&st)
		config.DB.Model(&st).Updates(map[string]any{
			"class_id":         classID,
			"grade_level":      grade,
			"stream":           stream,
			"parent_id":        parents[parentIdx].ID,
			"parent_name":      parents[parentIdx].Name,
			"parent_email":     parents[parentIdx].Email,
			"parent_phone":     parents[parentIdx].Phone,
			"academic_year":    academicYear,
			"enrolled_at":      time.Date(academicYear, 9, 1, 0, 0, 0, 0, time.UTC),
			"promotion_status": models.PromotionNormal,
		})
		config.DB.First(&st, st.ID)
		students = append(students, st)
	}

	for _, st := range students {
		config.DB.Where("student_id = ?", st.ID).Delete(&models.Enrollment{})
		for _, sub := range subjects {
			if sub.GradeLevel != st.GradeLevel {
				continue
			}
			if sub.Stream != "" && sub.Stream != st.Stream {
				continue
			}
			var ec int64
			config.DB.Model(&models.Enrollment{}).Where("student_id = ? AND subject_id = ?", st.ID, sub.ID).Count(&ec)
			if ec == 0 {
				config.DB.Create(&models.Enrollment{StudentID: st.ID, SubjectID: sub.ID})
			}
		}
	}

	// Attendance: seed 30 school days (~6 weeks) with realistic rates
	// ~80% Present, ~13% Late, ~7% Absent — varied per student using ID as offset
	attStatuses := []string{
		"Present", "Present", "Present", "Present", "Present",
		"Present", "Present", "Present", "Present", "Present",
		"Present", "Present", "Late", "Late",
		"Absent",
	}
	attStart := time.Now().AddDate(0, 0, -42) // go back 42 calendar days to collect 30 school days
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
					score := 72.0 + float64((st.ID+sub.ID)%25)
					if i == 0 {
						// Abinet Tadesse (Student 1) - Repeat (fail all G9 subjects)
						score = 40.0
					} else if i == 1 {
						// Bemnet Girma (Student 2) - Conditional (fail 1 G9 subject, pass others)
						if len(subs) > 0 && sub.ID == subs[0].ID {
							score = 42.0
						}
					} else if i == 3 && gt == "Final" {
						score = 42.0
					}
					var gc int64
					config.DB.Model(&models.Grade{}).Where("student_id = ? AND subject_id = ? AND type = ? AND semester = ? AND academic_year = ?",
						st.ID, sub.ID, gt, sem, academicYear).Count(&gc)
					if gc == 0 {
						var teacherID uint
						if sub.TeacherID != nil {
							teacherID = *sub.TeacherID
						}
						config.DB.Create(&models.Grade{
							StudentID: st.ID, SubjectID: sub.ID, TeacherID: teacherID,
							Score: score, MaxScore: 100, Type: gt, Semester: sem, AcademicYear: academicYear,
						})
					}
				}
			}
		}
	}

	for i, n := range []struct{ T, B string }{
		{"Welcome 2025", "New academic year for Grades 9–12."},
		{"Parent Meeting", "Friday 2pm — all parents invited."},
		{"Fee Deadline", "Tuition due end of month."},
		{"Exam Prep", "Grade 12 national exam preparation week."},
		{"Sports Day", "Inter-class sports on Meskerem 20."},
	} {
		var cnt int64
		config.DB.Model(&models.Notification{}).Where("title = ?", n.T).Count(&cnt)
		if cnt == 0 {
			notif := models.Notification{Title: n.T, Body: n.B, TargetRoles: "Student,Parent", SenderID: admins[0].ID}
			config.DB.Create(&notif)
			for _, st := range students {
				var rc int64
				config.DB.Model(&models.NotificationReceipt{}).Where("notification_id = ? AND user_id = ?", notif.ID, st.UserID).Count(&rc)
				if rc == 0 {
					config.DB.Create(&models.NotificationReceipt{NotificationID: notif.ID, UserID: st.UserID, IsRead: i%2 == 0})
				}
			}
		}
	}

	for i, student := range students {
		// Semester 1 payment
		rid := fmt.Sprintf("ETH-CBE-%06d", 300000+i)
		var tc int64
		config.DB.Model(&models.Transaction{}).Where("receipt_id = ?", rid).Count(&tc)
		if tc == 0 {
			txStatus := "Pending"
			if i%3 == 0 {
				txStatus = "Verified"
			}
			config.DB.Create(&models.Transaction{
				StudentID: student.ID, Amount: 8500, ReceiptID: rid, Type: "Tuition",
				Status: txStatus, Description: "Semester 1", CreatedBy: student.UserID,
				AcademicYear: academicYear, Semester: "Semester 1",
			})
		}

		// Semester 2 payment (seeded for a subset to demonstrate overdue states)
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
					Status: txStatus, Description: "Semester 2", CreatedBy: student.UserID,
					AcademicYear: academicYear, Semester: "Semester 2",
				})
			}
		}
	}

	for _, t := range teachers {
		for m := 1; m <= 6; m++ {
			var pc int64
			config.DB.Model(&models.Payroll{}).Where("teacher_id = ? AND month = ? AND year = ?", t.ID, m, academicYear).Count(&pc)
			if pc == 0 {
				paidAt := time.Now().AddDate(0, -(6 - m), 0)
				status := "Paid"
				if m >= 5 {
					status = "Pending"
					paidAt = time.Time{}
				}
				p := models.Payroll{TeacherID: t.ID, Amount: 12000 + float64(m*200), Month: m, Year: academicYear, Status: status}
				if status == "Paid" {
					p.PaidAt = &paidAt
				}
				config.DB.Create(&p)
			}
		}
	}

	for _, student := range students {
		var lfc int64
		config.DB.Model(&models.LockerFile{}).Where("student_id = ?", student.ID).Count(&lfc)
		if lfc == 0 {
			config.DB.Create(&models.LockerFile{
				StudentID:  student.ID,
				FileName:   "Grade_" + fmt.Sprintf("%d", student.GradeLevel) + "_Portfolio.pdf",
				FilePath:   "./uploads/locker/portfolio_" + fmt.Sprintf("%d", student.ID) + ".pdf",
				FileSize:   102400,
				FileType:   "pdf",
				Category:   "Portfolio",
				IsPublic:   true,
				UploadedAt: time.Now(),
			})
			config.DB.Create(&models.LockerFile{
				StudentID:  student.ID,
				FileName:   "Community_Service_Certificate.jpg",
				FilePath:   "./uploads/locker/cert_" + fmt.Sprintf("%d", student.ID) + ".jpg",
				FileSize:   204800,
				FileType:   "jpg",
				Category:   "Certificate",
				IsPublic:   false,
				UploadedAt: time.Now(),
			})
		}
	}

	log.Println("=== Seed complete (~10 per category) ===")
	log.Println("admin@school.et / Admin@1234")
	log.Println("teacher1@school.et / Teacher@1234")
	log.Println("student1@school.et / Student@1234")
	log.Println("parent1@school.et / Parent@1234")
}

func trimExcessData() {
	log.Println("Trimming database (removing bulk / legacy seed rows)...")
	logTableCounts("before trim")

	// Hard-remove legacy grade 9/10 classes that were incorrectly given a stream.
	// The Ethiopian curriculum has a common curriculum for grades 9-10 — no stream.
	var legacyClassIDs []uint
	config.DB.Unscoped().Model(&models.Class{}).
		Where("grade_level IN (9, 10) AND stream != ''").Pluck("id", &legacyClassIDs)
	if len(legacyClassIDs) > 0 {
		log.Printf("Removing %d legacy grade 9/10 stream classes", len(legacyClassIDs))
		config.DB.Where("class_id IN ?", legacyClassIDs).Delete(&models.Student{})
		config.DB.Unscoped().Where("id IN ?", legacyClassIDs).Delete(&models.Class{})
	}

	var fallbackTeacher uint
	if config.DB.Model(&models.Teacher{}).Order("id").Limit(1).Pluck("id", &fallbackTeacher); fallbackTeacher > 0 {
		config.DB.Exec("UPDATE classes SET teacher_id = ? WHERE teacher_id NOT IN (SELECT id FROM teachers)", fallbackTeacher)
		config.DB.Exec("UPDATE subjects SET teacher_id = ? WHERE teacher_id NOT IN (SELECT id FROM teachers)", fallbackTeacher)
	}

	keepStudentEmails := emails("student", sampleSize, "@school.et")
	keepTeacherEmails := emails("teacher", sampleSize, "@school.et")
	keepParentEmails := emails("parent", sampleSize, "@school.et")
	keepAdminEmails := []string{"admin@school.et", "selam@school.et"}
	var keepCodes []string
	for _, cur := range models.EthiopianGrades9to12Subjects() {
		for _, grade := range cur.Grades {
			if grade <= 10 && cur.Stream != "" {
				continue
			}
			keepCodes = append(keepCodes, models.CurriculumSubjectCode(cur.Code, cur.Stream, grade))
		}
	}

	// Purge huge legacy attendance/grade tables (old 50-student seed)
	var attTotal, gradeTotal int64
	config.DB.Model(&models.Attendance{}).Count(&attTotal)
	config.DB.Model(&models.Grade{}).Count(&gradeTotal)
	if attTotal > 500 {
		log.Printf("Deleting %d attendance rows (legacy bulk)", attTotal)
		config.DB.Exec("DELETE FROM attendances")
	}
	if gradeTotal > 800 {
		log.Printf("Deleting %d grade rows (legacy bulk)", gradeTotal)
		config.DB.Exec("DELETE FROM grades")
	}

	var removeStudentUserIDs []uint
	config.DB.Unscoped().Model(&models.User{}).Where("role = ? AND email NOT IN ?", models.RoleStudent, keepStudentEmails).
		Pluck("id", &removeStudentUserIDs)
	if len(removeStudentUserIDs) > 0 {
		var sids []uint
		config.DB.Unscoped().Model(&models.Student{}).Where("user_id IN ?", removeStudentUserIDs).Pluck("id", &sids)
		deleteForStudents(sids)
		config.DB.Unscoped().Where("user_id IN ?", removeStudentUserIDs).Delete(&models.Student{})
		config.DB.Unscoped().Where("id IN ?", removeStudentUserIDs).Delete(&models.User{})
	}

	var removeTeacherUserIDs []uint
	config.DB.Unscoped().Model(&models.User{}).Where("role = ? AND email NOT IN ?", models.RoleTeacher, keepTeacherEmails).
		Pluck("id", &removeTeacherUserIDs)
	if len(removeTeacherUserIDs) > 0 {
		var tids []uint
		config.DB.Unscoped().Model(&models.Teacher{}).Where("user_id IN ?", removeTeacherUserIDs).Pluck("id", &tids)
		var keepTid uint
		config.DB.Model(&models.Teacher{}).Joins("JOIN users ON users.id = teachers.user_id").
			Where("users.email IN ?", keepTeacherEmails).Order("teachers.id").Limit(1).Pluck("teachers.id", &keepTid)
		if keepTid > 0 {
			if len(tids) > 0 {
				config.DB.Model(&models.Class{}).Where("teacher_id IN ?", tids).Update("teacher_id", keepTid)
				config.DB.Model(&models.Subject{}).Where("teacher_id IN ?", tids).Update("teacher_id", keepTid)
			}
			// Legacy rows may reference user IDs in teacher_id column
			config.DB.Model(&models.Class{}).Where("teacher_id NOT IN (SELECT id FROM teachers)").Update("teacher_id", keepTid)
			config.DB.Model(&models.Subject{}).Where("teacher_id NOT IN (SELECT id FROM teachers)").Update("teacher_id", keepTid)
		}
		if len(tids) > 0 {
			config.DB.Where("teacher_id IN ?", tids).Delete(&models.Payroll{})
			config.DB.Unscoped().Where("id IN ?", tids).Delete(&models.Teacher{})
		}
		config.DB.Unscoped().Where("id IN ?", removeTeacherUserIDs).Delete(&models.User{})
	}

	var removeParentIDs []uint
	config.DB.Unscoped().Model(&models.User{}).Where("role = ? AND email NOT IN ?", models.RoleParent, keepParentEmails).
		Pluck("id", &removeParentIDs)
	if len(removeParentIDs) > 0 {
		config.DB.Unscoped().Where("id IN ?", removeParentIDs).Delete(&models.User{})
	}

	// Legacy admin accounts beyond the two defaults
	var removeAdminIDs []uint
	config.DB.Unscoped().Model(&models.User{}).Where("role = ? AND email NOT IN ?", models.RoleAdmin, keepAdminEmails).
		Pluck("id", &removeAdminIDs)
	if len(removeAdminIDs) > 0 {
		config.DB.Unscoped().Where("id IN ?", removeAdminIDs).Delete(&models.User{})
	}

	var removeSubIDs []uint
	config.DB.Model(&models.Subject{}).Where("code NOT IN ?", keepCodes).Pluck("id", &removeSubIDs)
	if len(removeSubIDs) > 0 {
		config.DB.Where("subject_id IN ?", removeSubIDs).Delete(&models.Enrollment{})
		config.DB.Where("subject_id IN ?", removeSubIDs).Delete(&models.Attendance{})
		config.DB.Where("subject_id IN ?", removeSubIDs).Delete(&models.Grade{})
		config.DB.Unscoped().Where("id IN ?", removeSubIDs).Delete(&models.Subject{})
	}

	// Orphan rows (student already deleted but attendance/grades remain)
	config.DB.Exec(`
		DELETE FROM attendances WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM grades WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM enrollments WHERE student_id NOT IN (SELECT id FROM students);
		DELETE FROM transactions WHERE student_id NOT IN (SELECT id FROM students);
	`)

	var keptStudentIDs []uint
	config.DB.Model(&models.Student{}).Pluck("id", &keptStudentIDs)
	if len(keptStudentIDs) > 0 {
		// Cap attendance: keep at most 80 rows per student (delete oldest excess)
		for _, sid := range keptStudentIDs {
			var excess []uint
			config.DB.Model(&models.Attendance{}).Where("student_id = ?", sid).
				Order("date ASC").Pluck("id", &excess)
			if len(excess) > 80 {
				config.DB.Where("id IN ?", excess[:len(excess)-80]).Delete(&models.Attendance{})
			}
		}
	}

	// Global cap: if attendance table is still huge, delete oldest rows
	config.DB.Model(&models.Attendance{}).Count(&attTotal)
	if attTotal > 800 {
		var dropIDs []uint
		config.DB.Model(&models.Attendance{}).Order("date ASC").Limit(int(attTotal-800)).Pluck("id", &dropIDs)
		if len(dropIDs) > 0 {
			config.DB.Where("id IN ?", dropIDs).Delete(&models.Attendance{})
		}
	}

	// Remove soft-deleted rows (old archives)
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Subject{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Class{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Student{})
	config.DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Teacher{})

	// Drop empty legacy classes (e.g. old "Grade 9 Natural" duplicates beyond sample)
	var classIDs []uint
	config.DB.Model(&models.Class{}).Pluck("id", &classIDs)
	if len(classIDs) > sampleSize+4 {
		for _, cid := range classIDs[sampleSize+4:] {
			var cnt int64
			config.DB.Model(&models.Student{}).Where("class_id = ?", cid).Count(&cnt)
			if cnt == 0 {
				config.DB.Unscoped().Delete(&models.Class{}, cid)
			}
		}
	}

	var notifIDs []uint
	config.DB.Model(&models.Notification{}).Order("id DESC").Pluck("id", &notifIDs)
	if len(notifIDs) > 5 {
		extra := notifIDs[5:]
		config.DB.Where("notification_id IN ?", extra).Delete(&models.NotificationReceipt{})
		config.DB.Where("id IN ?", extra).Delete(&models.Notification{})
	}

	log.Println("Trim finished — run with -trim to clean only, or full seed to repopulate sample data")
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

func emails(prefix string, n int, suffix string) []string {
	out := make([]string, n)
	for i := 1; i <= n; i++ {
		out[i-1] = fmt.Sprintf("%s%d%s", prefix, i, suffix)
	}
	return out
}

func deleteForStudents(studentIDs []uint) {
	if len(studentIDs) == 0 {
		return
	}
	config.DB.Where("student_id IN ?", studentIDs).Delete(&models.Enrollment{})
	config.DB.Where("student_id IN ?", studentIDs).Delete(&models.Attendance{})
	config.DB.Where("student_id IN ?", studentIDs).Delete(&models.Grade{})
	config.DB.Where("student_id IN ?", studentIDs).Delete(&models.Transaction{})
}
