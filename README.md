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
jlink-cli flash program --device STM32H750VB --file firmware.bin --address 0x08000000 --verify --dry-run
jlink-cli memory read --device STM32H750VB --address 0x20000000 --length 16 --width 8 --dry-run
jlink-cli memory read --device STM32H750VB --address 0x20000000 --length 16 --width 8 --halt --dry-run
jlink-cli breakpoint set --device STM32H750VB --address 0x08000100 --dry-run
jlink-cli callstack --device STM32H750VB --elf firmware.elf --dry-run
jlink-cli run --input '{"action":"ping"}'
```

`doctor` searches, in order, explicit flags, environment variables, `PATH`, and common SEGGER install directories. It does not connect to an attached device.

`connect` prepares a J-Link Commander session using a short script containing `connect` and `q`. It defaults to dry-run behavior unless `--yes` is passed, so agents can inspect the exact command before opening an exclusive physical probe session.

`memory read` has two modes. The default live mode does not issue `h` before reading; it is intended to approximate non-blocking reads for addresses that J-Link can access while the MCU is running. Passing `--halt` emits a halt/read/resume script for the stopped-at-breakpoint case.

`flash program` and `breakpoint` are potentially destructive or exclusive operations. They also default to dry-run behavior and require `--yes` before the target is touched.

`callstack` prepares a GDB Server plus `arm-none-eabi-gdb` backtrace command for the case where the MCU is already halted at a breakpoint. It defaults to dry-run and requires `--yes` before starting `JLinkGDBServerCL` and GDB.

## Design Notes

This project is modeled after the useful architecture patterns in `square/pylink`: command separation, structured errors, platform-aware J-Link discovery, installation troubleshooting, and broad test coverage. The implementation is Go-first because single-binary distribution is a better fit for CLI tools used by agents.
