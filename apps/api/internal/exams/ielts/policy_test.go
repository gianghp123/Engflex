package ielts

import (
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
)

func TestCalculateReadingBand(t *testing.T) {
	tests := []struct {
		name       string
		rawCorrect int
		subType    enums.ExamSubType
		expected   float64
	}{
		// Academic Reading
		{"Academic 40 correct", 40, enums.ExamSubTypeAcademic, 9.0},
		{"Academic 39 correct", 39, enums.ExamSubTypeAcademic, 9.0},
		{"Academic 37 correct", 37, enums.ExamSubTypeAcademic, 8.5},
		{"Academic 35 correct", 35, enums.ExamSubTypeAcademic, 8.0},
		{"Academic 33 correct", 33, enums.ExamSubTypeAcademic, 7.5},
		{"Academic 30 correct", 30, enums.ExamSubTypeAcademic, 7.0},
		{"Academic 27 correct", 27, enums.ExamSubTypeAcademic, 6.5},
		{"Academic 23 correct", 23, enums.ExamSubTypeAcademic, 6.0},
		{"Academic 19 correct", 19, enums.ExamSubTypeAcademic, 5.5},
		{"Academic 15 correct", 15, enums.ExamSubTypeAcademic, 5.0},
		{"Academic 13 correct", 13, enums.ExamSubTypeAcademic, 4.5},
		{"Academic 10 correct", 10, enums.ExamSubTypeAcademic, 4.0},
		{"Academic 4 correct", 4, enums.ExamSubTypeAcademic, 2.5},
		{"Academic 0 correct", 0, enums.ExamSubTypeAcademic, 0.0},
		{"Academic negative", -1, enums.ExamSubTypeAcademic, 0.0},
		{"Academic capped at 40", 45, enums.ExamSubTypeAcademic, 9.0},

		// General Training Reading
		{"GT 40 correct", 40, enums.ExamSubTypeGeneralTraining, 9.0},
		{"GT 39 correct", 39, enums.ExamSubTypeGeneralTraining, 8.5},
		{"GT 38 correct", 38, enums.ExamSubTypeGeneralTraining, 8.0},
		{"GT 36 correct", 36, enums.ExamSubTypeGeneralTraining, 7.5},
		{"GT 34 correct", 34, enums.ExamSubTypeGeneralTraining, 7.0},
		{"GT 32 correct", 32, enums.ExamSubTypeGeneralTraining, 6.5},
		{"GT 30 correct", 30, enums.ExamSubTypeGeneralTraining, 6.0},
		{"GT 27 correct", 27, enums.ExamSubTypeGeneralTraining, 5.5},
		{"GT 23 correct", 23, enums.ExamSubTypeGeneralTraining, 5.0},
		{"GT 19 correct", 19, enums.ExamSubTypeGeneralTraining, 4.5},
		{"GT 15 correct", 15, enums.ExamSubTypeGeneralTraining, 4.0},
		{"GT 12 correct", 12, enums.ExamSubTypeGeneralTraining, 3.5},
		{"GT 9 correct", 9, enums.ExamSubTypeGeneralTraining, 3.0},
		{"GT 6 correct", 6, enums.ExamSubTypeGeneralTraining, 2.5},
		{"GT 3 correct", 3, enums.ExamSubTypeGeneralTraining, 2.0},
		{"GT 1 correct", 1, enums.ExamSubTypeGeneralTraining, 1.0},
		{"GT 0 correct", 0, enums.ExamSubTypeGeneralTraining, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateReadingBand(tt.rawCorrect, tt.subType)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateListeningBand(t *testing.T) {
	tests := []struct {
		name       string
		rawCorrect int
		expected   float64
	}{
		{"Listening 40 correct", 40, 9.0},
		{"Listening 39 correct", 39, 9.0},
		{"Listening 37 correct", 37, 8.5},
		{"Listening 35 correct", 35, 8.0},
		{"Listening 32 correct", 32, 7.5},
		{"Listening 30 correct", 30, 7.0},
		{"Listening 26 correct", 26, 6.5},
		{"Listening 23 correct", 23, 6.0},
		{"Listening 18 correct", 18, 5.5},
		{"Listening 16 correct", 16, 5.0},
		{"Listening 13 correct", 13, 4.5},
		{"Listening 10 correct", 10, 4.0},
		{"Listening 0 correct", 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateListeningBand(tt.rawCorrect)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestValidateExamSections(t *testing.T) {
	// Reading tests (max 3 sections)
	require.NoError(t, ValidateExamSections(enums.ExamSkillReading, []int{1, 2, 3}, true))
	require.NoError(t, ValidateExamSections(enums.ExamSkillReading, []int{1}, false))
	require.NoError(t, ValidateExamSections(enums.ExamSkillReading, []int{2, 3}, false))
	require.ErrorIs(t, ValidateExamSections(enums.ExamSkillReading, []int{1, 2}, true), ErrExamModeRequiresAll)
	require.ErrorIs(t, ValidateExamSections(enums.ExamSkillReading, []int{1, 4}, false), ErrInvalidSectionIndex)
	require.ErrorIs(t, ValidateExamSections(enums.ExamSkillReading, []int{1, 1}, false), ErrDuplicateSections)
	require.ErrorIs(t, ValidateExamSections(enums.ExamSkillReading, []int{}, false), ErrInvalidSectionsCount)

	// Listening tests (max 4 sections)
	require.NoError(t, ValidateExamSections(enums.ExamSkillListening, []int{1, 2, 3, 4}, true))
	require.ErrorIs(t, ValidateExamSections(enums.ExamSkillListening, []int{1, 2, 3}, true), ErrExamModeRequiresAll)
}
