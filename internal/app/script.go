package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/guajun/jlink-cli/internal/jlink"
	"github.com/guajun/jlink-cli/internal/protocol"
	"github.com/guajun/jlink-cli/internal/scriptgen"
)

type scriptResult struct {
	Script    scriptgen.Script     `json:"script"`
	Execution *jlink.ConnectResult `json:"execution,omitempty"`
}

type scriptFlags struct {
	device       *string
	interfaceRef *string
	speed        *string
	jlinkPath    *string
	timeout      *string
	yes          *bool
}

func runScript(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "script requires subcommand: probe, connect, flash, memory-read, reset, halt, go, breakpoint-set, or breakpoint-clear", Exit: ExitUsage}
	}
	switch args[0] {
	case "probe":
		return runScriptProbe(args[1:], jsonOutput)
	case "connect":
		return runScriptConnect(args[1:], jsonOutput)
	case "flash":
		return runScriptFlash(args[1:], jsonOutput)
	case "memory-read":
		return runScriptMemoryRead(args[1:], jsonOutput)
	case "reset", "halt", "go":
		return runScriptTargetControl(args[0], args[1:], jsonOutput)
	case "breakpoint-set":
		return runScriptBreakpoint("set", args[1:], jsonOutput)
	case "breakpoint-clear":
		return runScriptBreakpoint("clear", args[1:], jsonOutput)
	default:
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown script subcommand %q", args[0]), Exit: ExitUsage}
	}
}

func runScriptProbe(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script probe")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	return ExitOK, scriptgen.ProbeScript(), []protocol.Diagnostic{{Level: "info", Code: "script.probe_only", Message: "generated probe-only J-Link Commander script"}}, nil
}

func runScriptConnect(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script connect")
	flags := addScriptFlags(fs)
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	script, err := scriptgen.ConnectScript(targetOptions(flags))
	if err != nil {
		return ExitUsage, scriptResult{Script: script}, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	return finishScript(script, flags)
}

func runScriptFlash(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script flash")
	flags := addScriptFlags(fs)
	file := fs.String("file", "", "firmware image path")
	addressValue := fs.String("address", "0x08000000", "program address")
	verify := fs.Bool("verify", true, "verify programmed flash")
	reset := fs.Bool("reset", false, "reset before programming")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	address, err := parseRequiredUint("address", *addressValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	script, err := scriptgen.FlashScript(scriptgen.FlashOptions{TargetOptions: targetOptions(flags), File: *file, Address: address, Verify: *verify, Reset: *reset})
	if err != nil {
		return ExitUsage, scriptResult{Script: script}, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	return finishScript(script, flags)
}

func runScriptMemoryRead(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script memory-read")
	flags := addScriptFlags(fs)
	addressValue := fs.String("address", "", "memory address, for example 0x20000000")
	lengthValue := fs.String("length", "", "number of units to read")
	width := fs.Uint("width", 8, "access width: 8, 16, or 32")
	halt := fs.Bool("halt", false, "halt the MCU before reading, then resume")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	address, err := parseRequiredUint("address", *addressValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	length, err := parseRequiredUint("length", *lengthValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	script, err := scriptgen.MemoryReadScript(scriptgen.MemoryReadOptions{TargetOptions: targetOptions(flags), Address: address, Length: length, Width: *width, Halt: *halt})
	if err != nil {
		return ExitUsage, scriptResult{Script: script}, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	return finishScript(script, flags)
}

func runScriptTargetControl(action string, args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script " + action)
	flags := addScriptFlags(fs)
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	script, err := scriptgen.TargetControlScript(targetOptions(flags), action)
	if err != nil {
		return ExitUsage, scriptResult{Script: script}, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	return finishScript(script, flags)
}

func runScriptBreakpoint(action string, args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script breakpoint-" + action)
	flags := addScriptFlags(fs)
	addressValue := fs.String("address", "", "breakpoint address")
	halt := fs.Bool("halt", true, "halt before changing breakpoint")
	runAfter := fs.Bool("run", false, "resume after breakpoint operation")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	address, err := parseRequiredUint("address", *addressValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	script, err := scriptgen.BreakpointScript(scriptgen.BreakpointOptions{TargetOptions: targetOptions(flags), Action: action, Address: address, Halt: *halt, Run: *runAfter})
	if err != nil {
		return ExitUsage, scriptResult{Script: script}, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	return finishScript(script, flags)
}

func runFlash(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "flash requires subcommand: program", Exit: ExitUsage}
	}
	if args[0] != "program" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown flash subcommand %q", args[0]), Exit: ExitUsage}
	}
	return runScriptFlash(args[1:], jsonOutput)
}

func runMemory(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "memory requires subcommand: read", Exit: ExitUsage}
	}
	if args[0] != "read" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown memory subcommand %q", args[0]), Exit: ExitUsage}
	}
	return runScriptMemoryRead(args[1:], jsonOutput)
}

func runTarget(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "target requires subcommand: reset, halt, or run", Exit: ExitUsage}
	}
	action := args[0]
	if action == "run" {
		action = "go"
	}
	if action != "reset" && action != "halt" && action != "go" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown target subcommand %q", args[0]), Exit: ExitUsage}
	}
	return runScriptTargetControl(action, args[1:], jsonOutput)
}

func addScriptFlags(fs interface {
	String(string, string, string) *string
	Bool(string, bool, string) *bool
}) scriptFlags {
	return scriptFlags{
		device:       fs.String("device", "STM32H750VB", "J-Link device name"),
		interfaceRef: fs.String("interface", "SWD", "target interface"),
		speed:        fs.String("speed", "4000", "target interface speed in kHz"),
		jlinkPath:    fs.String("jlink-path", "", "path to SEGGER J-Link executable"),
		timeout:      fs.String("timeout", "60s", "maximum execution time"),
		yes:          fs.Bool("yes", false, "execute the generated script"),
	}
}

func targetOptions(flags scriptFlags) scriptgen.TargetOptions {
	return scriptgen.TargetOptions{Device: *flags.device, Interface: *flags.interfaceRef, Speed: *flags.speed}
}

func finishScript(script scriptgen.Script, flags scriptFlags) (int, any, []protocol.Diagnostic, error) {
	diagnostics := scriptDiagnostics(script, *flags.yes)
	result := scriptResult{Script: script}
	if !*flags.yes {
		return ExitOK, result, diagnostics, nil
	}
	executable, err := resolveJLinkExecutable(*flags.jlinkPath)
	if err != nil {
		return ExitNotReady, result, diagnostics, err
	}
	timeout, err := time.ParseDuration(*flags.timeout)
	if err != nil {
		return ExitUsage, result, diagnostics, &cliError{Code: "usage.invalid_timeout", Message: err.Error(), Exit: ExitUsage}
	}
	execution, err := jlink.RunScript(jlink.ScriptOptions{Executable: executable, Script: script.Commands, Timeout: timeout})
	result.Execution = &execution
	if err != nil {
		status := ExitRuntime
		if execution.TimedOut {
			status = ExitCancelled
		}
		return status, result, diagnostics, &cliError{Code: "jlink.script_failed", Message: err.Error(), Exit: status}
	}
	return ExitOK, result, diagnostics, nil
}

func scriptDiagnostics(script scriptgen.Script, executing bool) []protocol.Diagnostic {
	if script.Classification == scriptgen.ProbeOnly {
		return []protocol.Diagnostic{{Level: "info", Code: "script.probe_only", Message: "generated probe-only J-Link Commander script"}}
	}
	message := "generated script connects to the target; pass --yes only after user confirmation"
	if executing {
		message = "executing confirmed J-Link Commander script"
	}
	level := "warning"
	if executing {
		level = "info"
	}
	return []protocol.Diagnostic{{Level: level, Code: "script.requires_confirmation", Message: message}}
}

func resolveJLinkExecutable(path string) (string, error) {
	report := jlink.Discover(jlink.Options{ExplicitPath: path})
	if strings.TrimSpace(report.Selected) == "" {
		return "", &cliError{Code: "jlink.not_found", Message: "install SEGGER J-Link or pass --jlink-path", Exit: ExitNotReady}
	}
	return report.Selected, nil
}
