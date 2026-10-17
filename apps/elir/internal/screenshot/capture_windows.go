package screenshot

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
)

//go:embed capture_windows.ps1
var captureScript string

// Windows PowerShell and WinForms ship with Windows 10/11. The helper takes a
// desktop snapshot before showing its selection window; Esc leaves no file.
func capture(ctx context.Context, path string) (bool, error) {
	cmd := captureCommand(ctx, path)
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

func captureCommand(ctx context.Context, path string) *exec.Cmd {
	home := filepath.Join(platform.SystemDirectory(), "WindowsPowerShell", "v1.0")
	cmd := exec.CommandContext(ctx, filepath.Join(home, "powershell.exe"), "-NoProfile", "-NonInteractive", "-STA", "-WindowStyle", "Hidden", "-Command", captureScript)
	cmd.Dir = platform.SystemDirectory()
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(name) {
		case "OPENROUTER_API_KEY", "SONIOX_API_KEY", "PSMODULEPATH", "ELIR_SCREENSHOT_PATH":
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "ELIR_SCREENSHOT_PATH="+path, "PSModulePath="+filepath.Join(home, "Modules"))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
