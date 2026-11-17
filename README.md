# jlink-cli

`jlink-cli` is a Go-first, binary-distributed command-line interface for AI agents that need predictable access to SEGGER J-Link tooling.

`doctor` only inspects local executable paths. Target operations generate a plan by default and execute only with `--yes`; hardware tests require authorization for the selected target and operation.

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
jlink-cli connect --device $Device --interface $Interface --speed $Speed --serial $Serial --dry-run
jlink-cli script probe --json
jlink-cli script connect --device $Device --interface $Interface --speed $Speed --json
jlink-cli script flash --device $Device --interface $Interface --speed $Speed --file build/app.elf --json
jlink-cli script memory-read --device $Device --interface $Interface --speed $Speed --address $MemoryAddress --length 16 --width 8 --serial $Serial --json
jlink-cli script breakpoint-set --device $Device --interface $Interface --speed $Speed --address $BreakpointAddress --serial $Serial --json
jlink-cli flash --device $Device --interface $Interface --speed $Speed --file build/app.bin --address $FlashAddress --verify
jlink-cli memory read --device $Device --interface $Interface --speed $Speed --address $MemoryAddress --length 16 --width 8 --serial $Serial
jlink-cli memory read --device $Device --interface $Interface --speed $Speed --address $MemoryAddress --length 16 --width 8 --halt
jlink-cli target halt --device $Device --interface $Interface --speed $Speed
jlink-cli target run --device $Device --interface $Interface --speed $Speed
jlink-cli target reset --device $Device --interface $Interface --speed $Speed
jlink-cli run --input '{"action":"ping"}'
```

The PowerShell variables above are placeholders to populate from the actual board/project configuration. There is no default chip or flash address: all target script commands require `--device`, and raw BIN flashing requires a resolved `--address` (zero is valid). ELF, HEX, MOT, S, S19, S37, and SREC use embedded locations, so omit `--address`. A supplied address is ignored for these formats and reported as `flash.address_ignored`; it does not relocate the image. The CLI retains SWD and 4000 kHz as compatibility defaults; explicitly select the interface and speed appropriate to your board. `examples/connect-stm32h750vb.json` is a board-specific example, not a default configuration.

For BIN, the agent should first resolve placement from the relevant linker/scatter file, partition table, map file, or build/download configuration, distinguishing flash load addresses from RAM execution addresses. Account for application offsets, packaging, and external flash. Report the configuration used as evidence. If the project cannot establish the location, use the exact chip's documented flash regions and board configuration to offer supported candidates; ask only when the intended placement remains unknown or ambiguous. The CLI itself does not scan project files or maintain a chip-address database, and never substitutes a universal flash base.

Address handling follows [SEGGER's LoadFile documentation](https://kb.segger.com/J-Link_Commander#LoadFile): the optional address applies to BIN, while addressed formats use their own locations.

`doctor` searches, in order, explicit flags, environment variables, common SEGGER install directories, and `PATH`. It does not connect to an attached device.

`connect` prepares a J-Link Commander session using a short script containing `connect` and `q`. It defaults to dry-run behavior unless `--yes` is passed, so agents can inspect the exact command before opening an exclusive physical probe session.

`script` generates J-Link CommanderScript text without executing it by default. `script probe` is probe-only, while target scripts such as `script connect`, `script flash`, `script memory-read`, and `script breakpoint-set` are marked as requiring confirmation. Passing `--yes` executes the generated script through `JLink.exe -NoGui 1 -CommanderScript <temp-script>`.

`flash`, `memory read`, and `target halt|run|reset` are direct aliases for the same script execution engine. They return the generated script as JSON by default and execute only when `--yes` is passed. `flash program` is kept as a compatibility spelling, but `flash --file ...` is the preferred form.

All target-session commands accept `--serial <number>` to select one probe in a multi-J-Link setup. This includes `connect`, target `script` commands, `flash`, `memory read`, and `target halt|run|reset`. On execution, the CLI passes the selection to J-Link Commander as `-SelectEmuBySN <number>`; the generated CommanderScript remains unchanged.

`flash --verify` emits `verifybin` only for raw `.bin` images. ELF, HEX, SREC, and MOT files are loaded with `loadfile` without `verifybin`, because J-Link Commander verifies raw binaries differently from structured image formats.

`memory read` has two modes. The default live mode does not issue `h` before reading; it is intended to approximate non-blocking reads for addresses that J-Link can access while the MCU is running. Passing `--halt` emits a halt/read/resume script, including when the target was already stopped. Use a custom Commander script if it must remain halted. Read lengths count access units: width 32 and length 4 read 16 bytes. Reads from peripheral registers may have side effects.

`flash program` and `breakpoint` are potentially destructive or exclusive operations. They also default to dry-run behavior and require `--yes` before the target is touched.

`run` accepts an explicit structured flash request and shares the `flash` validation/execution path. Natural-language flash requests are rejected rather than assigned a chip, address, or probe:

```json
{
  "action": "flash",
  "arguments": {
    "device": "<project J-Link device name>",
    "file": "build/app.elf",
    "interface": "<board interface>",
    "speed": "<supported speed>",
    "serial": "<selected probe serial>"
  }
}
```

Replace the placeholders before use. `device` and `file` are required strings. BIN also requires an `address` string resolved from the project or user input; addressed images omit it. Optional settings are `interface`, `speed`, `serial`, `jlink_path`, `timeout`, `verify`, and `reset`. Only an explicit `run --yes` executes the plan. Other prototype requests such as `ping` are parsed without hardware access.

## Agent skills (install separately)

The CLI binary and Agent Skills are installed separately. The one-line installers above install only the CLI binary and do not include skills. `gh skill install` installs skill files only and does not install the CLI binary. Use a GitHub CLI version with `gh skill install` support.

The built-in `jlink-cli skill install` command has been removed. Use GitHub CLI instead:

```powershell
gh skill install guajun/jlink-cli jlink-cli --agent github-copilot --scope user
gh skill install guajun/jlink-cli jlink-cli --agent claude-code --scope user
gh skill install guajun/jlink-cli jlink-cli --agent codex --scope user
```

## Design Notes

This project is modeled after the useful architecture patterns in `square/pylink`: command separation, structured errors, platform-aware J-Link discovery, installation troubleshooting, and broad test coverage. The implementation is Go-first because single-binary distribution is a better fit for CLI tools used by agents.

The `jlink-cli` skill teaches agents to use this CLI. For direct SEGGER J-Link Commander workflows, replace the skill name with `jlink-commander` in the commands above. SEGGER J-Link software must also be installed separately.
