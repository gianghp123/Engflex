package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// ListVocabulary sorts and filters the vocabulary studio. Pagination/search
// come from the embedded common.ListParams (tstype:",extends" -> `interface
// ListVocabulary extends ListParams`); json tags mirror the query keys so
// the generated TS field names match the wire.
type ListVocabulary struct {
	common.ListParams `tstype:",extends"`
	Sort              enums.VocabularySort   `form:"sort" json:"sort" binding:"omitempty,oneof=recent mastery alphabetical interval"`
	CEFR              enums.CEFR             `form:"cefr" json:"cefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Source            enums.VocabularySource `form:"source" json:"source" binding:"omitempty,oneof=lesson conversation video manual"`
}
