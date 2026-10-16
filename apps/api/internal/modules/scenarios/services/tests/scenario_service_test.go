package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/scenarios/dtos/requests"
	"engflex-api/internal/modules/scenarios/services"
	"gorm.io/gorm"
)

type boomErr struct{}

func (boomErr) Error() string { return "boom" }

func requireAppError(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, status, appErr.Status)
}

func TestScenarioList(t *testing.T) {
	topicID := "99999999-9999-9999-9999-999999999999"
	tests := []struct {
		name      string
		req       requests.ListScenarios
		setup     func(*repomocks.MockScenarioRepository)
		wantItems int
		wantTotal int64
		wantErr   int
	}{
		{
			name: "returns built-in scenarios",
			req:  requests.ListScenarios{ListParams: common.ListParams{Page: 1, PageSize: 20}},
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("List", mock.Anything, (*string)(nil), false, 20, 0).Return([]*models.Scenario{{ID: "a"}}, nil).Once()
				repo.On("Count", mock.Anything, (*string)(nil)).Return(int64(1), nil).Once()
			},
			wantItems: 1,
			wantTotal: 1,
		},
		{
			name: "topic filter is passed through",
			req: requests.ListScenarios{
				ListParams: common.ListParams{Page: 1, PageSize: 20},
				TopicID:    &topicID,
			},
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("List", mock.Anything, &topicID, false, 20, 0).Return([]*models.Scenario{}, nil).Once()
				repo.On("Count", mock.Anything, &topicID).Return(int64(0), nil).Once()
			},
			wantTotal: 0,
		},
		{
			name: "random order flag is passed through",
			req: requests.ListScenarios{
				ListParams: common.ListParams{Page: 1, PageSize: 2},
				Random:     true,
			},
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("List", mock.Anything, (*string)(nil), true, 2, 0).Return([]*models.Scenario{{ID: "a"}}, nil).Once()
				repo.On("Count", mock.Anything, (*string)(nil)).Return(int64(1), nil).Once()
			},
			wantItems: 1,
			wantTotal: 1,
		},
		{
			name: "list failure maps to 500",
			req:  requests.ListScenarios{ListParams: common.ListParams{Page: 1, PageSize: 20}},
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("List", mock.Anything, (*string)(nil), false, 20, 0).Return(nil, boomErr{}).Once()
			},
			wantErr: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockScenarioRepository(t)
			tt.setup(repo)
			svc := services.NewScenarioService(repo, repomocks.NewMockScenarioTopicRepository(t))

			items, total, err := svc.List(context.Background(), tt.req)

			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantItems)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestScenarioTopicsPreview(t *testing.T) {
	b2 := "B2"
	tests := []struct {
		name         string
		previewK     int
		difficulty   *string
		search       string
		setup        func(*repomocks.MockScenarioTopicRepository)
		wantTopics   int
		wantPreviews []int
		wantShort    []string
		wantErr      int
	}{
		{
			name:     "zero k clamps to three and banners carry topics with previews",
			previewK: 0,
			setup: func(topics *repomocks.MockScenarioTopicRepository) {
				topics.On("ListWithPreview", mock.Anything, 3, (*string)(nil), "").
					Return([]*models.ScenarioTopic{
						{
							ID: "t1", Slug: "job-interviews", ShortName: "Job interviews", Position: 1,
							Scenarios: []*models.Scenario{
								{ID: "s1", TopicID: "t1"},
								{ID: "s2", TopicID: "t1"},
							},
						},
						{
							ID: "t2", Slug: "product-pitch", ShortName: "Product pitch", Position: 4,
						},
					}, nil).Once()
			},
			wantTopics:   2,
			wantPreviews: []int{2, 0},
			wantShort:    []string{"Job interviews", "Product pitch"},
		},
		{
			name:       "huge k clamps to six and passes filters through",
			previewK:   999,
			difficulty: &b2,
			search:     "pitch",
			setup: func(topics *repomocks.MockScenarioTopicRepository) {
				topics.On("ListWithPreview", mock.Anything, 6, &b2, "pitch").
					Return([]*models.ScenarioTopic{}, nil).Once()
			},
			wantTopics:   0,
			wantPreviews: []int{},
			wantShort:    []string{},
		},
		{
			name:     "repo failure maps to 500",
			previewK: 3,
			setup: func(topics *repomocks.MockScenarioTopicRepository) {
				topics.On("ListWithPreview", mock.Anything, 3, (*string)(nil), "").
					Return(nil, boomErr{}).Once()
			},
			wantErr: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topics := repomocks.NewMockScenarioTopicRepository(t)
			tt.setup(topics)
			svc := services.NewScenarioService(repomocks.NewMockScenarioRepository(t), topics)

			got, err := svc.ListTopicsWithPreview(context.Background(), tt.previewK, tt.difficulty, tt.search)

			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tt.wantTopics)
			for i, topic := range got {
				assert.Len(t, topic.Scenarios, tt.wantPreviews[i])
				assert.Equal(t, tt.wantShort[i], topic.ShortName)
			}
		})
	}
}

func TestScenarioGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(*repomocks.MockScenarioRepository)
		wantID  string
		wantErr int
	}{
		{
			name: "returns the scenario",
			id:   "01b9a032-e2cf-5bb0-8df5-7bf49704bee8",
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, "01b9a032-e2cf-5bb0-8df5-7bf49704bee8").
					Return(&models.Scenario{ID: "01b9a032-e2cf-5bb0-8df5-7bf49704bee8", Title: "Behavioral questions on leadership"}, nil).Once()
			},
			wantID: "01b9a032-e2cf-5bb0-8df5-7bf49704bee8",
		},
		{
			name: "unknown id maps to 404",
			id:   "00000000-0000-0000-0000-000000000000",
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, "00000000-0000-0000-0000-000000000000").
					Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantErr: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockScenarioRepository(t)
			tt.setup(repo)
			svc := services.NewScenarioService(repo, repomocks.NewMockScenarioTopicRepository(t))

			got, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}

func TestScenarioGetDetail(t *testing.T) {
	scenarioID := "3a8b3191-c8a3-5855-8e05-0a268cedac3e"
	full := &models.Scenario{
		ID: scenarioID, Title: "Sprint retrospective",
		Persona: &models.Persona{ID: "473d21cd-6a87-5e3a-b4eb-b9638b9d2bb2", Name: "Tom"},
		Topic:   &models.ScenarioTopic{ID: "9735face-d870-5b51-ae3a-e18768a282fc", ShortName: "Architecture reviews"},
	}
	bare := &models.Scenario{
		ID: scenarioID, Title: "Sprint retrospective",
		Topic: &models.ScenarioTopic{ID: "9735face-d870-5b51-ae3a-e18768a282fc", ShortName: "Architecture reviews"},
	}
	tests := []struct {
		name        string
		id          string
		setup       func(*repomocks.MockScenarioRepository)
		wantPersona bool
		wantErr     int
	}{
		{
			name: "full detail with persona and topic",
			id:   scenarioID,
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("GetDetail", mock.Anything, scenarioID).Return(full, nil).Once()
			},
			wantPersona: true,
		},
		{
			name: "missing persona degrades to nil",
			id:   scenarioID,
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("GetDetail", mock.Anything, scenarioID).Return(bare, nil).Once()
			},
			wantPersona: false,
		},
		{
			name: "unknown scenario maps to 404",
			id:   "00000000-0000-0000-0000-000000000000",
			setup: func(repo *repomocks.MockScenarioRepository) {
				repo.On("GetDetail", mock.Anything, "00000000-0000-0000-0000-000000000000").
					Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantErr: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockScenarioRepository(t)
			tt.setup(repo)
			svc := services.NewScenarioService(repo, repomocks.NewMockScenarioTopicRepository(t))

			got, err := svc.GetDetail(context.Background(), tt.id)

			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, scenarioID, got.ID)
			require.NotNil(t, got.Topic)
			assert.Equal(t, "Architecture reviews", got.Topic.ShortName)
			if tt.wantPersona {
				require.NotNil(t, got.Persona)
				assert.Equal(t, "Tom", got.Persona.Name)
			} else {
				assert.Nil(t, got.Persona)
			}
		})
	}
}
