---
name: jlink-commander
description: "Use when: generating or running SEGGER J-Link Commander / JLinkExe scripts for AI-agent-safe probe diagnostics, target connect, memory read, flash programming, breakpoint control, or MCU debug automation."
---

# J-Link Commander Skill

Use this skill to plan, generate, dry-run, and optionally execute SEGGER J-Link Commander (`JLink.exe`, `JLinkExe`) scripts.

## Safety Model

Classify every requested operation before running tools.

- Probe-only operations: May run without asking the user. These can open the J-Link probe but must not connect to the target MCU. Examples: `ShowEmuList`, `ShowHWStatus`, `ShowFWInfo`, `USB`, `q`.
- Target-session operations: Ask the user before running. These connect to or control the MCU. Examples: `Device`, `SelectInterface`, `Speed`, `connect`, `mem8`, `mem16`, `mem32`, `regs`, `halt`, `go`, `SetBP`, `ClearBP`.
- Destructive operations: Ask the user with an explicit warning before running. Examples: `loadfile`, `erase`, `write`, `wreg`, `Power Off`, flash programming, reset when it may affect an active workflow.

Always show the exact command or script before any target-session or destructive execution.

## Preferred Execution Pattern

Prefer a temporary CommanderScript over a long interactive session.

```powershell
& "C:\Program Files\SEGGER\JLink\JLink.exe" -NoGui 1 -CommanderScript .\script.jlink
```

Use stdin piping only for very small probe-only experiments:

```powershell
@("ShowEmuList", "q") | & "C:\Program Files\SEGGER\JLink\JLink.exe" -NoGui 1
```

## Script Template

Start scripts with `ExitOnError 1` unless there is a specific reason not to.

Probe-only script:

```text
ExitOnError 1
ShowEmuList
q
```

Target connection script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
q
```

CLI generation:

```powershell
jlink-cli script connect --json
jlink-cli script flash --file build/app.elf --address 0x08000000 --json
jlink-cli script memory-read --address 0x20000000 --length 16 --width 8 --json
jlink-cli script breakpoint-set --address 0x08000100 --json
```

Only run target-session or destructive scripts after confirmation:

```powershell
jlink-cli script flash --file build/app.elf --address 0x08000000 --yes --json
```

Live memory read script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
mem8 0x20000000, 0x10
q
```

Halt/read/resume script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
halt
mem32 0x20000000, 0x4
go
q
```

Flash programming script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
loadfile firmware.elf, 0x08000000
q
```

Use `verifybin` only for raw `.bin` files. Do not generate `verifybin` for `.elf`, `.hex`, `.srec`, or `.mot`; J-Link Commander treats `verifybin` input as a raw binary image.

## Output Parsing Notes

J-Link Commander output includes a banner, prompt echoes, and script status lines. Do not assume stdout is machine-readable.

Useful stable fragments observed on Windows V8.66:

- Banner: `SEGGER J-Link Commander V...`
- Probe connection: `Connecting to J-Link via USB...O.K.`
- Serial: `S/N: <serial>`
- VTref: `VTref=<voltage>V`
- Probe list: `J-Link[0]: Connection: USB, Serial number: ..., ProductName: ...`
- Script mode: `J-Link Command File read successfully.` and `Script processing completed.`

For CLI implementation, preserve raw stdout/stderr and add parsed summaries only when patterns are stable.

## Defaults

Use these defaults unless the user says otherwise:

- Device: `STM32H750VB`
- Interface: `SWD`
- Speed: `4000`
- Windows executable: `C:\Program Files\SEGGER\JLink\JLink.exe`

## Agent Response Checklist

When generating a script, include:

- Operation classification: probe-only, target-session, or destructive.
- Exact script content.
- Exact execution command.
- Whether user confirmation is required before running.
- Parsed highlights after execution, plus the raw output location if output is long.---
name: jlink-commander
description: "Use when: generating or running SEGGER J-Link Commander / JLinkExe scripts for AI-agent-safe probe diagnostics, target connect, memory read, flash programming, breakpoint control, or MCU debug automation."
---

# J-Link Commander Skill

Use this skill to plan, generate, dry-run, and optionally execute SEGGER J-Link Commander (`JLink.exe`, `JLinkExe`) scripts.

## Safety Model

Classify every requested operation before running tools.

- Probe-only operations: May run without asking the user. These can open the J-Link probe but must not connect to the target MCU. Examples: `ShowEmuList`, `ShowHWStatus`, `ShowFWInfo`, `USB`, `q`.
- Target-session operations: Ask the user before running. These connect to or control the MCU. Examples: `Device`, `SelectInterface`, `Speed`, `connect`, `mem8`, `mem16`, `mem32`, `regs`, `halt`, `go`, `SetBP`, `ClearBP`.
- Destructive operations: Ask the user with an explicit warning before running. Examples: `loadfile`, `erase`, `write`, `wreg`, `Power Off`, flash programming, reset when it may affect an active workflow.

Always show the exact command or script before any target-session or destructive execution.

## Preferred Execution Pattern

Prefer a temporary CommanderScript over a long interactive session.

```powershell
& "C:\Program Files\SEGGER\JLink\JLink.exe" -NoGui 1 -CommanderScript .\script.jlink
```

Use stdin piping only for very small probe-only experiments:

```powershell
@("ShowEmuList", "q") | & "C:\Program Files\SEGGER\JLink\JLink.exe" -NoGui 1
```

## Script Template

Start scripts with `ExitOnError 1` unless there is a specific reason not to.

Probe-only script:

```text
ExitOnError 1
ShowEmuList
q
```

Target connection script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
q
```

CLI generation:

```powershell
jlink-cli script connect --json
jlink-cli script flash --file build/app.elf --address 0x08000000 --json
jlink-cli script memory-read --address 0x20000000 --length 16 --width 8 --json
jlink-cli script breakpoint-set --address 0x08000100 --json
```

Only run target-session or destructive scripts after confirmation:

```powershell
jlink-cli script flash --file build/app.elf --address 0x08000000 --yes --json
```

Live memory read script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
mem8 0x20000000, 0x10
q
```

Halt/read/resume script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
halt
mem32 0x20000000, 0x4
go
q
```

Flash programming script:

```text
ExitOnError 1
Device STM32H750VB
SelectInterface SWD
Speed 4000
connect
loadfile firmware.elf, 0x08000000
q
```

Use `verifybin` only for raw `.bin` files. Do not generate `verifybin` for `.elf`, `.hex`, `.srec`, or `.mot`; J-Link Commander treats `verifybin` input as a raw binary image.

## Output Parsing Notes

J-Link Commander output includes a banner, prompt echoes, and script status lines. Do not assume stdout is machine-readable.

Useful stable fragments observed on Windows V8.66:

- Banner: `SEGGER J-Link Commander V...`
- Probe connection: `Connecting to J-Link via USB...O.K.`
- Serial: `S/N: <serial>`
- VTref: `VTref=<voltage>V`
- Probe list: `J-Link[0]: Connection: USB, Serial number: ..., ProductName: ...`
- Script mode: `J-Link Command File read successfully.` and `Script processing completed.`

For CLI implementation, preserve raw stdout/stderr and add parsed summaries only when patterns are stable.

## Defaults

Use these defaults unless the user says otherwise:

- Device: `STM32H750VB`
- Interface: `SWD`
- Speed: `4000`
- Windows executable: `C:\Program Files\SEGGER\JLink\JLink.exe`

## Agent Response Checklist

When generating a script, include:

- Operation classification: probe-only, target-session, or destructive.
- Exact script content.
- Exact execution command.
- Whether user confirmation is required before running.
- Parsed highlights after execution, plus the raw output location if output is long.
