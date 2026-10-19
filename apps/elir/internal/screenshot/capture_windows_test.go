package screenshot

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
)

func TestCaptureCommandUsesTrustedSystemPaths(t *testing.T) {
	untrusted := t.TempDir()
	t.Setenv("PATH", untrusted)
	t.Setenv("SystemRoot", untrusted)
	t.Setenv("PSModulePath", untrusted)
	t.Setenv("OPENROUTER_API_KEY", "must-not-reach-child")
	t.Setenv("SONIOX_API_KEY", "must-not-reach-child")
	t.Setenv("ELIR_SCREENSHOT_PATH", "wrong-output")
	path := filepath.Join(t.TempDir(), "image ' & $;.png")
	cmd := captureCommand(context.Background(), path)
	home := filepath.Join(platform.SystemDirectory(), "WindowsPowerShell", "v1.0")
	if cmd.Path != filepath.Join(home, "powershell.exe") || cmd.Dir != platform.SystemDirectory() {
		t.Fatalf("unsafe process location: %s in %s", cmd.Path, cmd.Dir)
	}
	if cmd.Args[len(cmd.Args)-1] != captureScript || strings.Contains(cmd.Args[len(cmd.Args)-1], path) {
		t.Fatal("output path must be data, not embedded in the command")
	}
	var destination, modules string
	for _, entry := range cmd.Env {
		name, value, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(name) {
		case "OPENROUTER_API_KEY", "SONIOX_API_KEY":
			t.Fatal("API key inherited by screenshot helper")
		case "ELIR_SCREENSHOT_PATH":
			destination = value
		case "PSMODULEPATH":
			modules = value
		}
	}
	if destination != path || modules != filepath.Join(home, "Modules") {
		t.Fatal("unsafe helper environment")
	}
}

func TestScreenshotPowerShellPrelude(t *testing.T) {
	// Compile the embedded interop declarations without reading the screen or
	// opening the selection window, using the same restricted child environment.
	cmd := captureCommand(context.Background(), filepath.Join(t.TempDir(), "unused.png"))
	cmd.Args[len(cmd.Args)-1] = strings.SplitN(captureScript, "[void][ElirCapture]", 2)[0]
	if _, err := os.Stat(cmd.Path); err != nil {
		t.Skip("Windows PowerShell unavailable")
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("helper declarations: %v: %s", err, out)
	}
}
