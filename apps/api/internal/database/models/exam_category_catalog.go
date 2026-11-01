package models

// ExamCategoryCatalog is the catalog read model. Its count is a projection,
// never a column on ExamCategory (which is also used in joined relations).
type ExamCategoryCatalog struct {
	ExamCategory `gorm:"embedded"`
	TestsCount   int64
}
