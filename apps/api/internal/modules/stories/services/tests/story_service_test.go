package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/stories/dtos/requests"
	"engflex-api/internal/modules/stories/services"
)

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}

func TestStoryService_ListStories(t *testing.T) {
	stories := []*models.Story{
		{ID: "11111111-1111-1111-1111-111111111111", Slug: "cinderella", Title: "Cinderella"},
		{ID: "22222222-2222-2222-2222-222222222222", Slug: "snow-white", Title: "Snow White"},
	}

	tests := []struct {
		name       string
		req        requests.ListStories
		mockReturn []*models.Story
		mockCount  int64
		mockErr    error
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "list without filters",
			req:        requests.ListStories{},
			mockReturn: stories,
			mockCount:  2,
			wantErr:    false,
		},
		{
			name:       "database error returns 500",
			req:        requests.ListStories{},
			mockReturn: nil,
			mockErr:    gorm.ErrInvalidDB,
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStoryRepository(t)
			f := repositories.StoryFilter{
				CEFR:          string(tt.req.CEFR),
				Topic:         tt.req.Topic,
				Search:        tt.req.Search,
				OnlyPublished: true,
				Limit:         10,
				Offset:        0,
			}
			repo.On("List", mock.Anything, f).Return(tt.mockReturn, tt.mockErr)
			if !tt.wantErr {
				repo.On("Count", mock.Anything, f).Return(tt.mockCount, nil)
			}
			svc := services.NewStoryService(repo, nil, nil)

			list, total, err := svc.ListStories(context.Background(), tt.req, 10, 0, false)
			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, list)
				return
			}
			require.NoError(t, err)
			assert.Len(t, list, 2)
			assert.Equal(t, int64(2), total)
		})
	}
}

func TestStoryService_GetBySlug(t *testing.T) {
	slug := "cinderella"
	story := &models.Story{
		ID:     "11111111-1111-1111-1111-111111111111",
		Slug:   slug,
		Title:  "Cinderella",
		Status: enums.PublishStatusPublished,
		Sentences: []*models.StorySentence{
			{ID: "33333333-3333-3333-3333-333333333333", Sequence: 1, EnglishText: "Once upon a time..."},
		},
	}

	tests := []struct {
		name       string
		slug       string
		mockReturn *models.Story
		mockErr    error
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "found story",
			slug:       slug,
			mockReturn: story,
			wantErr:    false,
		},
		{
			name:       "not found returns 404",
			slug:       "unknown",
			mockReturn: nil,
			mockErr:    gorm.ErrRecordNotFound,
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStoryRepository(t)
			repo.On("GetBySlug", mock.Anything, tt.slug).Return(tt.mockReturn, tt.mockErr)
			svc := services.NewStoryService(repo, nil, nil)

			got, err := svc.GetBySlug(context.Background(), tt.slug, "user-1", false)
			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, story.Title, got.Title)
			assert.Len(t, got.Sentences, 1)
		})
	}
}

func TestStoryService_ListTopics(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []string
		mockErr    error
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "success",
			mockReturn: []string{"andersen", "grimm"},
			wantErr:    false,
		},
		{
			name:       "db error returns 500",
			mockReturn: nil,
			mockErr:    gorm.ErrInvalidDB,
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStoryRepository(t)
			repo.On("ListTopics", mock.Anything).Return(tt.mockReturn, tt.mockErr)
			svc := services.NewStoryService(repo, nil, nil)

			got, err := svc.ListTopics(context.Background())
			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.mockReturn, got)
		})
	}
}
