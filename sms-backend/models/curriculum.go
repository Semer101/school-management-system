package models

// Ethiopian high school streams (Grades 9–12)
const (
	StreamNatural = "Natural Science"
	StreamSocial  = "Social Science"
)

// Promotion outcomes after year-end assessment
const (
	PromotionNormal      = "normal"
	PromotionConditional = "conditional"
	PromotionRepeat      = "repeat"
)

// CurriculumSubject maps a subject code to grade levels and stream
type CurriculumSubject struct {
	Code       string
	Name       string
	Grades     []int // 9, 10, 11, 12
	Stream     string // empty = common to both streams
	TeacherIdx int    // index into seed teachers slice
}

// EthiopianGrades9to12Subjects returns fixed subjects per stream and grade
func EthiopianGrades9to12Subjects() []CurriculumSubject {
	common := []CurriculumSubject{
		{Code: "AMH", Name: "Amharic", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 2},
		{Code: "ENG", Name: "English", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 2},
		{Code: "MATH", Name: "Mathematics", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 0},
		{Code: "CIV", Name: "Civics & Ethical Education", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 5},
		{Code: "ICT", Name: "ICT", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 6},
		{Code: "PE", Name: "Physical Education", Grades: []int{9, 10, 11, 12}, Stream: "", TeacherIdx: 7},
	}
	natural := []CurriculumSubject{
		{Code: "PHY", Name: "Physics", Grades: []int{9, 10, 11, 12}, Stream: StreamNatural, TeacherIdx: 1},
		{Code: "CHEM", Name: "Chemistry", Grades: []int{9, 10, 11, 12}, Stream: StreamNatural, TeacherIdx: 3},
		{Code: "BIO", Name: "Biology", Grades: []int{9, 10, 11, 12}, Stream: StreamNatural, TeacherIdx: 4},
	}
	social := []CurriculumSubject{
		{Code: "GEO", Name: "Geography", Grades: []int{9, 10, 11, 12}, Stream: StreamSocial, TeacherIdx: 5},
		{Code: "HIST", Name: "History", Grades: []int{9, 10, 11, 12}, Stream: StreamSocial, TeacherIdx: 5},
		{Code: "ECON", Name: "Economics", Grades: []int{10, 11, 12}, Stream: StreamSocial, TeacherIdx: 5},
	}
	out := append([]CurriculumSubject{}, common...)
	out = append(out, natural...)
	out = append(out, social...)
	return out
}

// SubjectCodesForStreamGrade returns subject codes for a student's stream and grade
func SubjectCodesForStreamGrade(stream string, grade int) []string {
	var codes []string
	for _, s := range EthiopianGrades9to12Subjects() {
		if s.Stream != "" && s.Stream != stream {
			continue
		}
		for _, g := range s.Grades {
			if g == grade {
				codes = append(codes, s.Code)
				break
			}
		}
	}
	return codes
}
