// Package openrouter calls the two OpenRouter endpoints the app needs:
// speech to text and chat completions. One API key covers both.
package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

// DefaultBaseURL is OpenRouter's API.
const DefaultBaseURL = "https://openrouter.ai/api/v1"

// Client is safe for concurrent use.
type Client struct {
	BaseURL string
	// Key returns the API key; it is read on every request so that a key
	// changed in the settings applies at once.
	Key  func() string
	HTTP *http.Client
}

// ErrNoKey is returned when no API key is set.
var ErrNoKey = errors.New("openrouter: no API key")

// Error is a failed request, with OpenRouter's message.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("openrouter: %d %s", e.Status, e.Message)
}

// Transcription is the result of speech to text.
type Transcription struct {
	Text string
	// Seconds is the audio length OpenRouter billed, Cost its price in
	// USD; both are 0 when it does not say.
	Seconds float64
	Cost    float64
}

// Audio is a recording in one of the formats OpenRouter takes: "wav",
// "flac", "mp3", "m4a", "ogg", "webm" or "aac".
type Audio struct {
	Data   []byte
	Format string
}

type transcriptionResponse struct {
	Text  string `json:"text"`
	Usage struct {
		Seconds float64 `json:"seconds"`
		Cost    float64 `json:"cost"`
	} `json:"usage"`
}

func (r transcriptionResponse) result() Transcription {
	return Transcription{Text: strings.TrimSpace(r.Text), Seconds: r.Usage.Seconds, Cost: r.Usage.Cost}
}

// Transcribe turns audio into text, sent base64 in JSON, which every model
// takes. An empty language lets the model detect it, which mixed Chinese
// and English speech needs.
func (c *Client) Transcribe(ctx context.Context, model string, audio Audio, language string) (Transcription, error) {
	body := map[string]any{
		"model": model,
		"input_audio": map[string]string{
			"data":   base64.StdEncoding.EncodeToString(audio.Data),
			"format": audio.Format,
		},
		"temperature": 0,
	}
	if language != "" {
		body["language"] = language
	}
	b, err := json.Marshal(body)
	if err != nil {
		return Transcription{}, err
	}
	var res transcriptionResponse
	if err := c.post(ctx, "/audio/transcriptions", "application/json", b, &res); err != nil {
		return Transcription{}, err
	}
	return res.result(), nil
}

// TranscribeFile is Transcribe with the audio uploaded as a file, the
// OpenAI way: a third fewer bytes than base64, up to 25 MB.
func (c *Client) TranscribeFile(ctx context.Context, model string, audio Audio, language string) (Transcription, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := [][2]string{{"model", model}, {"temperature", "0"}}
	if language != "" {
		fields = append(fields, [2]string{"language", language})
	}
	for _, f := range fields {
		if err := w.WriteField(f[0], f[1]); err != nil {
			return Transcription{}, err
		}
	}
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="file"; filename="audio.` + audio.Format + `"`},
		"Content-Type":        {"audio/" + audio.Format},
	})
	if err != nil {
		return Transcription{}, err
	}
	if _, err := part.Write(audio.Data); err != nil {
		return Transcription{}, err
	}
	if err := w.Close(); err != nil {
		return Transcription{}, err
	}
	var res transcriptionResponse
	if err := c.post(ctx, "/audio/transcriptions", w.FormDataContentType(), buf.Bytes(), &res); err != nil {
		return Transcription{}, err
	}
	return res.result(), nil
}

// Message is one chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is a chat completion request. Zero fields are left out.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	// Reasoning is the effort: "none", "minimal", "low"…; "" leaves the
	// model's default.
	Reasoning string `json:"-"`
	// Providers are tried in this order before OpenRouter's own choice.
	Providers []string `json:"-"`
}

// Chat returns the assistant's reply.
func (c *Client) Chat(ctx context.Context, r ChatRequest) (string, error) {
	body := map[string]any{
		"model":       r.Model,
		"messages":    r.Messages,
		"temperature": r.Temperature,
		"stream":      false,
	}
	if r.MaxTokens > 0 {
		body["max_tokens"] = r.MaxTokens
	}
	if r.Reasoning != "" {
		body["reasoning"] = map[string]any{"effort": r.Reasoning, "exclude": true}
	}
	if len(r.Providers) > 0 {
		body["provider"] = map[string]any{"order": r.Providers, "allow_fallbacks": true}
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	if err := c.post(ctx, "/chat/completions", "application/json", b, &res); err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", errors.New("openrouter: no choices in the response")
	}
	return res.Choices[0].Message.Content, nil
}

func (c *Client) post(ctx context.Context, path, contentType string, body []byte, out any) error {
	key := ""
	if c.Key != nil {
		key = c.Key()
	}
	if key == "" {
		return ErrNoKey
	}
	base := c.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("HTTP-Referer", "https://lesenelir.me")
	req.Header.Set("X-Title", "Voice")
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	// OpenRouter may also answer 200 with an error object, when the
	// upstream provider fails mid-request.
	var probe struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(data, &probe)
	if resp.StatusCode/100 != 2 || probe.Error != nil {
		msg := ""
		if probe.Error != nil {
			msg = probe.Error.Message
		}
		if msg == "" {
			msg = strings.TrimSpace(string(data[:min(len(data), 200)]))
		}
		return &Error{Status: resp.StatusCode, Message: msg}
	}
	return json.Unmarshal(data, out)
}
