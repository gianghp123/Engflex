package tests

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/schemas"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestGenerateWordDraft pins the draft endpoint's service contract: the
// engine output is normalized and mapped; engine failure and unusable output
// are both a 503 with nothing written; a blank term never reaches the engine.
func TestGenerateWordDraft(t *testing.T) {
	valid := schemas.WordDraft{
		Definition:   "Arriving exactly on time.",
		CEFR:         "B2",
		PartOfSpeech: "adjective",
		Details: schemas.WordDraftDetails{
			IPA: "ˈpʌŋk.tʃu.əl",
			Contexts: []schemas.WordDraftContext{
				{Label: "At work", Quote: "She is always punctual."},
			},
			Collocations: []schemas.WordDraftCollocation{
				{Phrase: "a punctual start", Pattern: "adj + noun", Example: "We had a punctual start."},
			},
			WordForms: []schemas.WordDraftWordForm{
				{FormType: "noun", Forms: []string{"punctuality"}},
			},
		},
	}

	fill := func(out schemas.WordDraft) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			*args.Get(2).(*schemas.WordDraft) = out
		}
	}

	tests := []struct {
		name    string
		req     requests.GenerateWordDraft
		setup   func(*llmmocks.MockLLMClient)
		wantErr int
		check   func(t *testing.T, d *responses.WordDraft)
	}{
		{
			name: "draft is normalized and mapped",
			req:  requests.GenerateWordDraft{Term: "punctual", Context: "She is always punctual."},
			setup: func(client *llmmocks.MockLLMClient) {
				client.On("CompleteStructured", mock.Anything, mock.Anything, mock.Anything, schemas.WordDraftSchemaName, mock.Anything).
					Run(fill(valid)).Return(nil).Once()
			},
			check: func(t *testing.T, d *responses.WordDraft) {
				require.Equal(t, "Arriving exactly on time.", d.Definition)
				require.Equal(t, "B2", d.CEFR)
				require.Equal(t, "adjective", d.PartOfSpeech)
				// The nested slices cross package boundaries (schemas →
				// responses); the conversion must survive non-empty rows.
				require.Equal(t, "ˈpʌŋk.tʃu.əl", d.Details.IPA)
				require.Len(t, d.Details.Contexts, 1)
				require.Equal(t, "At work", d.Details.Contexts[0].Label)
				require.Len(t, d.Details.Collocations, 1)
				require.Len(t, d.Details.WordForms, 1)
			},
		},
		{
			name:    "blank term is a 400",
			req:     requests.GenerateWordDraft{Term: "  "},
			setup:   func(client *llmmocks.MockLLMClient) {},
			wantErr: http.StatusBadRequest,
		},
		{
			name: "engine failure is a 503",
			req:  requests.GenerateWordDraft{Term: "punctual"},
			setup: func(client *llmmocks.MockLLMClient) {
				client.On("CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("boom")).Once()
			},
			wantErr: http.StatusServiceUnavailable,
		},
		{
			name: "unusable draft is a 503",
			req:  requests.GenerateWordDraft{Term: "punctual"},
			setup: func(client *llmmocks.MockLLMClient) {
				bad := valid
				bad.CEFR = "B1+"
				client.On("CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(fill(bad)).Return(nil).Once()
			},
			wantErr: http.StatusServiceUnavailable,
		},
		{
			name: "output language vi asks for the vietnamese brief",
			req:  requests.GenerateWordDraft{Term: "punctual", OutputLanguage: "vi"},
			setup: func(client *llmmocks.MockLLMClient) {
				client.On("CompleteStructured", mock.Anything,
					mock.MatchedBy(func(prompt string) bool {
						return strings.Contains(prompt, "Write the explanations in Vietnamese")
					}), mock.Anything, schemas.WordDraftSchemaName, mock.Anything).
					Run(fill(valid)).Return(nil).Once()
			},
			check: func(t *testing.T, d *responses.WordDraft) {
				require.Equal(t, "Arriving exactly on time.", d.Definition)
			},
		},
		{
			name: "output language en keeps the english brief",
			req:  requests.GenerateWordDraft{Term: "punctual", OutputLanguage: "en"},
			setup: func(client *llmmocks.MockLLMClient) {
				client.On("CompleteStructured", mock.Anything,
					mock.MatchedBy(func(prompt string) bool {
						return strings.Contains(prompt, "Write everything in English.")
					}), mock.Anything, schemas.WordDraftSchemaName, mock.Anything).
					Run(fill(valid)).Return(nil).Once()
			},
			check: func(t *testing.T, d *responses.WordDraft) {
				require.Equal(t, "B2", d.CEFR)
			},
		},
		{
			name:    "an unassigned language code is a 400",
			req:     requests.GenerateWordDraft{Term: "punctual", OutputLanguage: "zz"},
			setup:   func(client *llmmocks.MockLLMClient) {},
			wantErr: http.StatusBadRequest,
		},
		{
			name:    "a malformed language code is a 400",
			req:     requests.GenerateWordDraft{Term: "punctual", OutputLanguage: "not a tag"},
			setup:   func(client *llmmocks.MockLLMClient) {},
			wantErr: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := llmmocks.NewMockLLMClient(t)
			tt.setup(client)
			svc := services.NewVocabularyWordDraftService(client)
			got, err := svc.Generate(context.Background(), tt.req)
			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}
