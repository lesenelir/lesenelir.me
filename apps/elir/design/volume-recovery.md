# Windows startup volume recovery

Follow-up to the review of `2531932` and `59d07e7`.

## Contract

`Voice.Start -> quieter.recover -> RestoreOutput -> withOutput` runs during
MyGo's startup on a Windows STA thread. Normal recording callbacks use worker
threads, typically uninitialized or MTA. Support both apartments without changing
an existing apartment or uninitializing COM owned by the caller. Every successful
CoInitializeEx, including S_FALSE, must have one matching CoUninitialize.

When native restoration fails, retain the saved recovery file and in-memory
state. Retry before applying another recording-time volume change; do not
overwrite a pending recovery with another device/state. Remove a valid record
only after restoration succeeds. Keep the existing file schema and Darwin
backend; malformed JSON continues to be discarded as before.

## Verification plan

- [x] Confirm callers and MyGo v0.5.5 main-thread STA initialization.
- [x] Reproduce with a native STA regression, without changing audio devices.
- [x] Fix COM initialization and failed-recovery retention.
- [x] Test uninitialized, existing MTA and existing STA threads; check apartment
  preservation and balanced reference counts on failure and repeated calls.
- [x] Test retained records, retries and successful cleanup with fake audio APIs.
- [x] Run Windows suite/MyGo vet/build and compile macOS tests.
- [x] Independent ponytail-review before a separate fix commit: Ship.

## Actual verification

The new native COM regression failed on the original code for STA with
`CoInitializeEx: HRESULT 0x80010106`; uninitialized and MTA cases passed. After
the fix, all three cases passed. The invalid test endpoint prevents real volume
access; changing apartment after releasing the caller's own reference checks that
S_FALSE references did not leak. Shared recovery tests inject audio functions and
verify that failed restoration preserves the file and blocks replacement until a
retry succeeds. The old Darwin test that expected deletion after an invalid
device failure was replaced with shared deterministic success/failure tests.

`CGO_ENABLED=0 go test ./... -count=1` and `go tool mygo vet .` passed on Windows.
`go tool mygo build -platform windows/amd64` rebuilt the executable and NSIS
installer. All macOS test binaries cross-compiled for arm64 and amd64 after the
build completed; MyGo build's temporary `.syso` resource requires these steps to
run sequentially. The macOS binaries were not executed on Windows.

Independent ponytail-review rechecked initialization pairing, cleanup order and
recovery callers, and ran the new COM and quieter tests: Ship, no new findings.
`git diff --check` passed. Actual volume changes and a process-crash/restart with
real devices were not exercised.
