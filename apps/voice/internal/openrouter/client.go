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
	"net/http"
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

// Transcribe turns WAV audio into text. An empty language lets the model
// detect it, which mixed Chinese and English speech needs.
func (c *Client) Transcribe(ctx context.Context, model string, wav []byte, language string) (Transcription, error) {
	body := map[string]any{
		"model": model,
		"input_audio": map[string]string{
			"data":   base64.StdEncoding.EncodeToString(wav),
			"format": "wav",
		},
		"temperature": 0,
	}
	if language != "" {
		body["language"] = language
	}
	var res struct {
		Text  string `json:"text"`
		Usage struct {
			Seconds float64 `json:"seconds"`
			Cost    float64 `json:"cost"`
		} `json:"usage"`
	}
	if err := c.post(ctx, "/audio/transcriptions", body, &res); err != nil {
		return Transcription{}, err
	}
	return Transcription{Text: strings.TrimSpace(res.Text), Seconds: res.Usage.Seconds, Cost: res.Usage.Cost}, nil
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
	if err := c.post(ctx, "/chat/completions", body, &res); err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", errors.New("openrouter: no choices in the response")
	}
	return res.Choices[0].Message.Content, nil
}

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	key := ""
	if c.Key != nil {
		key = c.Key()
	}
	if key == "" {
		return ErrNoKey
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	base := c.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
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
