---
name: jlink-commander
description: "Generate or run SEGGER J-Link Commander / JLinkExe scripts for probe diagnostics, target connection, memory access, flash programming, or breakpoint control."
---

# J-Link Commander

Prefer `jlink-cli` for supported workflows when it is available. Use direct Commander when needed for unsupported operations or custom scripts.

## Resolve the environment and target

Find the installed SEGGER Commander through `jlink-cli doctor --json`, host configuration, or known SEGGER installation locations. Use the resolved executable path: Windows commonly uses `JLink.exe`, Linux/macOS `JLinkExe`. Do not assume a fixed directory or confuse it with Java's `jlink.exe`.

There is no default chip, interface, speed, memory map, architecture, or flash address for this skill. Use the user's target and project configuration. Probe enumeration cannot identify the attached MCU model. Resolve missing settings before generating a target script; ask only when they cannot be established from available context.

## Resolve flash placement

- For ELF, HEX, MOT, S, S19, S37, and SREC images, use the addresses embedded in the file. Omit `--address`; Commander does not use it to relocate these formats. Inspect image sections/segments when placement needs validation.
- For BIN, first inspect the relevant project's linker script/scatter file, partition table, map file, and build/download configuration. Use the load/program address for the specific image, not a RAM execution address or ELF entry point. Account for bootloader/application offsets, image packaging, and external flash. If there is one unambiguous placement, pass that value as `--address` and report its source; do not ask the user to repeat information already available.
- If only the exact chip is known, consult its documented programmable memory regions (and board configuration for external flash) to narrow down candidates. Explain each candidate and its source. A flash-region base alone does not establish this image's application offset; do not silently select it when several placements remain possible.
- Ask only for unresolved placement information or a choice between supported candidates. Conflicting configurations must be resolved rather than taking the first match. Explicit user placement takes precedence for BIN; check it against the known memory map and image size.

Project and chip inference is performed by the agent using available files/documentation. The CLI does not automatically scan arbitrary projects or contain a chip-address database: it requires a resolved BIN address, accepts zero, and never falls back to a universal address.

## Effects and authorization

Probe enumeration with `ShowEmuList` opens a probe-only workflow. Target connection, memory access, halt/resume, breakpoints, flashing, writes, and resets can affect the target. Memory-mapped peripheral reads may have side effects.

Generate and inspect scripts before execution. Use existing authorization for the requested target and effects. When it is absent, show the exact script/command and describe the effects before asking. Do not require repeated confirmation for an already authorized operation.

Start scripts with `ExitOnError 1` unless the workflow needs otherwise. Prefer temporary CommanderScript files for reproducible operations.

## Probe selection

Probe-only enumeration script:

```text
ExitOnError 1
ShowEmuList
q
```

Establish the serial-to-board mapping from user input or hardware evidence. Use a known mapping directly; enumerate when selection is ambiguous. Do not guess a serial number or assume enumeration identifies the MCU.

With PowerShell variables `$JLink`, `$Serial`, and `$ScriptPath` populated from the actual environment:

```powershell
& $JLink -NoGui 1 -SelectEmuBySN $Serial -CommanderScript $ScriptPath
```

POSIX shell equivalent:

```sh
"$JLINK" -NoGui 1 -SelectEmuBySN "$SERIAL" -CommanderScript "$SCRIPT_PATH"
```

Omit the serial argument when no specific probe selection is needed. `SelectEmuBySN` is an executable argument, not a CommanderScript command.

## Target scripts

The angle-bracket values below are template placeholders. Replace them from the actual target configuration before execution; do not pass the placeholders literally to Commander.

```text
ExitOnError 1
Device <J-Link-device-name>
SelectInterface <target-interface>
Speed <supported-speed>
connect
q
```

Insert only the requested operation before `q`:

- Read memory: `mem8 <address>, <count>`, or `mem16` / `mem32` with the number of access units. Match width and count to the target region.
- Halt and read: insert `halt` before the read. Add `go` only when the requested final state is running; do not automatically resume a target stopped at a breakpoint.
- Program an addressed image: `loadfile <firmware.elf>` (also HEX/S-record); use its embedded locations. Program raw BIN with `loadfile <firmware.bin>, <resolved-program-address>`. Resolve the BIN placement from the project first, then documented chip/board candidates, asking only if it remains ambiguous.
- Verify a raw BIN: `verifybin <binary-path>, <program-address>`. Do not apply `verifybin` to ELF, HEX, SREC, or MOT images; it treats input as raw bytes.

For CLI generation, pass explicit target settings. These PowerShell variables must be populated from the project, not example hardware:

```powershell
jlink-cli script connect --device $Device --interface $Interface --speed $Speed --json
jlink-cli script flash --device $Device --interface $Interface --speed $Speed --file build/app.elf --json
jlink-cli script flash --device $Device --interface $Interface --speed $Speed --file build/app.bin --address $FlashAddress --json
jlink-cli script memory-read --device $Device --interface $Interface --speed $Speed --address $MemoryAddress --length 16 --width 8 --json
jlink-cli script breakpoint-set --device $Device --interface $Interface --speed $Speed --address $BreakpointAddress --json
```

Add `--serial $Serial` when selecting a probe. CLI commands generate plans by default; `--yes` executes. `memory-read --halt` always emits `go` after reading; use a custom script if the target must stay halted. Breakpoint operations halt by default and resume only with `--run`.

## Output

Commander stdout contains banners, prompt echoes, and script status; it is not a structured protocol. Preserve stdout/stderr. Interpret serial numbers, reference voltage, connection status, and errors in context: wording varies across versions, transports, and hosts. Report useful highlights and retain raw output for diagnostics.
