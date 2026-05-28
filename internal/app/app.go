package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/guajun/jlink-cli/internal/buildinfo"
	"github.com/guajun/jlink-cli/internal/jlink"
	"github.com/guajun/jlink-cli/internal/protocol"
)

const (
	ExitOK        = 0
	ExitUsage     = 2
	ExitRuntime   = 1
	ExitNotReady  = 3
	ExitCancelled = 130
)

type cliError struct {
	Code    string
	Message string
	Exit    int
	Details map[string]string
}

func (err *cliError) Error() string {
	return err.Message
}

func Main(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		writeHumanUsage(stderr)
		return ExitUsage
	}

	jsonOutput := false
	status, result, diagnostics, err := run(args, &jsonOutput)
	if jsonOutput {
		if err != nil {
			writeJSON(stdout, protocol.Failure(errorCode(err), err.Error(), errorDetails(err), diagnostics))
		} else {
			writeJSON(stdout, protocol.Success(result, diagnostics))
		}
	} else {
		writeHuman(stdout, stderr, args[0], result, diagnostics, err)
	}
	return status
}

func run(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	command := args[0]
	switch command {
	case "version", "--version", "-v":
		return runVersion(args[1:], jsonOutput)
	case "doctor":
		return runDoctor(args[1:], jsonOutput)
	case "inspect":
		return runInspect(args[1:], jsonOutput)
	case "connect":
		return runConnect(args[1:], jsonOutput)
	case "flash":
		return runFlash(args[1:], jsonOutput)
	case "memory":
		return runMemory(args[1:], jsonOutput)
	case "breakpoint":
		return runBreakpoint(args[1:], jsonOutput)
	case "callstack":
		return runCallStack(args[1:], jsonOutput)
	case "dll":
		return runDLL(args[1:], jsonOutput)
	case "run":
		return runAgent(args[1:], jsonOutput)
	case "help", "--help", "-h":
		return ExitOK, usage(), nil, nil
	default:
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_command", Message: fmt.Sprintf("unknown command %q", command), Exit: ExitUsage}
	}
}

func runVersion(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("version")
	fs.BoolVar(jsonOutput, "json", false, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	return ExitOK, buildinfo.Current(), nil, nil
}

func runDoctor(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("doctor")
	binaryPath := fs.String("jlink-path", "", "path to a SEGGER J-Link command-line executable")
	fs.BoolVar(jsonOutput, "json", false, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}

	report := jlink.Discover(jlink.Options{ExplicitPath: *binaryPath})
	diagnostics := []protocol.Diagnostic{}
	if report.Selected == "" {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Level:   "warning",
			Code:    "jlink.not_found",
			Message: "no executable J-Link command-line tool was found; install SEGGER J-Link or pass --jlink-path",
		})
	} else {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Level:   "info",
			Code:    "jlink.found",
			Message: "found a J-Link command-line executable",
			Details: map[string]string{"path": report.Selected},
		})
	}
	return ExitOK, report, diagnostics, nil
}

func runInspect(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("inspect")
	fs.BoolVar(jsonOutput, "json", false, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	return ExitOK, map[string]any{
		"build":    buildinfo.Current(),
		"commands": []string{"version", "doctor", "inspect", "connect", "flash", "memory", "breakpoint", "callstack", "dll", "run"},
		"platform": map[string]string{"goos": runtime.GOOS, "goarch": runtime.GOARCH},
	}, nil, nil
}

func runConnect(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("connect")
	device := fs.String("device", "", "J-Link device name, for example STM32H750VB")
	targetInterface := fs.String("interface", "SWD", "target interface, for example SWD or JTAG")
	speed := fs.String("speed", "4000", "target interface speed in kHz or adaptive")
	serial := fs.String("serial", "", "optional J-Link serial number")
	binaryPath := fs.String("jlink-path", "", "path to a SEGGER J-Link command-line executable")
	timeoutValue := fs.String("timeout", "15s", "maximum time to wait for J-Link Commander")
	dryRun := fs.Bool("dry-run", false, "print the J-Link command without opening a session")
	yes := fs.Bool("yes", false, "allow opening an exclusive J-Link session")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	if strings.TrimSpace(*device) == "" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_device", Message: "connect requires --device, for example --device STM32H750VB", Exit: ExitUsage}
	}
	timeout, err := time.ParseDuration(*timeoutValue)
	if err != nil {
		return ExitUsage, nil, nil, &cliError{Code: "usage.invalid_timeout", Message: err.Error(), Exit: ExitUsage}
	}

	report := jlink.Discover(jlink.Options{ExplicitPath: *binaryPath})
	if report.Selected == "" {
		return ExitNotReady, report, []protocol.Diagnostic{{Level: "error", Code: "jlink.not_found", Message: "no executable SEGGER J-Link command-line tool was found"}}, &cliError{Code: "jlink.not_found", Message: "install SEGGER J-Link or pass --jlink-path", Exit: ExitNotReady}
	}

	connectOptions := jlink.ConnectOptions{
		Executable: report.Selected,
		Device:     *device,
		Interface:  *targetInterface,
		Speed:      *speed,
		Serial:     *serial,
		Timeout:    timeout,
		DryRun:     *dryRun || !*yes,
	}
	result, err := jlink.Connect(connectOptions)
	diagnostics := []protocol.Diagnostic{{Level: "info", Code: "jlink.selected", Message: "selected J-Link command-line executable", Details: map[string]string{"path": report.Selected}}}
	if connectOptions.DryRun && !*yes {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "warning", Code: "connect.dry_run", Message: "no physical J-Link session was opened; pass --yes to connect"})
	}
	if err != nil {
		status := ExitRuntime
		if result.TimedOut {
			status = ExitCancelled
		}
		return status, result, diagnostics, &cliError{Code: "jlink.connect_failed", Message: err.Error(), Exit: status}
	}
	return ExitOK, result, diagnostics, nil
}

func runAgent(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("run")
	input := fs.String("input", "", "inline JSON request")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	if strings.TrimSpace(*input) == "" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_input", Message: "run requires --input JSON for the initial non-interactive prototype", Exit: ExitUsage}
	}

	var request map[string]any
	if err := json.Unmarshal([]byte(*input), &request); err != nil {
		return ExitUsage, nil, nil, &cliError{Code: "protocol.invalid_json", Message: err.Error(), Exit: ExitUsage}
	}

	return ExitOK, map[string]any{
		"request": request,
		"status":  "accepted",
		"note":    "prototype run command parsed the request without touching a physical J-Link device",
	}, []protocol.Diagnostic{{Level: "info", Code: "run.prototype", Message: "no exclusive J-Link operation was performed"}}, nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func usageError(err error) error {
	return &cliError{Code: "usage.invalid_flags", Message: err.Error(), Exit: ExitUsage}
}

func errorCode(err error) string {
	var cliErr *cliError
	if errors.As(err, &cliErr) {
		return cliErr.Code
	}
	return "runtime.error"
}

func errorDetails(err error) map[string]string {
	var cliErr *cliError
	if errors.As(err, &cliErr) {
		return cliErr.Details
	}
	return nil
}

func writeJSON(w io.Writer, response protocol.Response) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(response)
}

func writeHuman(stdout io.Writer, stderr io.Writer, command string, result any, diagnostics []protocol.Diagnostic, err error) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err.Error())
		return
	}
	if command == "help" || command == "--help" || command == "-h" {
		fmt.Fprint(stdout, result)
		return
	}
	writeJSON(stdout, protocol.Success(result, nil))
}

func writeHumanUsage(stderr io.Writer) {
	fmt.Fprint(stderr, usage())
}

func usage() string {
	return `jlink-cli is a non-interactive command-line interface for AI agents working with SEGGER J-Link tools.

Usage:
  jlink-cli <command> [flags]

Commands:
  version   Print build and protocol version information
  doctor    Diagnose local J-Link CLI availability without opening a device session
  inspect   Print command and platform metadata
	connect   Test a J-Link Commander connection; requires --yes to open a physical session
	flash     Program firmware through J-Link Commander
	memory    Read target memory, live or halt/resume depending on flags
	breakpoint Set or clear target breakpoints
	callstack Prepare GDB commands for halted-target backtraces
  dll       Diagnose direct J-Link DLL loading without opening a probe session
  run       Parse an agent request; physical J-Link operations are not enabled in the prototype

Global conventions:
  Machine-readable results go to stdout. Diagnostics and errors go to stderr.
  Use --json on commands that support human output.
`
}
