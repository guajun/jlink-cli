package app

import (
	"fmt"

	"github.com/guajun/jlink-cli/internal/jlinkdll"
	"github.com/guajun/jlink-cli/internal/protocol"
)

func runDLL(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "dll requires subcommand: doctor or connect", Exit: ExitUsage}
	}
	subcommand := args[0]
	switch subcommand {
	case "doctor":
		return runDLLDoctor(args[1:], jsonOutput)
	case "connect":
		return runDLLConnect(args[1:], jsonOutput)
	default:
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown dll subcommand %q", subcommand), Exit: ExitUsage}
	}
}

func runDLLDoctor(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("dll doctor")
	dllPath := fs.String("dll-path", "", "path to JLink_x64.dll or platform equivalent")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}

	report := jlinkdll.Doctor(jlinkdll.Options{ExplicitPath: *dllPath})
	diagnostics := []protocol.Diagnostic{}
	if report.Selected == "" {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "error", Code: "jlinkdll.not_found", Message: "no loadable J-Link DLL was found"})
		return ExitNotReady, report, diagnostics, &cliError{Code: "jlinkdll.not_found", Message: "install SEGGER J-Link or pass --dll-path", Exit: ExitNotReady}
	}
	diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "jlinkdll.loaded", Message: "loaded J-Link DLL and probed required symbols", Details: map[string]string{"path": report.Selected}})
	if len(report.MissingRequired) > 0 {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "warning", Code: "jlinkdll.missing_symbols", Message: "some required symbols were not exported by the selected J-Link DLL"})
	}
	return ExitOK, report, diagnostics, nil
}

func runDLLConnect(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("dll connect")
	dllPath := fs.String("dll-path", "", "path to JLink_x64.dll or platform equivalent")
	device := fs.String("device", "STM32H750VB", "J-Link device name")
	targetInterface := fs.String("interface", "SWD", "target interface, for example SWD or JTAG")
	speed := fs.String("speed", "4000", "target interface speed in kHz, or auto")
	dryRun := fs.Bool("dry-run", false, "print DLL call plan without opening a physical J-Link session")
	yes := fs.Bool("yes", false, "allow opening an exclusive J-Link session through the DLL")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}

	report := jlinkdll.Doctor(jlinkdll.Options{ExplicitPath: *dllPath})
	diagnostics := []protocol.Diagnostic{}
	if report.Selected == "" {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "error", Code: "jlinkdll.not_found", Message: "no loadable J-Link DLL was found"})
		return ExitNotReady, report, diagnostics, &cliError{Code: "jlinkdll.not_found", Message: "install SEGGER J-Link or pass --dll-path", Exit: ExitNotReady}
	}
	diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "jlinkdll.loaded", Message: "loaded J-Link DLL", Details: map[string]string{"path": report.Selected}})
	connectOptions := jlinkdll.ConnectOptions{
		DLLPath:   report.Selected,
		Device:    *device,
		Interface: *targetInterface,
		Speed:     *speed,
		DryRun:    *dryRun || !*yes,
	}
	result, err := jlinkdll.Connect(connectOptions)
	if connectOptions.DryRun && !*yes {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "warning", Code: "jlinkdll.connect_dry_run", Message: "no physical J-Link session was opened; pass --yes to connect through the DLL"})
	}
	if err != nil {
		return ExitRuntime, result, diagnostics, &cliError{Code: "jlinkdll.connect_failed", Message: err.Error(), Exit: ExitRuntime}
	}
	return ExitOK, result, diagnostics, nil
}
