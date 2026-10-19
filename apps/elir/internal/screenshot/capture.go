//go:build darwin

package screenshot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

// capture runs macOS's own interactive screenshot into path, as ⌘⇧4 does:
// drag over an area, or press Space and click a window; Esc cancels. Those
// keys are macOS's own there, so they are not among the shortcuts to set on
// macOS. It reports whether there is a screenshot, none when the user
// cancelled.
func capture(ctx context.Context, path string, _ shortcut.Bindings, _ *shortcut.Router) (bool, error) {
	out, err := exec.CommandContext(ctx, "/usr/sbin/screencapture", "-i", path).CombinedOutput()
	if _, statErr := os.Stat(path); statErr == nil {
		return true, nil
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return false, statErr
	}
	// Cancelling leaves no file and says nothing; a failure says why.
	if msg := strings.TrimSpace(string(out)); err != nil && msg != "" {
		return false, fmt.Errorf("screencapture: %s", msg)
	}
	return false, nil
}
