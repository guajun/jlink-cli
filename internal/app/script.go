package app

import (
	"fmt"

	"github.com/guajun/jlink-cli/internal/protocol"
	"github.com/guajun/jlink-cli/internal/scriptgen"
)

func runScript(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "script requires subcommand: probe or memory-read", Exit: ExitUsage}
	}
	switch args[0] {
	case "probe":
		return runScriptProbe(args[1:], jsonOutput)
	case "memory-read":
		return runScriptMemoryRead(args[1:], jsonOutput)
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

func runScriptMemoryRead(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("script memory-read")
	device := fs.String("device", "STM32H750VB", "J-Link device name")
	targetInterface := fs.String("interface", "SWD", "target interface, for example SWD or JTAG")
	speed := fs.String("speed", "4000", "target interface speed in kHz")
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
	script, err := scriptgen.MemoryReadScript(scriptgen.MemoryReadOptions{TargetOptions: scriptgen.TargetOptions{Device: *device, Interface: *targetInterface, Speed: *speed}, Address: address, Length: length, Width: *width, Halt: *halt})
	if err != nil {
		return ExitUsage, script, nil, &cliError{Code: "usage.invalid_script", Message: err.Error(), Exit: ExitUsage}
	}
	diagnostics := []protocol.Diagnostic{{Level: "warning", Code: "script.requires_confirmation", Message: "generated script connects to the target; ask before executing it"}}
	return ExitOK, script, diagnostics, nil
}
