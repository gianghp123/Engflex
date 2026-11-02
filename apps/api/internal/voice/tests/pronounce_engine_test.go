package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/voice"
)

func testPronounceEngine(t *testing.T, handler http.Handler) voice.PronounceEngine {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return voice.NewHTTPPronounceEngine(srv.URL, 5*time.Second)
}

func readRecorded(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return raw
}

// recordedPronounceServer serves the recorded container responses: the raw
// /pronunciation assessment plus the /phonemes word map.
func recordedPronounceServer(t *testing.T, assessment, phonemes []byte) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pronunciation":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(assessment)
		case "/phonemes":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(phonemes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func TestHTTPPronounceEngineRecordedMismatch(t *testing.T) {
	assessment := readRecorded(t, "container_pronunciation.json")
	phonemes := readRecorded(t, "container_phonemes.json")
	engine := testPronounceEngine(t, recordedPronounceServer(t, assessment, phonemes))

	got, err := engine.Pronounce(context.Background(), "Good morning my name is Kenji Sato", "en", []byte("fake-webm"), "audio/webm")
	require.NoError(t, err)
	require.NotNil(t, got)

	// PER 1.125 / WER 1.4286 both clip to zero: a fully-mismatched attempt
	// scores 0.0, not the container's blended library score.
	assert.Equal(t, 0.0, got.Score)
	assert.Equal(t, "THE WEATHER IS LOVELY IN PARIS THIS TIME OF YEAR", got.Transcription)
	assert.Equal(t, 1.125, got.PhonemeErrorRate)
	assert.Equal(t, 1.4286, got.WordErrorRate)
	assert.Equal(t, 16.55, got.AcousticDistance)

	require.Len(t, got.Errors, 7)
	first := got.Errors[0]
	assert.Equal(t, "good", first.Word)
	assert.Equal(t, "ɡʊd", first.Expected)
	assert.Equal(t, "ðəw", first.Heard)
	assert.Equal(t, 1.0, first.Confidence)
	require.NotEmpty(t, first.Phones)
	assert.Equal(t, "ɡ", first.Phones[0].Expected)
	assert.Equal(t, "ð", first.Phones[0].Heard)
	assert.Equal(t, 0.996, first.Phones[0].Confidence)

	// Reference phones come from /phonemes, grouped per word.
	require.Len(t, got.ReferencePhones, 7)
	assert.Equal(t, voice.ReferencePhone{Word: "Good", Phones: "ɡʊd"}, got.ReferencePhones[0])
	assert.Equal(t, voice.ReferencePhone{Word: "morning", Phones: "mɔːɹnɪŋ"}, got.ReferencePhones[1])

	// 46-point vectors and 107-point prosody stay whole (≤120 kept as-is).
	require.Len(t, got.ModelCurve, 46)
	assert.Equal(t, 609.0, got.ModelCurve[0])
	require.Len(t, got.LearnerCurve, 46)
	require.Len(t, got.Prosody.F0, 107)
	require.Len(t, got.Prosody.Energy, 107)
}

func TestHTTPPronounceEngineRecordedClean(t *testing.T) {
	assessment := readRecorded(t, "container_pronunciation_clean.json")
	phonemes := readRecorded(t, "container_phonemes.json")
	engine := testPronounceEngine(t, recordedPronounceServer(t, assessment, phonemes))

	got, err := engine.Pronounce(context.Background(), "Good morning my name is Kenji Sato", "en", []byte("fake-webm"), "audio/webm")
	require.NoError(t, err)
	require.NotNil(t, got)

	// Learning score from the recorded rates: 0.6*95.83 + 0.4*85.71 = 91.78.
	// A camelCase tag would parse PER/WER as 0 and score 100.0, so this pins
	// the snake_case tags too.
	assert.Equal(t, 91.78, got.Score)
	assert.Equal(t, "GOOD MORNING MY NAME IS KENJI SOTTO", got.Transcription)
	assert.Equal(t, 0.0417, got.PhonemeErrorRate)
	assert.Equal(t, 0.1429, got.WordErrorRate)

	// A clean reading has no errors, but the wire still says [] not null.
	require.NotNil(t, got.Errors)
	assert.Empty(t, got.Errors)
	raw, err := json.Marshal(got)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"errors":[]`)

	require.Len(t, got.ReferencePhones, 7)
	assert.Equal(t, "Good", got.ReferencePhones[0].Word)
	assert.Equal(t, "ɡʊd", got.ReferencePhones[0].Phones)
}

func TestHTTPPronounceEngineLearningScorePins(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{
			name: "blended rates",
			body: `{"acoustic_distance":8.434,"transcribe":"HELLO","differences":{"phoneme_error_rate":0.125,"word_error_rate":0.2857,"errors":[]}}`,
			want: 81.07,
		},
		{
			name: "rates above one clip to zero",
			body: `{"acoustic_distance":9.0,"transcribe":"HELLO","differences":{"phoneme_error_rate":1.75,"word_error_rate":1.5,"errors":[]}}`,
			want: 0.0,
		},
		{
			name: "perfect reading scores full",
			body: `{"acoustic_distance":0.0,"transcribe":"HELLO","differences":{"phoneme_error_rate":0.0,"word_error_rate":0.0,"errors":[]}}`,
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := testPronounceEngine(t, recordedPronounceServer(t, []byte(tt.body), []byte(`{"phonemes":[],"words":[]}`)))
			got, err := engine.Pronounce(context.Background(), "hello", "en", []byte("fake-webm"), "audio/webm")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Score)
		})
	}
}

func TestHTTPPronounceEngineNullAndMissing(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		check func(*testing.T, *voice.PronounceResult)
	}{
		{
			name: "null actual maps to empty heard",
			body: `{"transcribe":"HI","differences":{"phoneme_error_rate":0.5,"word_error_rate":0.5,` +
				`"errors":[{"word":"hi","expected":"haɪ","actual":null}]}}`,
			check: func(t *testing.T, got *voice.PronounceResult) {
				t.Helper()
				require.Len(t, got.Errors, 1)
				assert.Equal(t, "", got.Errors[0].Heard)
				// Word-path entries carry no confidence or phones.
				assert.Equal(t, 0.0, got.Errors[0].Confidence)
				require.NotNil(t, got.Errors[0].Phones)
				assert.Empty(t, got.Errors[0].Phones)
			},
		},
		{
			name: "missing differences: rates default to zero, score 100 without panic",
			body: `{"transcribe":"HI"}`,
			check: func(t *testing.T, got *voice.PronounceResult) {
				t.Helper()
				assert.Equal(t, "HI", got.Transcription)
				assert.Equal(t, 100.0, got.Score)
				assert.Equal(t, 0.0, got.PhonemeErrorRate)
				assert.Equal(t, 0.0, got.WordErrorRate)
				assert.Equal(t, 0.0, got.AcousticDistance)
				require.NotNil(t, got.Errors)
				assert.Empty(t, got.Errors)
				require.NotNil(t, got.ModelCurve)
				require.NotNil(t, got.LearnerCurve)
				require.NotNil(t, got.Prosody.F0)
				require.NotNil(t, got.Prosody.Energy)
				require.NotNil(t, got.ReferencePhones)
				raw, err := json.Marshal(got)
				require.NoError(t, err)
				assert.Contains(t, string(raw), `"errors":[]`)
				assert.Contains(t, string(raw), `"reference_phones":[]`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := testPronounceEngine(t, recordedPronounceServer(t, []byte(tt.body), []byte(`{"phonemes":[],"words":[]}`)))
			got, err := engine.Pronounce(context.Background(), "hi", "en", []byte("fake-webm"), "audio/webm")
			require.NoError(t, err)
			require.NotNil(t, got)
			tt.check(t, got)
		})
	}
}

func TestHTTPPronounceEngineDownsamplesCurves(t *testing.T) {
	points := func(n int) []float64 {
		out := make([]float64, n)
		for i := range out {
			out[i] = float64(i) + 0.5
		}
		return out
	}
	vec, err := json.Marshal(points(500))
	require.NoError(t, err)
	body := `{"transcribe":"HI","differences":{"phoneme_error_rate":0.0,"word_error_rate":0.0,"errors":[],` +
		`"expected_vector":` + string(vec) + `,"transcribed_vector":` + string(vec) + `},` +
		`"prosody":{"f0":` + string(vec) + `,"energy":` + string(vec) + `}}`
	engine := testPronounceEngine(t, recordedPronounceServer(t, []byte(body), []byte(`{"phonemes":[],"words":[]}`)))

	got, err := engine.Pronounce(context.Background(), "hi", "en", []byte("fake-webm"), "audio/webm")
	require.NoError(t, err)
	for name, curve := range map[string][]float64{
		"model": got.ModelCurve, "learner": got.LearnerCurve,
		"f0": got.Prosody.F0, "energy": got.Prosody.Energy,
	} {
		require.LessOrEqual(t, len(curve), 120, name)
		require.NotEmpty(t, curve, name)
		assert.Equal(t, 0.5, curve[0], name+" keeps the first frame")
	}
}

func TestHTTPPronounceEngineReferencePhonesGroupRuns(t *testing.T) {
	engine := testPronounceEngine(t, recordedPronounceServer(t,
		[]byte(`{"transcribe":"HI","differences":{"errors":[]}}`),
		// Two tokens for one word merge; a repeated adjacent word merges too.
		[]byte(`{"phonemes":["h","aɪ","w","ɜː","l","d","baɪ"],"words":["hi","hi","world","world","world","world","bye"]}`),
	))

	got, err := engine.Pronounce(context.Background(), "hi world world bye", "en", []byte("fake-webm"), "audio/webm")
	require.NoError(t, err)
	require.Len(t, got.ReferencePhones, 3)
	assert.Equal(t, voice.ReferencePhone{Word: "hi", Phones: "haɪ"}, got.ReferencePhones[0])
	assert.Equal(t, voice.ReferencePhone{Word: "world", Phones: "wɜːld"}, got.ReferencePhones[1])
	assert.Equal(t, voice.ReferencePhone{Word: "bye", Phones: "baɪ"}, got.ReferencePhones[2])
}

func TestHTTPPronounceEngineRequestShape(t *testing.T) {
	tests := []struct {
		name       string
		mime       string
		wantSuffix string
	}{
		{name: "webm", mime: "audio/webm", wantSuffix: ".webm"},
		{name: "mp4", mime: "audio/mp4", wantSuffix: ".mp4"},
		{name: "wav", mime: "audio/wav", wantSuffix: ".wav"},
		{name: "mpeg", mime: "audio/mpeg", wantSuffix: ".mp3"},
		{name: "ogg", mime: "audio/ogg", wantSuffix: ".ogg"},
		{name: "unknown defaults to webm", mime: "audio/x-unknown", wantSuffix: ".webm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotSuffix    string
				gotPartMIME  string
				gotText      string
				gotLang      string
				gotPhoneText string
				gotPhoneLang string
			)
			engine := testPronounceEngine(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/pronunciation":
					require.NoError(t, r.ParseMultipartForm(1<<20))
					_, header, err := r.FormFile("file")
					require.NoError(t, err)
					gotSuffix = header.Filename[strings.LastIndex(header.Filename, "."):]
					gotPartMIME = header.Header.Get("Content-Type")
					gotText = r.FormValue("expected_text")
					gotLang = r.FormValue("lang")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"transcribe":"HI","differences":{"errors":[]}}`))
				case "/phonemes":
					assert.Contains(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
					require.NoError(t, r.ParseForm())
					gotPhoneText = r.FormValue("text")
					gotPhoneLang = r.FormValue("lang")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"phonemes":[],"words":[]}`))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))

			_, err := engine.Pronounce(context.Background(), "Say hi.", "en", []byte("fake-audio"), tt.mime)
			require.NoError(t, err)
			assert.Equal(t, tt.wantSuffix, gotSuffix)
			assert.Equal(t, tt.mime, gotPartMIME)
			assert.Equal(t, "Say hi.", gotText)
			assert.Equal(t, "en", gotLang)
			assert.Equal(t, "Say hi.", gotPhoneText)
			assert.Equal(t, "en", gotPhoneLang)
		})
	}
}

func TestHTTPPronounceEngineNon2xx(t *testing.T) {
	tests := []struct {
		name       string
		pronStatus int
		phonStatus int
		wantPath   string
		wantStatus int
	}{
		{name: "pronunciation 422", pronStatus: http.StatusUnprocessableEntity, phonStatus: http.StatusOK, wantPath: "/pronunciation", wantStatus: http.StatusUnprocessableEntity},
		{name: "phonemes 500", pronStatus: http.StatusOK, phonStatus: http.StatusInternalServerError, wantPath: "/phonemes", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := testPronounceEngine(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/pronunciation":
					w.WriteHeader(tt.pronStatus)
					if tt.pronStatus >= 300 {
						_, _ = w.Write([]byte(`{"detail":"assessment failed"}`))
					} else {
						_, _ = w.Write([]byte(`{"transcribe":"HI","differences":{"errors":[]}}`))
					}
				case "/phonemes":
					w.WriteHeader(tt.phonStatus)
					if tt.phonStatus >= 300 {
						_, _ = w.Write([]byte(`{"detail":"phonemes failed"}`))
					} else {
						_, _ = w.Write([]byte(`{"phonemes":[],"words":[]}`))
					}
				}
			}))

			_, err := engine.Pronounce(context.Background(), "hi", "en", []byte("fake-audio"), "audio/webm")
			require.Error(t, err)
			var status *voice.TranscriptStatusError
			require.ErrorAs(t, err, &status)
			assert.Equal(t, tt.wantPath, status.Path)
			assert.Equal(t, tt.wantStatus, status.Status)
			assert.Contains(t, string(status.Body), "detail")
		})
	}
}
