package enums

// ExamType identifies the examination format of a category (exam_categories.type).
type ExamType string

const (
	ExamTypeTOEIC ExamType = "toeic"
	ExamTypeIELTS ExamType = "ielts"
)
