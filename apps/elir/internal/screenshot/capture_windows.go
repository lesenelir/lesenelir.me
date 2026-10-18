package screenshot

import (
	"context"

	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

func capture(ctx context.Context, path string, bindings shortcut.Bindings, router *shortcut.Router) (bool, error) {
	return platform.CaptureScreenshot(platform.WithScreenshotShortcuts(ctx, bindings, func(key string) { mygo.RunOnMain(func() { router.Claim(key) }) }), path)
}
