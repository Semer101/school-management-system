package helpers

import (
	"bytes"
	"fmt"
	"time"

	"sms-backend/models"

	"github.com/jung-kurt/gofpdf"
)

type CourseGrade struct {
	Title       string
	Score       float64
	LetterGrade string
}

func GenerateReportCardPDF(student models.Student, courses []CourseGrade) ([]byte, error) {

	// ── 1. Create the PDF document ────────────────────
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// ── 2. Header ─────────────────────────────────────
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(190, 15, "OFFICIAL DIGITAL REPORT CARD", "0", 1, "C", false, 0, "")
	pdf.Ln(5)

	// ── 3. Student info ───────────────────────────────
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Student Name:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(150, 8, student.User.Name) // direct field access
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Student ID:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(150, 8, student.StudentCode) // use StudentCode, more readable than numeric ID
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Class:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(150, 8, student.Class.Name) // requires Preload("Class") in controller
	pdf.Ln(15)

	// ── 4. Grades table header ─────────────────────────
	pdf.SetFillColor(200, 220, 255)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(100, 10, "Course Title", "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 10, "Score", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, "Grade", "1", 1, "C", true, 0, "")

	// ── 5. Grades table rows ───────────────────────────
	pdf.SetFont("Arial", "", 12)
	for i, course := range courses {
		// Alternate row background: white and light gray
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(245, 245, 245)
		}
		pdf.CellFormat(100, 10, course.Title, "1", 0, "L", true, 0, "")
		pdf.CellFormat(55, 10, fmt.Sprintf("%.1f", course.Score), "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 10, course.LetterGrade, "1", 1, "C", true, 0, "")
	}
	pdf.Ln(10)

	// ── 6. Footer ─────────────────────────────────────
	pdf.SetY(-30) // 30mm from bottom — works regardless of content length
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 10, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))

	// ── 7. Return as bytes — no file saved to disk ─────
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
