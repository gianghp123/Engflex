package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/services"
)

func TestLessonService_ListSections(t *testing.T) {
	b1 := &models.LessonSection{ID: "sec-b1", Slug: "b1", Title: "B1 Threshold", Position: 3}
	b2 := &models.LessonSection{ID: "sec-b2", Slug: "b2", Title: "B2 Vantage", Position: 4}

	tests := []struct {
		name      string
		req       requests.ListLessons
		setupMock func(m *mocks.MockLessonSectionRepository, a *mocks.MockAttemptRepository)
		wantSecs  int
		wantUnits int
	}{
		{
			name: "unfiltered returns sections with units attached",
			req:  requests.ListLessons{},
			setupMock: func(m *mocks.MockLessonSectionRepository, a *mocks.MockAttemptRepository) {
				m.On("ListWithUnits", mock.Anything, repositories.LessonUnitFilter{}, false).
					Return([]*models.LessonSection{
						{ID: b1.ID, Slug: b1.Slug, Title: b1.Title, Position: b1.Position,
							Units: []*models.Lesson{{ID: "u1", SectionID: b1.ID}}},
						{ID: b2.ID, Slug: b2.Slug, Title: b2.Title, Position: b2.Position,
							Units: []*models.Lesson{{ID: "u2", SectionID: b2.ID}}},
					}, nil)
			},
			wantSecs:  2,
			wantUnits: 2,
		},
		{
			name: "level filter passes through to the repository",
			req:  requests.ListLessons{Level: enums.CEFR("B1")},
			setupMock: func(m *mocks.MockLessonSectionRepository, a *mocks.MockAttemptRepository) {
				m.On("ListWithUnits", mock.Anything, repositories.LessonUnitFilter{Level: enums.CEFR("B1")}, false).
					Return([]*models.LessonSection{
						{ID: b1.ID, Slug: b1.Slug, Title: b1.Title, Position: b1.Position,
							Units: []*models.Lesson{{ID: "u1", SectionID: b1.ID}}},
					}, nil)
			},
			wantSecs:  1,
			wantUnits: 1,
		},
		{
			name: "search and skill filters pass through to the repository",
			req:  requests.ListLessons{Skill: enums.ActivityTypeReading},
			setupMock: func(m *mocks.MockLessonSectionRepository, a *mocks.MockAttemptRepository) {
				m.On("ListWithUnits", mock.Anything, repositories.LessonUnitFilter{Skill: enums.ActivityTypeReading}, false).
					Return([]*models.LessonSection{
						{ID: b1.ID, Slug: b1.Slug, Title: b1.Title, Position: b1.Position,
							Units: []*models.Lesson{{ID: "u1", SectionID: b1.ID}}},
					}, nil)
			},
			wantSecs:  1,
			wantUnits: 1,
		},
		{
			name: "status filter keeps only matching units and drops empty sections",
			req:  requests.ListLessons{Status: enums.LessonStatusCompleted},
			setupMock: func(m *mocks.MockLessonSectionRepository, a *mocks.MockAttemptRepository) {
				m.On("ListWithUnits", mock.Anything, repositories.LessonUnitFilter{}, false).
					Return([]*models.LessonSection{
						{ID: b1.ID, Slug: b1.Slug, Title: b1.Title, Position: b1.Position,
							Units: []*models.Lesson{{ID: "u1", SectionID: b1.ID}, {ID: "u2", SectionID: b1.ID}}},
						{ID: b2.ID, Slug: b2.Slug, Title: b2.Title, Position: b2.Position,
							Units: []*models.Lesson{{ID: "u3", SectionID: b2.ID}}},
					}, nil)
				lessonID := "u1"
				a.On("ListLessonAttempts", mock.Anything, "user_1").Return([]*models.Attempt{
					{LessonID: &lessonID, Status: enums.AttemptStatusCompleted},
				}, nil)
			},
			wantSecs:  1,
			wantUnits: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secMock := mocks.NewMockLessonSectionRepository(t)
			attemptsMock := mocks.NewMockAttemptRepository(t)
			tt.setupMock(secMock, attemptsMock)
			svc := services.NewLessonService(secMock, mocks.NewMockLessonRepository(t), attemptsMock, fakePremiumChecker{})

			got, err := svc.ListSections(context.Background(), "user_1", tt.req, false)

			require.NoError(t, err)
			require.Len(t, got, tt.wantSecs)
			n := 0
			for _, s := range got {
				n += len(s.Units)
			}
			assert.Equal(t, tt.wantUnits, n)
		})
	}
}
