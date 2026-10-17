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
- Reject writes after a secret file read/decryption/JSON error. Publish changed
  keys in memory only after the file replacement succeeds. Keep existing DPAPI
  ciphertext and macOS JSON formats; do not reset or migrate damaged files.

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
