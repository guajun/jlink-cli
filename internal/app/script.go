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
	if isHelpArg(args[0]) {
		return ExitOK, scriptUsage(), nil, nil
	}
	switch args[0] {
	case "probe":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptProbeUsage(), nil, nil
		}
		return runScriptProbe(args[1:], jsonOutput)
	case "connect":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptConnectUsage(), nil, nil
		}
		return runScriptConnect(args[1:], jsonOutput)
	case "flash":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptFlashUsage(), nil, nil
		}
		return runScriptFlash(args[1:], jsonOutput)
	case "memory-read":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptMemoryReadUsage(), nil, nil
		}
		return runScriptMemoryRead(args[1:], jsonOutput)
	case "reset", "halt", "go":
		if hasHelpArg(args[1:]) {
			return ExitOK, targetActionUsage(args[0]), nil, nil
		}
		return runScriptTargetControl(args[0], args[1:], jsonOutput)
	case "breakpoint-set":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptBreakpointUsage("set"), nil, nil
		}
		return runScriptBreakpoint("set", args[1:], jsonOutput)
	case "breakpoint-clear":
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptBreakpointUsage("clear"), nil, nil
		}
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
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_file", Message: "flash requires --file", Exit: ExitUsage}
	}
	if isHelpArg(args[0]) {
		return ExitOK, flashUsage(), nil, nil
	}
	if args[0] == "program" {
		if hasHelpArg(args[1:]) {
			return ExitOK, scriptFlashUsage(), nil, nil
		}
		return runScriptFlash(args[1:], jsonOutput)
	}
	if strings.HasPrefix(args[0], "-") {
		return runScriptFlash(args, jsonOutput)
	}
	if args[0] != "program" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown flash subcommand %q", args[0]), Exit: ExitUsage}
	}
	return runScriptFlash(args, jsonOutput)
}

func runMemory(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "memory requires subcommand: read", Exit: ExitUsage}
	}
	if isHelpArg(args[0]) {
		return ExitOK, memoryUsage(), nil, nil
	}
	if args[0] != "read" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown memory subcommand %q", args[0]), Exit: ExitUsage}
	}
	if hasHelpArg(args[1:]) {
		return ExitOK, scriptMemoryReadUsage(), nil, nil
	}
	return runScriptMemoryRead(args[1:], jsonOutput)
}

func runTarget(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "target requires subcommand: reset, halt, or run", Exit: ExitUsage}
	}
	if isHelpArg(args[0]) {
		return ExitOK, targetUsage(), nil, nil
	}
	action := args[0]
	if action == "run" {
		action = "go"
	}
	if action != "reset" && action != "halt" && action != "go" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown target subcommand %q", args[0]), Exit: ExitUsage}
	}
	if hasHelpArg(args[1:]) {
		return ExitOK, targetActionUsage(args[0]), nil, nil
	}
	return runScriptTargetControl(action, args[1:], jsonOutput)
}

func isHelpArg(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

func hasHelpArg(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func scriptUsage() string {
	return "Usage:\n" +
		"  jlink-cli script <subcommand> [flags]\n" +
		"\n" +
		"Subcommands:\n" +
		"  probe             Generate a probe-only J-Link Commander script\n" +
		"  connect           Generate or execute a target connect script\n" +
		"  flash             Generate or execute a flash programming script\n" +
		"  memory-read       Generate or execute a memory read script\n" +
		"  reset             Generate or execute a target reset script\n" +
		"  halt              Generate or execute a target halt script\n" +
		"  go                Generate or execute a target resume script\n" +
		"  breakpoint-set    Generate or execute a breakpoint set script\n" +
		"  breakpoint-clear  Generate or execute a breakpoint clear script\n"
}

func flashUsage() string {
	return "Usage:\n" +
		"  jlink-cli flash --file <path> [flags]\n" +
		"  jlink-cli flash program --file <path> [flags]\n" +
		"\n" + scriptFlashFlags()
}

func memoryUsage() string {
	return "Usage:\n" +
		"  jlink-cli memory read --address <addr> --length <n> [flags]\n" +
		"\n" + scriptMemoryReadFlags()
}

func targetUsage() string {
	return "Usage:\n" +
		"  jlink-cli target <reset|halt|run> [flags]\n" +
		"\n" +
		"Subcommands:\n" +
		"  reset  Reset the target\n" +
		"  halt   Halt the target CPU\n" +
		"  run    Resume target execution\n" +
		"  go     Alias for run\n" +
		"\n" + commonTargetFlags()
}

func scriptProbeUsage() string {
	return "Usage:\n" +
		"  jlink-cli script probe [--json]\n"
}

func scriptConnectUsage() string {
	return "Usage:\n" +
		"  jlink-cli script connect [flags]\n" +
		"\n" + commonTargetFlags()
}

func scriptFlashUsage() string {
	return "Usage:\n" +
		"  jlink-cli script flash --file <path> [flags]\n" +
		"\n" + scriptFlashFlags()
}

func scriptMemoryReadUsage() string {
	return "Usage:\n" +
		"  jlink-cli script memory-read --address <addr> --length <n> [flags]\n" +
		"\n" + scriptMemoryReadFlags()
}

func targetActionUsage(action string) string {
	if action == "run" {
		action = "run"
	}
	return "Usage:\n" +
		fmt.Sprintf("  jlink-cli target %s [flags]\n", action) +
		"\n" + commonTargetFlags()
}

func scriptBreakpointUsage(action string) string {
	return "Usage:\n" +
		fmt.Sprintf("  jlink-cli script breakpoint-%s --address <addr> [flags]\n", action) +
		"\n" +
		"Flags:\n" +
		"  --address <addr>     Breakpoint address\n" +
		"  --halt               Halt before changing breakpoint (default true)\n" +
		"  --run                Resume after breakpoint operation\n" +
		commonTargetFlagLines()
}

func scriptFlashFlags() string {
	return "Flags:\n" +
		"  --file <path>        Firmware image path\n" +
		"  --address <addr>     Program address (default 0x08000000)\n" +
		"  --verify             Verify programmed flash (default true)\n" +
		"  --reset              Reset before programming\n" +
		commonTargetFlagLines()
}

func scriptMemoryReadFlags() string {
	return "Flags:\n" +
		"  --address <addr>     Memory address, for example 0x20000000\n" +
		"  --length <n>         Number of units to read\n" +
		"  --width <bits>       Access width: 8, 16, or 32 (default 8)\n" +
		"  --halt               Halt the MCU before reading, then resume\n" +
		commonTargetFlagLines()
}

func commonTargetFlags() string {
	return "Flags:\n" + commonTargetFlagLines()
}

func commonTargetFlagLines() string {
	return "  --device <name>      J-Link device name (default STM32H750VB)\n" +
		"  --interface <name>   Target interface (default SWD)\n" +
		"  --speed <value>      Target interface speed in kHz (default 4000)\n" +
		"  --jlink-path <path>  Path to SEGGER J-Link executable\n" +
		"  --timeout <duration> Maximum execution time (default 60s)\n" +
		"  --yes               Execute the generated script\n" +
		"  --json              Write JSON output\n"
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
