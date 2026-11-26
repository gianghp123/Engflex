// Package responses holds the access module's wire DTOs (tygo inputs).
package responses

import (
	"time"

	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
)

// AIReviewOutput is the cached paid AI result wire shape (spec §6.4). It
// mirrors models.AIOutput minus the owner column: userId never crosses the
// wire (owner-only reads), so one utils.Map call copies the rest whole and
// the payload crosses as the stored JSON document.
type AIReviewOutput struct {
	ID             string             `json:"id"`
	SubjectType    enums.UsageRefType `json:"subjectType"`
	SubjectID      string             `json:"subjectId"`
	Kind           enums.AIOutputKind `json:"kind"`
	SchemaVersion  int                `json:"schemaVersion"`
	Payload        datatypes.JSON     `json:"payload" swaggertype:"object"`
	CreditsCharged int                `json:"creditsCharged"`
	CreatedAt      time.Time          `json:"createdAt"`
}
