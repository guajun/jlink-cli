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
- Target-session commands: Ask before running with `--yes`. Examples: `connect --yes`, `script connect --yes`, `memory read --yes`, `target halt --yes`, `target reset --yes`.
- Destructive commands: Ask with an explicit warning before running with `--yes`. Examples: `flash --yes`, `script flash --yes`, breakpoint changes, erase/write-style J-Link operations.

Always show the exact command before any target-session or destructive execution.

## Install and Skill Setup

Install the latest local source build:

```powershell
go install ./cmd/jlink-cli
```

Install this skill for a local agent host:

```powershell
jlink-cli skill install --agent copilot --json
jlink-cli skill install --agent claude-code --json
jlink-cli skill install --agent codex --json
```

Default local skill locations:

- GitHub Copilot: `~/.copilot/skills/jlink-cli/SKILL.md`
- Claude Code: `~/.claude/skills/jlink-cli/SKILL.md`
- Codex: `~/.codex/skills/jlink-cli/SKILL.md`

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

## Defaults

Unless the user says otherwise, use:

- Device: `STM32H750VB`
- Interface: `SWD`
- Speed: `4000`
- Flash address: `0x08000000`

## Troubleshooting

If J-Link tools are not found, run:

```powershell
jlink-cli doctor --json
```

If the command is unavailable after `go install`, check that `$(go env GOPATH)\bin` is on `PATH`.

## Agent Response Checklist

When helping with `jlink-cli`, include:

- Whether the requested command is discovery, script generation, target-session, or destructive.
- The exact command to run.
- Whether `--yes` is needed and why.
- The expected JSON result shape or important fields.
- Any installed skill path when using `jlink-cli skill install`.---
name: jlink-cli
description: "Use when: installing, discovering, generating scripts with, or safely running the jlink-cli tool for SEGGER J-Link workflows."
---

# jlink-cli Skill

Use this skill when working with the `jlink-cli` command-line tool in this repository or from an installed binary.

## Safety Model

Treat `jlink-cli` commands according to what they do:

- Discovery commands: May run without confirmation. Examples: `jlink-cli version --json`, `jlink-cli doctor --json`, `jlink-cli inspect --json`, `jlink-cli script probe --json`.
- Script generation commands: May run without confirmation because they do not touch the target unless `--yes` is present. Examples: `jlink-cli script connect --json`, `jlink-cli script flash --file build/app.elf --json`, `jlink-cli memory read ...` without `--yes`.
- Target-session commands: Ask before running with `--yes`. Examples: `connect --yes`, `script connect --yes`, `memory read --yes`, `target halt --yes`, `target reset --yes`.
- Destructive commands: Ask with an explicit warning before running with `--yes`. Examples: `flash --yes`, `script flash --yes`, breakpoint changes, erase/write-style J-Link operations.

Always show the exact command before any target-session or destructive execution.

## Install and Skill Setup

Install the latest local source build:

```powershell
go install ./cmd/jlink-cli
```

Install this skill for a local agent host:

```powershell
jlink-cli skill install --agent copilot --json
jlink-cli skill install --agent claude-code --json
jlink-cli skill install --agent codex --json
```

Default local skill locations:

- GitHub Copilot: `~/.copilot/skills/jlink-cli/SKILL.md`
- Claude Code: `~/.claude/skills/jlink-cli/SKILL.md`
- Codex: `~/.codex/skills/jlink-cli/SKILL.md`

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

## Defaults

Unless the user says otherwise, use:

- Device: `STM32H750VB`
- Interface: `SWD`
- Speed: `4000`
- Flash address: `0x08000000`

## Troubleshooting

If J-Link tools are not found, run:

```powershell
jlink-cli doctor --json
```

If the command is unavailable after `go install`, check that `$(go env GOPATH)\bin` is on `PATH`.

## Agent Response Checklist

When helping with `jlink-cli`, include:

- Whether the requested command is discovery, script generation, target-session, or destructive.
- The exact command to run.
- Whether `--yes` is needed and why.
- The expected JSON result shape or important fields.
- Any installed skill path when using `jlink-cli skill install`.