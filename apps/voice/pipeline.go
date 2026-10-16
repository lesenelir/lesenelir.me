//go:build darwin

package main

import (
	"context"
	"errors"
	"log"
	"time"
	"unicode/utf8"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

// transcriber sends recordings to the speech model in the settings.
type transcriber struct {
	client *openrouter.Client
	store  *store.Store
}

func (t transcriber) Transcribe(ctx context.Context, wav []byte) (string, error) {
	s := t.store.Get()
	start := time.Now()
	res, err := t.client.Transcribe(ctx, s.ASRModel, wav, s.ASRLanguage)
	log.Printf("asr %s: %v, %d bytes, err=%v", s.ASRModel, time.Since(start).Round(time.Millisecond), len(wav), err)
	if errors.Is(err, openrouter.ErrNoKey) {
		return "", controller.ErrNoKey
	}
	return res.Text, err
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
