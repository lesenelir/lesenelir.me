//go:build darwin

package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/audio"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

// transcriber sends recordings to the speech model in the settings: as a
// FLAC file, about half the bytes of the base64 WAV that every model takes,
// unless the model has refused one.
type transcriber struct {
	client *openrouter.Client
	store  *store.Store

	mu      sync.Mutex
	wavOnly map[string]bool // models that refused a FLAC upload
}

func (t *transcriber) Transcribe(ctx context.Context, pcm []byte) (string, error) {
	s := t.store.Get()
	model := s.ASRModel
	t.mu.Lock()
	wavOnly := t.wavOnly[model]
	t.mu.Unlock()

	if !wavOnly {
		if flac, err := audio.FLAC(pcm); err != nil {
			log.Printf("flac: %v; sending WAV", err)
		} else {
			start := time.Now()
			res, err := t.client.TranscribeFile(ctx, model, openrouter.Audio{Data: flac, Format: "flac"}, s.ASRLanguage)
			log.Printf("asr %s: %v, flac file %d bytes for %v of audio, err=%v",
				model, time.Since(start).Round(time.Millisecond), len(flac), audio.Duration(pcm), err)
			if !refusedFormat(err) {
				return res.Text, noKey(err)
			}
			t.mu.Lock()
			if t.wavOnly == nil {
				t.wavOnly = map[string]bool{}
			}
			t.wavOnly[model] = true
			t.mu.Unlock()
		}
	}

	wav := audio.WAV(pcm)
	start := time.Now()
	res, err := t.client.Transcribe(ctx, model, openrouter.Audio{Data: wav, Format: "wav"}, s.ASRLanguage)
	log.Printf("asr %s: %v, base64 wav %d bytes, err=%v", model, time.Since(start).Round(time.Millisecond), len(wav), err)
	return res.Text, noKey(err)
}

// refusedFormat reports whether a request failed for what it sent rather
// than for the account, the network or the provider's health.
func refusedFormat(err error) bool {
	var e *openrouter.Error
	return errors.As(err, &e) && (e.Status == 400 || e.Status == 415 || e.Status == 422)
}

func noKey(err error) error {
	if errors.Is(err, openrouter.ErrNoKey) {
		return controller.ErrNoKey
	}
	return err
}

// rewriter cleans up or translates with the chat model in the settings.
type rewriter struct {
	client *openrouter.Client
	store  *store.Store
}

func (r rewriter) Rewrite(ctx context.Context, req prompt.Request) (string, error) {
	s := r.store.Get()
	start := time.Now()
	out, err := r.client.Chat(ctx, openrouter.ChatRequest{
		Model: s.LLMModel,
		Messages: []openrouter.Message{
			{Role: "system", Content: prompt.System(req)},
			{Role: "user", Content: prompt.User(req)},
		},
		Temperature: 0.1,
		// Room for the answer and a little reasoning, and a cap on a model
		// that runs away.
		MaxTokens: 1024 + 4*utf8.RuneCountInString(req.Transcript),
		Reasoning: s.ReasoningEffort(),
		Providers: s.Providers,
	})
	log.Printf("llm %s: %v, err=%v", s.LLMModel, time.Since(start).Round(time.Millisecond), err)
	return out, err
}
