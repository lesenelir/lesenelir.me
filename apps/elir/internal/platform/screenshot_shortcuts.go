package platform

import (
	"context"
	"runtime"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

type screenshotKeys struct{}
type screenshotInput struct {
	bindings shortcut.Bindings
	claim    func(string)
}

func WithScreenshotShortcuts(ctx context.Context, b shortcut.Bindings, claim func(string)) context.Context {
	return context.WithValue(ctx, screenshotKeys{}, screenshotInput{shortcut.Clone(b), claim})
}
func screenshotBindings(ctx context.Context) shortcut.Bindings {
	if ctx != nil {
		if input, ok := ctx.Value(screenshotKeys{}).(screenshotInput); ok {
			return input.bindings
		}
	}
	return shortcut.Defaults(runtime.GOOS)
}
func claimScreenshotKey(ctx context.Context, key string) {
	if ctx != nil {
		if input, ok := ctx.Value(screenshotKeys{}).(screenshotInput); ok && input.claim != nil {
			input.claim(key)
		}
	}
}
