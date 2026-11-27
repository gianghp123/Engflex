package requests

import "engflex-api/internal/common/enums"

type CreateCategory struct {
	Slug    string             `json:"slug" binding:"required"`
	Title   string             `json:"title" binding:"required"`
	Type    enums.ExamType     `json:"type" binding:"required,oneof=toeic ielts"`
	SubType *enums.ExamSubType `json:"subType,omitempty" binding:"omitempty,oneof=academic general_training"`
}
