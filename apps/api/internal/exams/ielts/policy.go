package ielts

import (
	"errors"
	"fmt"
	"sort"

	"engflex-api/internal/common/enums"
)

const (
	// Duration limits in minutes for standard IELTS test formats.
	ReadingDurationMinutes   = 60
	ListeningDurationMinutes = 40
	WritingDurationMinutes   = 60

	// Official IELTS section counts per skill.
	ReadingSectionsCount   = 3
	ListeningSectionsCount = 4
	WritingTasksCount      = 2

	// Standard objective question total per test.
	StandardTotalQuestions = 40
)

var (
	ErrInvalidSectionsCount = errors.New("invalid sections count")
	ErrExamModeRequiresAll  = errors.New("exam mode requires all sections to be selected")
	ErrInvalidSectionIndex  = errors.New("section index out of range")
	ErrDuplicateSections    = errors.New("duplicate section index in selection")
)

// CalculateReadingBand returns the IELTS Band Score (0.0 - 9.0) from raw correct count (0 - 40).
func CalculateReadingBand(rawCorrect int, subType enums.ExamSubType) float64 {
	if rawCorrect <= 0 {
		return 0.0
	}
	if rawCorrect > StandardTotalQuestions {
		rawCorrect = StandardTotalQuestions
	}

	if subType == enums.ExamSubTypeGeneralTraining {
		switch {
		case rawCorrect == 40:
			return 9.0
		case rawCorrect == 39:
			return 8.5
		case rawCorrect >= 37:
			return 8.0
		case rawCorrect == 36:
			return 7.5
		case rawCorrect >= 34:
			return 7.0
		case rawCorrect >= 32:
			return 6.5
		case rawCorrect >= 30:
			return 6.0
		case rawCorrect >= 27:
			return 5.5
		case rawCorrect >= 23:
			return 5.0
		case rawCorrect >= 19:
			return 4.5
		case rawCorrect >= 15:
			return 4.0
		case rawCorrect >= 12:
			return 3.5
		case rawCorrect >= 9:
			return 3.0
		case rawCorrect >= 6:
			return 2.5
		case rawCorrect >= 3:
			return 2.0
		default:
			return 1.0
		}
	}

	// Default to Academic Reading band score cutoffs
	switch {
	case rawCorrect >= 39:
		return 9.0
	case rawCorrect >= 37:
		return 8.5
	case rawCorrect >= 35:
		return 8.0
	case rawCorrect >= 33:
		return 7.5
	case rawCorrect >= 30:
		return 7.0
	case rawCorrect >= 27:
		return 6.5
	case rawCorrect >= 23:
		return 6.0
	case rawCorrect >= 19:
		return 5.5
	case rawCorrect >= 15:
		return 5.0
	case rawCorrect >= 13:
		return 4.5
	case rawCorrect >= 10:
		return 4.0
	case rawCorrect >= 8:
		return 3.5
	case rawCorrect >= 6:
		return 3.0
	case rawCorrect >= 4:
		return 2.5
	case rawCorrect >= 2:
		return 2.0
	default:
		return 1.0
	}
}

// CalculateListeningBand returns the IELTS Band Score (0.0 - 9.0) from raw correct count (0 - 40).
func CalculateListeningBand(rawCorrect int) float64 {
	if rawCorrect <= 0 {
		return 0.0
	}
	if rawCorrect > StandardTotalQuestions {
		rawCorrect = StandardTotalQuestions
	}

	switch {
	case rawCorrect >= 39:
		return 9.0
	case rawCorrect >= 37:
		return 8.5
	case rawCorrect >= 35:
		return 8.0
	case rawCorrect >= 32:
		return 7.5
	case rawCorrect >= 30:
		return 7.0
	case rawCorrect >= 26:
		return 6.5
	case rawCorrect >= 23:
		return 6.0
	case rawCorrect >= 18:
		return 5.5
	case rawCorrect >= 16:
		return 5.0
	case rawCorrect >= 13:
		return 4.5
	case rawCorrect >= 10:
		return 4.0
	case rawCorrect >= 8:
		return 3.5
	case rawCorrect >= 6:
		return 3.0
	case rawCorrect >= 4:
		return 2.5
	case rawCorrect >= 2:
		return 2.0
	default:
		return 1.0
	}
}

// ValidateExamSections checks whether the requested sections array is valid for the given skill and mode.
func ValidateExamSections(skill enums.ExamSkill, sections []int, isExamMode bool) error {
	if len(sections) == 0 {
		return ErrInvalidSectionsCount
	}

	var maxSections int
	switch skill {
	case enums.ExamSkillReading:
		maxSections = ReadingSectionsCount
	case enums.ExamSkillListening:
		maxSections = ListeningSectionsCount
	case enums.ExamSkillWriting:
		maxSections = WritingTasksCount
	default:
		maxSections = ReadingSectionsCount
	}

	if len(sections) > maxSections {
		return fmt.Errorf("%w: max %d allowed for %s", ErrInvalidSectionsCount, maxSections, skill)
	}

	sorted := append([]int(nil), sections...)
	sort.Ints(sorted)

	for i, s := range sorted {
		if s < 1 || s > maxSections {
			return fmt.Errorf("%w: section %d is out of [1, %d]", ErrInvalidSectionIndex, s, maxSections)
		}
		if i > 0 && s == sorted[i-1] {
			return fmt.Errorf("%w: duplicate section %d", ErrDuplicateSections, s)
		}
	}

	if isExamMode && len(sorted) != maxSections {
		return fmt.Errorf("%w: %s requires all %d sections", ErrExamModeRequiresAll, skill, maxSections)
	}

	return nil
}
