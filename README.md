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

## Install from release

Windows PowerShell:

```powershell
irm https://github.com/guajun/jlink-cli/releases/latest/download/install.ps1 | iex
```

Linux shell:

```sh
curl -fsSL https://github.com/guajun/jlink-cli/releases/latest/download/install.sh | sh
```

These installer scripts are published as GitHub Release assets and download the latest binary archive from the [GitHub Releases](https://github.com/guajun/jlink-cli/releases/latest) page.

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
jlink-cli connect --device STM32H750VB --interface SWD --speed 4000 --serial 123456789 --dry-run
jlink-cli script probe --json
jlink-cli script connect --json
jlink-cli script flash --file build/app.elf --address 0x08000000 --json
jlink-cli script memory-read --address 0x20000000 --length 16 --width 8 --serial 123456789 --json
jlink-cli script breakpoint-set --address 0x08000100 --serial 123456789 --json
jlink-cli flash --device STM32H750VB --file firmware.bin --address 0x08000000 --verify
jlink-cli memory read --device STM32H750VB --address 0x20000000 --length 16 --width 8 --serial 123456789
jlink-cli memory read --device STM32H750VB --address 0x20000000 --length 16 --width 8 --halt
jlink-cli target halt --device STM32H750VB
jlink-cli target run --device STM32H750VB
jlink-cli target reset --device STM32H750VB
jlink-cli skill install --json
jlink-cli skill install --skill jlink-commander --json
jlink-cli skill install --agent claude-code --json
jlink-cli skill install --agent codex --json
jlink-cli run --input '{"action":"ping"}'
```

`doctor` searches, in order, explicit flags, environment variables, `PATH`, and common SEGGER install directories. It does not connect to an attached device.

`connect` prepares a J-Link Commander session using a short script containing `connect` and `q`. It defaults to dry-run behavior unless `--yes` is passed, so agents can inspect the exact command before opening an exclusive physical probe session.

`script` generates J-Link CommanderScript text without executing it by default. `script probe` is probe-only, while target scripts such as `script connect`, `script flash`, `script memory-read`, and `script breakpoint-set` are marked as requiring confirmation. Passing `--yes` executes the generated script through `JLink.exe -NoGui 1 -CommanderScript <temp-script>`.

`flash`, `memory read`, and `target halt|run|reset` are direct aliases for the same script execution engine. They return the generated script as JSON by default and execute only when `--yes` is passed. `flash program` is kept as a compatibility spelling, but `flash --file ...` is the preferred form.

All target-session commands accept `--serial <number>` to select one probe in a multi-J-Link setup. This includes `connect`, target `script` commands, `flash`, `memory read`, and `target halt|run|reset`. On execution, the CLI passes the selection to J-Link Commander as `-SelectEmuBySN <number>`; the generated CommanderScript remains unchanged.

`flash --verify` emits `verifybin` only for raw `.bin` images. ELF, HEX, SREC, and MOT files are loaded with `loadfile` without `verifybin`, because J-Link Commander verifies raw binaries differently from structured image formats.

`memory read` has two modes. The default live mode does not issue `h` before reading; it is intended to approximate non-blocking reads for addresses that J-Link can access while the MCU is running. Passing `--halt` emits a halt/read/resume script for the stopped-at-breakpoint case.

`flash program` and `breakpoint` are potentially destructive or exclusive operations. They also default to dry-run behavior and require `--yes` before the target is touched.

`callstack` prepares a GDB Server plus `arm-none-eabi-gdb` backtrace command for the case where the MCU is already halted at a breakpoint. In this version it returns the command plan as JSON rather than starting the long-running server itself.

`skill install` installs bundled Agent Skills into local agent host directories. The default skill is `jlink-cli`, which teaches agents how to use this CLI. The separate `jlink-commander` skill is for direct SEGGER J-Link Commander / `JLinkExe` workflows. Use `--skill jlink-commander` to install only the Commander skill, or `--skill all` to install both. Existing installs are updated in place.

Supported local agent targets:

| Agent | `--agent` values | Default install path |
| --- | --- | --- |
| GitHub Copilot | `github-copilot`, `copilot` | `~/.copilot/skills/<skill-name>/SKILL.md` |
| Claude Code | `claude-code`, `claude` | `~/.claude/skills/<skill-name>/SKILL.md` |
| Codex | `codex` | `~/.codex/skills/<skill-name>/SKILL.md` |
| Custom path | `--dir <path>` | `<path>/<skill-name>/SKILL.md` |

The default `--agent all` installs to GitHub Copilot, Claude Code, and Codex. For broader agent-host support, version pinning, provenance metadata, and updates, use GitHub CLI's Agent Skills installer:

```powershell
gh skill install guajun/jlink-cli jlink-cli --agent github-copilot --scope user
gh skill install guajun/jlink-cli jlink-cli --agent claude-code --scope user
gh skill install guajun/jlink-cli jlink-cli --agent codex --scope user
```

## Design Notes

This project is modeled after the useful architecture patterns in `square/pylink`: command separation, structured errors, platform-aware J-Link discovery, installation troubleshooting, and broad test coverage. The implementation is Go-first because single-binary distribution is a better fit for CLI tools used by agents.
