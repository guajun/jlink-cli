---
name: jlink-cli
description: "Discover the J-Link environment, generate Commander scripts, and run structured J-Link workflows with jlink-cli."
---

# jlink-cli

Use the installed binary or this repository's CLI for structured J-Link workflows. SEGGER J-Link software is a separate dependency.

## Resolve target settings

There is no default chip or flash address. Determine the exact J-Link device name from the user's request, board/project configuration, or device documentation. Ask for missing information that cannot be established from those sources. Probe enumeration identifies probes, not the attached MCU model.

Use the project's interface, supported speed, memory map, image placement, and access width. Do not assume SWD, 4000 kHz, a particular RAM/flash base, or an ARM architecture applies to every target. The CLI retains SWD and 4000 kHz as compatibility defaults; override them explicitly for the actual board. All target script commands require `--device`. Resolve flash placement using the format-specific rules below.

The examples below use PowerShell variables populated from the actual configuration: `$Device`, `$Interface`, `$Speed`, `$FlashAddress`, `$MemoryAddress`, `$BreakpointAddress`, `$Image`, and `$Serial`. They are placeholders, not predefined values. Adapt syntax to the user's shell.

## Resolve flash placement

- For ELF, HEX, MOT, S, S19, S37, and SREC images, use the addresses embedded in the file. Omit `--address`; Commander does not use it to relocate these formats. Inspect image sections/segments when placement needs validation.
- For BIN, first inspect the relevant project's linker script/scatter file, partition table, map file, and build/download configuration. Use the load/program address for the specific image, not a RAM execution address or ELF entry point. Account for bootloader/application offsets, image packaging, and external flash. If there is one unambiguous placement, pass that value as `--address` and report its source; do not ask the user to repeat information already available.
- If only the exact chip is known, consult its documented programmable memory regions (and board configuration for external flash) to narrow down candidates. Explain each candidate and its source. A flash-region base alone does not establish this image's application offset; do not silently select it when several placements remain possible.
- Ask only for unresolved placement information or a choice between supported candidates. Conflicting configurations must be resolved rather than taking the first match. Explicit user placement takes precedence for BIN; check it against the known memory map and image size.

Project and chip inference is performed by the agent using available files/documentation. The CLI does not automatically scan arbitrary projects or contain a chip-address database: it requires a resolved BIN address, accepts zero, and never falls back to a universal address.

## Execution and authorization

- `version`, `doctor`, and `inspect` inspect local information without opening a target session.
- `script probe` only generates a `ShowEmuList` script. It does not enumerate hardware and does not accept `--yes`.
- Target commands generate scripts by default. `--yes` opens a physical session and executes the operation.
- Use existing authorization when it covers the exact target and operation. If authorization is missing, present the exact command/script and its effects before asking. Flashing, writes, resets, halt/resume, and breakpoint changes can affect the active firmware.

Generate and inspect a plan before execution. Include `--yes` only for an authorized operation. Avoid repeating confirmation already given for that operation; ask again if the target or effects materially change.

## Check environment

Check that `jlink-cli` is available, then inspect its version and discover SEGGER Commander:

```sh
jlink-cli version --json
jlink-cli doctor --json
jlink-cli inspect --json
```

Discover the Commander executable with `doctor`; use `--jlink-path` for a nonstandard location. Do not assume a fixed Windows install directory. On Unix Commander is commonly named `JLinkExe`; on Windows distinguish SEGGER `JLink.exe` from the Java JDK's unrelated `jlink.exe`.

If a dependency is missing, consult the [installation instructions](https://github.com/guajun/jlink-cli/blob/main/README.md#install-from-source) and handle setup within the current task and existing authorization. The CLI binary, this skill, and SEGGER J-Link software are installed separately. Do not reinstall available dependencies or assume a source checkout is present.

## Generate and run

```powershell
jlink-cli script connect --device $Device --interface $Interface --speed $Speed --json
# Addressed image (ELF/HEX/S-record): use embedded addresses
jlink-cli script flash --device $Device --interface $Interface --speed $Speed --file build/app.elf --json
# Raw BIN: $FlashAddress has been resolved from the project or user input
jlink-cli script flash --device $Device --interface $Interface --speed $Speed --file build/app.bin --address $FlashAddress --json
jlink-cli memory read --device $Device --interface $Interface --speed $Speed --address $MemoryAddress --length 16 --width 8 --json
jlink-cli script breakpoint-set --device $Device --interface $Interface --speed $Speed --address $BreakpointAddress --json
```

The read length is a number of access units, not always bytes: width 32 and length 4 read 16 bytes. Choose an address, width, and count valid for the target. Live reads may affect memory-mapped peripherals; do not treat every address as passive RAM. `memory read --halt` emits halt/read/**resume** even if the target was already stopped. Use a custom Commander script when the target must remain halted. Breakpoint commands halt by default and resume only with `--run`.

`flash --verify` emits `verifybin` only for raw BIN images. Structured image formats do not get a separate `verifybin` command.

For an authorized execution, rerun the reviewed command with `--yes`, preserving the target and probe settings.

`run` accepts structured flash requests with `action: "flash"` and `arguments` containing `device` and `file` (strings), plus `address` for BIN. Omit `address` for addressed images. Optional arguments are `interface`, `speed`, `serial`, `jlink_path`, `timeout`, `verify`, and `reset`. It does not infer hardware or paths from natural language. Prefer the explicit flash command when preparing a request interactively.

## Probe selection

When probe identity is ambiguous, enumerate using Commander without connecting to a target. Set `$JLink` to the resolved SEGGER Commander executable:

```powershell
@("ShowEmuList", "q") | & $JLink -NoGui 1
```

On a POSIX shell, use `printf 'ShowEmuList\nq\n' | "$JLINK" -NoGui 1` with a resolved executable path. Use a serial-to-board mapping supplied by the user or established from the hardware; enumeration alone does not establish that mapping. Do not guess serial numbers or require enumeration when a reliable mapping is already known.

Pass `--serial $Serial` on every relevant command when selecting a specific probe. Execution passes `-SelectEmuBySN` as an executable argument, not a script command. Inspect `execution.command.args` when troubleshooting selection.

## Results

Use JSON for agent-readable CLI results. Report the generated script or executed operation, relevant target/probe settings, and errors or useful output. Preserve raw Commander output when diagnosing failures; do not infer success solely from the process having started.
