package screenshot

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

//go:embed capture_windows.ps1
var captureScript string

// Windows PowerShell and WinForms ship with Windows 10/11. The helper takes a
// desktop snapshot before showing its selection window; Esc leaves no file.
func capture(ctx context.Context, path string) (bool, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-STA", "-WindowStyle", "Hidden", "-Command", captureScript)
	cmd.Env = append(os.Environ(), "ELIR_SCREENSHOT_PATH="+path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 2 {
			return false, nil
		}
		return false, fmt.Errorf("screen capture: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if _, err := os.Stat(path); err != nil {
		return false, err
	}
	return true, nil
}
