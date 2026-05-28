# Changelog

All notable changes to this project will be documented in this file.

## 0.1.0-dev

- Bootstrap Go-first `jlink-cli` prototype.
- Add JSON-first `version`, `doctor`, `inspect`, and prototype `run` commands.
- Add non-exclusive J-Link executable discovery for local diagnostics.
- Add explicit `connect` command for short J-Link Commander connection checks.
- Add staged `flash program`, `memory read`, `breakpoint`, and `callstack` command surfaces.
- Separate live memory reads from halt/read/resume memory reads.
- Add `dll doctor` for non-invasive J-Link DLL loading and symbol probing.
- Validate initial connection flow with STM32H750VB over SWD.
- Add Go tests and CI workflow.
