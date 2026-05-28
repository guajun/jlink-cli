# jlink-cli

`jlink-cli` is a Go-first, binary-distributed command-line interface for AI agents that need predictable access to SEGGER J-Link tooling.

The first prototype is intentionally non-invasive: `doctor` only inspects local executable paths and `run` only parses an agent request. Commands that open or occupy a physical J-Link session should be added behind explicit flags and tested only with user approval.

## Goals

- Single-file binaries for Windows, Linux, and macOS.
- Stable JSON output for AI agents.
- stdout for machine-readable results, stderr for diagnostics and errors.
- No interactive prompts by default.
- Clear diagnostics for J-Link executable discovery.

## Install from source

```powershell
go install github.com/guajun/jlink-cli/cmd/jlink-cli@latest
```

For local development:

```powershell
go test ./...
go run ./cmd/jlink-cli version --json
go run ./cmd/jlink-cli doctor --json
```

## Commands

```powershell
jlink-cli version --json
jlink-cli doctor --json
jlink-cli inspect --json
jlink-cli connect --device STM32H750VB --interface SWD --speed 4000 --dry-run
jlink-cli run --input '{"action":"ping"}'
```

`doctor` searches, in order, explicit flags, environment variables, `PATH`, and common SEGGER install directories. It does not connect to an attached device.

`connect` prepares a J-Link Commander session using a short script containing `connect` and `q`. It defaults to dry-run behavior unless `--yes` is passed, so agents can inspect the exact command before opening an exclusive physical probe session.

## Design Notes

This project is modeled after the useful architecture patterns in `square/pylink`: command separation, structured errors, platform-aware J-Link discovery, installation troubleshooting, and broad test coverage. The implementation is Go-first because single-binary distribution is a better fit for CLI tools used by agents.
