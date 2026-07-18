# Changelog

All notable changes to this project will be documented in this file.

## 0.2.0

- Add `--serial` probe selection to memory, breakpoint, flash, and target control commands.
- Update bundled agent skills with multi-probe workflows for GitHub Copilot, Claude Code, and Codex.

## 0.1.0

- Bootstrap Go-first `jlink-cli` prototype.
- Add JSON-first `version`, `doctor`, `inspect`, and prototype `run` commands.
- Add non-exclusive J-Link executable discovery for local diagnostics.
- Add explicit `connect` command for short J-Link Commander connection checks.
- Validate initial connection flow with STM32H750VB over SWD.
- Add Go tests and CI workflow.
