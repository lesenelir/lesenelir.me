package openrouter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func server(t *testing.T, handler func(path string, body map[string]any) (int, string)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("Authorization = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		status, res := handler(r.URL.Path, body)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(res))
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Key: func() string { return "sk-test" }}
}

func TestTranscribe(t *testing.T) {
	c := server(t, func(path string, body map[string]any) (int, string) {
		if path != "/audio/transcriptions" {
			t.Errorf("path %s", path)
		}
		audio := body["input_audio"].(map[string]any)
		if audio["format"] != "wav" {
			t.Errorf("format %v", audio["format"])
		}
		if data, _ := base64.StdEncoding.DecodeString(audio["data"].(string)); string(data) != "RIFF" {
			t.Errorf("audio %q", data)
		}
		if _, ok := body["language"]; ok {
			t.Error("language sent while detecting it")
		}
		return 200, `{"text":" 你好 world ","usage":{"seconds":2.5,"cost":0.00002}}`
	})
	got, err := c.Transcribe(context.Background(), "qwen/qwen3-asr-1.7b", Audio{Data: []byte("RIFF"), Format: "wav"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "你好 world" || got.Seconds != 2.5 {
		t.Fatalf("got %+v", got)
	}
}

func TestTranscribeFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/transcriptions" || r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("%s with %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("model") != "openai/whisper-large-v3-turbo" || r.FormValue("language") != "zh" {
			t.Errorf("fields %v", r.MultipartForm.Value)
		}
		f, h, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(f)
		if string(data) != "fLaC" || h.Filename != "audio.flac" || h.Header.Get("Content-Type") != "audio/flac" {
			t.Errorf("file %q named %q of %q", data, h.Filename, h.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte(`{"text":"你好"}`))
	}))
	t.Cleanup(srv.Close)
	c := &Client{BaseURL: srv.URL, Key: func() string { return "sk-test" }}
	got, err := c.TranscribeFile(context.Background(), "openai/whisper-large-v3-turbo", Audio{Data: []byte("fLaC"), Format: "flac"}, "zh")
	if err != nil || got.Text != "你好" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestChat(t *testing.T) {
	c := server(t, func(path string, body map[string]any) (int, string) {
		if path != "/chat/completions" {
			t.Errorf("path %s", path)
		}
		if body["reasoning"].(map[string]any)["effort"] != "low" {
			t.Errorf("reasoning %v", body["reasoning"])
		}
		order := body["provider"].(map[string]any)["order"].([]any)
		if len(order) != 2 || order[0] != "cerebras" {
			t.Errorf("provider order %v", order)
		}
		return 200, `{"choices":[{"message":{"role":"assistant","content":"Hello."}}]}`
	})
	got, err := c.Chat(context.Background(), ChatRequest{
		Model:     "openai/gpt-oss-120b",
		Messages:  []Message{{Role: "user", Content: "hi"}},
		Reasoning: "low",
		Providers: []string{"cerebras", "groq"},
	})
	if err != nil || got != "Hello." {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestErrors(t *testing.T) {
	c := server(t, func(string, map[string]any) (int, string) {
		return 402, `{"error":{"code":402,"message":"Insufficient credits"}}`
	})
	_, err := c.Chat(context.Background(), ChatRequest{Model: "m"})
	var e *Error
	if !errors.As(err, &e) || e.Status != 402 || e.Message != "Insufficient credits" {
		t.Fatalf("err = %v", err)
	}

	c = server(t, func(string, map[string]any) (int, string) {
		return 200, `{"error":{"code":502,"message":"Provider returned error"}}`
	})
	if _, err := c.Chat(context.Background(), ChatRequest{Model: "m"}); !errors.As(err, &e) || e.Message != "Provider returned error" {
		t.Fatalf("error in a 200 = %v", err)
	}

	empty := &Client{Key: func() string { return "" }}
	if _, err := empty.Chat(context.Background(), ChatRequest{}); !errors.Is(err, ErrNoKey) {
		t.Fatalf("no key = %v", err)
	}
}
