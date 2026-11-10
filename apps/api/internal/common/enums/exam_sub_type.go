package enums

// ExamSubType identifies the examination track or module (exam_categories.sub_type, exam_tests.sub_type).
type ExamSubType string

const (
	ExamSubTypeAcademic        ExamSubType = "academic"
	ExamSubTypeGeneralTraining ExamSubType = "general_training"
)
