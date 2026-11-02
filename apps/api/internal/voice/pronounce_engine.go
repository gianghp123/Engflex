package voice

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

// PronounceEngine scores one attempt against its reference sentence on the
// OpenPronounce container.
type PronounceEngine interface {
	Pronounce(ctx context.Context, expectedText, lang string, audio []byte, mime string) (*PronounceResult, error)
}

// pronounceMimeSuffix maps browser upload containers to the filename suffix
// the container needs to decode via libsndfile/ffmpeg. Unknown containers
// default to .webm, the container the web recorder produces.
var pronounceMimeSuffix = map[string]string{
	"audio/webm": ".webm",
	"audio/mp4":  ".mp4",
	"audio/wav":  ".wav",
	"audio/mpeg": ".mp3",
	"audio/ogg":  ".ogg",
}

type httpPronounceEngine struct {
	baseURL string
	http    *http.Client
	timeout time.Duration
}

func NewHTTPPronounceEngine(serviceURL string, timeout time.Duration) PronounceEngine {
	return &httpPronounceEngine{baseURL: serviceURL, http: &http.Client{}, timeout: timeout}
}

func (e *httpPronounceEngine) Pronounce(ctx context.Context, expectedText, lang string, audio []byte, mime string) (*PronounceResult, error) {
	// One timeout around both calls: it covers a cold container start.
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	assessmentRaw, err := e.postPronunciation(ctx, expectedText, lang, audio, mime)
	if err != nil {
		return nil, err
	}
	phonemesRaw, err := e.postPhonemes(ctx, expectedText, lang)
	if err != nil {
		return nil, err
	}
	return mapContainerAssessment(assessmentRaw, phonemesRaw)
}

func (e *httpPronounceEngine) postPronunciation(ctx context.Context, expectedText, lang string, audio []byte, mime string) ([]byte, error) {
	suffix := pronounceMimeSuffix[mime]
	if suffix == "" {
		suffix = ".webm"
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	// CreatePart, not CreateFormFile: the latter labels every part
	// application/octet-stream, and the engine gates on the real container.
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="attempt`+suffix+`"`)
	partHeader.Set("Content-Type", mime)
	fw, err := w.CreatePart(partHeader)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(audio); err != nil {
		return nil, err
	}
	if err := w.WriteField("expected_text", expectedText); err != nil {
		return nil, err
	}
	if err := w.WriteField("lang", lang); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/pronunciation", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &TranscriptStatusError{Path: "/pronunciation", Status: resp.StatusCode, Body: raw}
	}
	return raw, nil
}

func (e *httpPronounceEngine) postPhonemes(ctx context.Context, expectedText, lang string) ([]byte, error) {
	form := url.Values{"text": {expectedText}, "lang": {lang}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/phonemes", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &TranscriptStatusError{Path: "/phonemes", Status: resp.StatusCode, Body: raw}
	}
	return raw, nil
}
