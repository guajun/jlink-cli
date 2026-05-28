package app

import (
	"fmt"

	"github.com/guajun/jlink-cli/internal/jlinkdll"
	"github.com/guajun/jlink-cli/internal/protocol"
)

func runDLL(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "dll requires subcommand: doctor", Exit: ExitUsage}
	}
	subcommand := args[0]
	switch subcommand {
	case "doctor":
		return runDLLDoctor(args[1:], jsonOutput)
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
