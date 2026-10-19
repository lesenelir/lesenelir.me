# Windows security follow-up

Separate commit after `2531932`. Scope: the Windows integrations added by that
commit and their shared secret-store caller. Single-user desktop application;
malware already running as the same Windows user is outside the DPAPI boundary.

## Changes and compatibility

- Resolve PowerShell and added Windows DLLs from the real system directory,
  without trusting PATH, the current directory or SystemRoot environment input.
  Load DLLs with `LOAD_LIBRARY_SEARCH_SYSTEM32` so their dependencies also use the
  system directory; retain the handles for the process lifetime.
  Keep the embedded screenshot script and existing command arguments.
  Restrict the helper's module search to system modules and omit API key variables.
- Exclude temporary auto-paste text from Windows clipboard history and cloud
  synchronization. Explicit user copy actions keep their current behavior.
- Reject writes after a secret file read error. A file read but not decrypted
  or decoded (DPAPI fails after a password reset or on another machine) is
  renamed to `secrets.json.unreadable-<time>`, unchanged, and the store starts
  empty, so keys can be entered again. Publish changed keys in memory only after
  the file replacement succeeds. Keep existing DPAPI ciphertext and macOS JSON formats.

No new dependencies or elevation. Review DLL/procedure callers, screenshot
launch, Paster and its tests, and Store/Secret callers before editing.

## Verification status

- [x] Read native-loading, screenshot, clipboard and secret-store call chains.
- [x] Add focused regression tests and run affected tests.
- [x] Run full Windows tests, MyGo vet, Windows build and macOS cross-compile.
- [x] Independent ponytail-review and final DLL-loader re-review: Ship. Commit
  this change separately from Windows feature support.

## Runtime limits

DPAPI protects stored keys for the Windows user, not against another process
running with that user's credentials. Clipboard exclusion requests prevent
Windows history/sync; other local clipboard readers can still observe clipboard
contents. Real device and multi-monitor interaction remain manual checks.

## Actual verification (2026-10-17)

- `CGO_ENABLED=0 go test ./...` and `go tool mygo vet .`: passed after the final
  DLL loading change. No real clipboard, recording or screenshot interaction.
- Tests cover actual system DLL paths under misleading environment variables,
  screenshot executable/module paths and stripped API-key environment variables,
  C# helper compilation, clipboard privacy marker, corrupt/decryption-failed secret
  files, and disk-write failure without an in-memory key change.
- Windows production executable and NSIS installer built successfully.
- macOS arm64 test binaries cross-compiled, including shared secret-store tests;
  these binaries were not executed on Windows.
- Independent ponytail-review checked the security boundaries and callers;
  no confirmed P1/P2 findings. Its DLL dependency-search observation was also
  addressed with the explicit system-directory loading flag before commit.
- Full Windows tests ran outside the sandbox's temp-file and loopback restrictions.

References: [Microsoft DLL security guidance](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-security)
and [Windows clipboard history/cloud exclusion formats](https://learn.microsoft.com/en-us/windows/win32/dataxchg/clipboard-formats#cloud-clipboard-and-clipboard-history-formats).

## Native screenshot migration (2026-10-18)

The screenshot helper, script embedding, C# declarations and helper-specific tests
have been removed on `codex/windows-native-screenshot`. Capture now runs entirely
inside Elir, so screenshot capture no longer starts a child process or constructs
a child environment. The earlier helper checks above are historical evidence.

New user32/gdi32/dwmapi calls retain the real-System32 loader and dependency-search
restriction. The optional old-OS shcore probe uses the same loader through an
error-returning variant, releases its reference, and fails only the screenshot
feature when unavailable. PNG output is encoded into an owned temporary file and
committed with MoveFileExW without REPLACE_EXISTING; cancellation or failure before
commit removes only the temporary file. Existing targets and the explicit-copy
clipboard policy are preserved. Resource and file-contract tests are recorded in
[the native screenshot design](windows-native-screenshot.md#10-交接状态).
