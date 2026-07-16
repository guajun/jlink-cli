---
name: jlink-cli
description: "Use when: installing, discovering, generating scripts with, or safely running the jlink-cli tool for SEGGER J-Link workflows."
---

# jlink-cli Skill

Use this skill when working with the `jlink-cli` command-line tool in this repository or from an installed binary.

## Safety Model

Treat `jlink-cli` commands according to what they do:

- Discovery commands: May run without confirmation. Examples: `jlink-cli version --json`, `jlink-cli doctor --json`, `jlink-cli inspect --json`, `jlink-cli script probe --json`.
- Script generation commands: May run without confirmation because they do not touch the target unless `--yes` is present. Examples: `jlink-cli script connect --json`, `jlink-cli script flash --file build/app.elf --json`, `jlink-cli memory read ...` without `--yes`.
- Target-session commands: Ask before running with `--yes`. Examples: `connect --yes`, `script connect --yes`, `memory read --yes`, breakpoint changes, `target halt --yes`, and `target reset --yes`.
- Destructive commands: Ask with an explicit warning before running with `--yes`. Examples: `flash --yes`, `script flash --yes`, erase/write-style J-Link operations, and reset when it may affect an active workflow.

Always show the exact command before any target-session or destructive execution.

## Install and Skill Setup

Install the latest local source build:

```powershell
go install ./cmd/jlink-cli
```

Install the bundled skills for local agent hosts:

```powershell
jlink-cli skill install --skill all --agent copilot --json
jlink-cli skill install --skill all --agent claude-code --json
jlink-cli skill install --skill all --agent codex --json
```

Default local skill locations:

- GitHub Copilot: `~/.copilot/skills/<skill-name>/SKILL.md`
- Claude Code: `~/.claude/skills/<skill-name>/SKILL.md`
- Codex: `~/.codex/skills/<skill-name>/SKILL.md`

Use GitHub CLI when provenance, updates, pinning, or broader host support is needed:

```powershell
gh skill install guajun/jlink-cli jlink-cli --agent github-copilot --scope user
gh skill install guajun/jlink-cli jlink-cli --agent claude-code --scope user
gh skill install guajun/jlink-cli jlink-cli --agent codex --scope user
```

## Preferred Command Patterns

Use JSON output for agent-readable results:

```powershell
jlink-cli version --json
jlink-cli doctor --json
jlink-cli inspect --json
```

Generate scripts before executing them:

```powershell
jlink-cli script connect --json
jlink-cli script flash --file build/app.elf --address 0x08000000 --json
jlink-cli script memory-read --address 0x20000000 --length 16 --width 8 --json
jlink-cli script breakpoint-set --address 0x08000100 --json
```

Only execute target-session or destructive operations after confirmation:

```powershell
jlink-cli script flash --file build/app.elf --address 0x08000000 --yes --json
```

## Multi-Probe Workflow

Enumerate attached probes before connecting to any MCU. `jlink-cli script probe --json` generates the probe-only `ShowEmuList` script; to perform enumeration, run the same probe-only operation directly with J-Link Commander:

```powershell
@("ShowEmuList", "q") | & "C:\Program Files\SEGGER\JLink\JLink.exe" -NoGui 1
```

Map each reported serial number to its board, then pass `--serial <number>` to every target-session command that must use that probe:

```powershell
jlink-cli connect --device STM32H750VB --serial 123456789 --dry-run --json
jlink-cli memory read --device STM32H750VB --address 0x20000000 --length 16 --width 8 --serial 123456789 --json
jlink-cli script breakpoint-set --address 0x08000100 --serial 123456789 --json
jlink-cli script breakpoint-clear --address 0x08000100 --serial 123456789 --json
jlink-cli target halt --device STM32H750VB --serial 123456789 --json
jlink-cli flash --device STM32H750VB --file firmware.bin --address 0x08000000 --serial 123456789 --json
```

The shared target commands support `--serial`: target `script` commands, `flash`, `memory read`, and `target halt|run|reset`. When `--yes` executes an operation, `execution.command.args` must include `-SelectEmuBySN` followed by the selected serial number. The selection is a J-Link executable argument, so it does not appear in the generated CommanderScript text.

Generate and inspect without `--yes` first. After confirmation, rerun the exact command with `--yes`; direct J-Link Commander fallback is not needed for serial-specific memory reads or breakpoint changes.

## Defaults

Unless the user says otherwise, use:

- Device: `STM32H750VB`
- Interface: `SWD`
- Speed: `4000`
- Flash address: `0x08000000`

Do not guess a serial number. Enumerate probes or use a mapping supplied by the user.

## Troubleshooting

If J-Link tools are not found, run:

```powershell
jlink-cli doctor --json
```

If the command is unavailable after `go install`, check that `$(go env GOPATH)\bin` is on `PATH`.

If a serial-specific execution selects the wrong probe or no probe, inspect `execution.command.args` and verify the value reported by `ShowEmuList`.

## Agent Response Checklist

When helping with `jlink-cli`, include:

- Whether the requested command is discovery, script generation, target-session, or destructive.
- The exact command to run, including `--serial` in multi-probe setups.
- Whether `--yes` is needed and why.
- The expected JSON result shape or important fields.
- Any installed skill path when using `jlink-cli skill install`.