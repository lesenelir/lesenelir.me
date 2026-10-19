# Windows support

## Contract and scope

Keep MyGo v0.5.5 and the existing settings UI, transcription clients, controller,
history and screenshot library. Enable the desktop entry point on Windows and
separate OS integration behind `internal/platform`. Preserve the macOS backend.

Windows uses F8 / Shift+F8 for dictation / translation, Alt+T / Alt+R for selected
text, and Ctrl+Shift+2 for screenshots. Esc cancels. Implement microphone capture,
clipboard/key injection and interactive area capture with Windows APIs. Keep
secrets protected on Windows; do not rely on POSIX file mode bits there.

Platform-specific settings and help must describe Windows rather than macOS.
Provide Windows development/build commands without the macOS signing script.
Do not publish releases or change the existing macOS update feed.

## Verification

- [x] Inspect entry point, callers and pinned MyGo backend.
- [x] Create branch `codex/windows-support`.
- [x] Share application/UI code and add Windows platform integration.
- [x] Add regression tests for new platform logic and update documentation.
- [x] Run Windows tests, MyGo vet and Windows build.
- [x] Cross-compile macOS to check compatibility.
- [x] Record actual checks and hardware/runtime limitations below.

## Progress

Initial inspection: the non-macOS entry point exits immediately. Most application
files are gated by `darwin`; MyGo itself already supplies a Windows UI backend.
System integrations currently import the Darwin backend directly.

Implemented: shared platform entry point, a thin Darwin facade and Windows-native
keyboard hooks, WinMM PCM recording, input injection and endpoint-volume control.
The audio worker waits for native events and uses three reusable 100 ms buffers;
it drains them in queue order, including the partial tail after reset. Keyboard
hook lives on a thread of its own with its own message loop; it hands each key to
the main thread, where the handlers run as on macOS, and waits at most 200 ms
for the answer, so that a busy UI thread neither stalls typing in every app nor
gets the hook removed (`LowLevelHooksTimeout`). A swallowed key with Alt or Win
held, and the modifiers released and restored around injected Ctrl+C / Ctrl+V,
are followed by an unassigned key (VK 0xE8), so that the app in front does not
take a lone Alt for its menu bar or a lone Win for the Start menu.
No dependencies were added. Screenshot selection starts a short-lived built-in
PowerShell/WinForms helper; it takes a snapshot before opening the overlay.

API keys are encrypted with user-scoped DPAPI before writing. Clipboard read/write
failures stop before input injection. OS-specific permissions and Dock settings are
hidden on Windows. Windows auto-updates are disabled until a feed is published.
The macOS release script already loads its signing key through the environment;
removing the hardcoded private-key path lets ordinary Windows builds succeed.

## Actual verification (2026-10-17)

- Windows `CGO_ENABLED=0 go test ./...`: passed. Includes hotkey normalization,
  injected-input ABI sizes, audio ownership/order, encrypted secret persistence
  and clipboard failure regression tests. HTTP/WebSocket clients use test servers.
- `go tool mygo vet .`: passed.
- Windows `go build` and `go tool mygo build -platform windows/amd64`: passed;
  generated `build/windows-amd64/Elir.exe` and the NSIS installer. NSIS was fetched
  from MyGo's official distribution. No signing key or publishing was required.
- `GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./...`: passed; macOS test
  binaries also cross-compiled with `go test -c` (not executed on Windows).
- Screenshot helper: PowerShell parser check and embedded C# compilation passed.
- Windows startup smoke check with a workspace-isolated APPDATA: process stayed
  alive, keyboard hook started, no error log; private memory about 88.6 MiB and
  0 ms CPU increase over a single 3-second idle sample. This is a short smoke
  observation, not a performance benchmark or a visual UI inspection.
- Independent reviewer followed the ponytail-review skill. Clipboard failure and
  Windows privilege-message findings were fixed and re-reviewed. Verdict: Ship.
- `git diff --check`: passed.

The sandbox blocks test loopback sockets and some replacement operations in its
default temp directory. Full tests were rerun outside those restrictions; the
reviewer's independent tests also passed with a workspace temporary directory.

## Remaining manual verification

Actual microphone audio and device unplug/reconnect, paste into standard/elevated
applications, screenshot interaction on mixed-DPI/multiple monitors, and macOS
release execution were not tested. No external speech/LLM requests were made.
Windows PowerShell or WinForms disabled by policy prevents screenshot capture.
This first verified build targets Windows amd64; arm64 has not been run on hardware.

The PowerShell limitation above describes the 2026-10-17 implementation. The
2026-10-18 native screenshot migration replaces that path; see the follow-up below.

## Tray interaction follow-up (2026-10-18)

Windows left-click restores the existing settings window through `showSettings`,
retaining the last page and reloading feature settings as other open actions do.
Right-click keeps the tray menu; macOS keeps its existing menu click behavior.
Register the listener only after successful tray creation, using the pinned
MyGo v0.5.5 `Tray.OnClick` API. No new window or additional event infrastructure.

Verification: `go tool mygo vet .` and a Windows build passed. A native
click-message smoke check launched an isolated `ElirTraySmoke.exe` with workspace
APPDATA, closed Settings, sent MyGo's left-click tray callback message and verified
that the same settings window became visible. The right-click message entered
native menu mode. The helper targeted only its own process and closed it afterward;
it did not use the clipboard, microphone or speech services. This checks the real
message dispatch and window behavior, not a physical taskbar click or focus policy.
Independent ponytail-review: Ship. Windows `CGO_ENABLED=0 go test ./... -count=1`
passed, and macOS arm64/amd64 `go build ./...` passed. The Windows executable and
NSIS installer were rebuilt with `go tool mygo build -platform windows/amd64`.
A physical taskbar click and macOS runtime interaction were not exercised.

## Native screenshot migration (2026-10-18)

Branch `codex/windows-native-screenshot` preserves the existing screenshot library,
clipboard and thumbnail flow, replacing the PowerShell/WinForms helper with
in-process GDI capture and a Win32 selection window. Thread-local per-monitor DPI,
physical coordinates, DWM visible window bounds, context cancellation, native
cleanup and a non-overwriting PNG commit belong to `internal/platform`.
No screenshot dependency, CGO, helper process or fallback backend was added.

Current evidence and outstanding desktop acceptance are tracked in
[windows-native-screenshot.md](windows-native-screenshot.md#10-交接状态).
Historical verification above remains scoped to the earlier implementation.
