package screenshot

import (
	"context"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
)

func capture(ctx context.Context, path string) (bool, error) {
	return platform.CaptureScreenshot(ctx, path)
}
