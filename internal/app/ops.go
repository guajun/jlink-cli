package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guajun/jlink-cli/internal/jlink"
	"github.com/guajun/jlink-cli/internal/protocol"
)

type commonJLinkFlags struct {
	device       *string
	iface        *string
	speed        *string
	serial       *string
	jlinkPath    *string
	timeoutValue *string
	dryRun       *bool
	yes          *bool
}

func runFlash(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "flash requires subcommand: program", Exit: ExitUsage}
	}
	subcommand := args[0]
	if subcommand != "program" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown flash subcommand %q", subcommand), Exit: ExitUsage}
	}

	fs := newFlagSet("flash program")
	common := addCommonJLinkFlags(fs, "60s")
	file := fs.String("file", "", "firmware image path")
	addressValue := fs.String("address", "", "program address, for example 0x08000000")
	verify := fs.Bool("verify", true, "verify programmed flash")
	reset := fs.Bool("reset", false, "reset before and after programming")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args[1:]); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	address, err := parseRequiredUint("address", *addressValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	commonOptions, diagnostics, err := resolveCommonJLinkOptions(common)
	if err != nil {
		return ExitNotReady, nil, diagnostics, err
	}

	result, err := jlink.ProgramFlash(jlink.FlashOptions{
		CommonOptions: commonOptions,
		File:          *file,
		Address:       address,
		Verify:        *verify,
		Reset:         *reset,
	})
	diagnostics = appendDryRunDiagnostic(diagnostics, commonOptions.DryRun, "flash.dry_run", "no flash programming was performed; pass --yes to program the target")
	return finishJLinkResult(result, diagnostics, err)
}

func runMemory(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "memory requires subcommand: read", Exit: ExitUsage}
	}
	subcommand := args[0]
	if subcommand != "read" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown memory subcommand %q", subcommand), Exit: ExitUsage}
	}

	fs := newFlagSet("memory read")
	common := addCommonJLinkFlags(fs, "20s")
	addressValue := fs.String("address", "", "memory address, for example 0x20000000")
	lengthValue := fs.String("length", "", "number of units to read")
	width := fs.Uint("width", 8, "access width: 8, 16, or 32")
	halt := fs.Bool("halt", false, "halt the MCU before reading, then resume")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args[1:]); err != nil {
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
	commonOptions, diagnostics, err := resolveCommonJLinkOptions(common)
	if err != nil {
		return ExitNotReady, nil, diagnostics, err
	}

	result, err := jlink.ReadMemory(jlink.MemoryReadOptions{
		CommonOptions: commonOptions,
		Address:       address,
		Length:        length,
		Width:         *width,
		Halt:          *halt,
	})
	if *halt {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "memory.halted", Message: "memory read script will halt and resume the MCU"})
	} else {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "memory.live", Message: "memory read script does not issue halt before reading"})
	}
	diagnostics = appendDryRunDiagnostic(diagnostics, commonOptions.DryRun, "memory.dry_run", "no physical memory read was performed; pass --yes to read target memory")
	return finishJLinkResult(result, diagnostics, err)
}

func runBreakpoint(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "breakpoint requires subcommand: set or clear", Exit: ExitUsage}
	}
	action := args[0]
	if action != "set" && action != "clear" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown breakpoint subcommand %q", action), Exit: ExitUsage}
	}

	fs := newFlagSet("breakpoint " + action)
	common := addCommonJLinkFlags(fs, "20s")
	addressValue := fs.String("address", "", "breakpoint address, for example 0x08000100")
	runAfter := fs.Bool("run", false, "resume execution after breakpoint operation")
	halt := fs.Bool("halt", true, "halt before changing breakpoint")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args[1:]); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	address, err := parseRequiredUint("address", *addressValue)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	commonOptions, diagnostics, err := resolveCommonJLinkOptions(common)
	if err != nil {
		return ExitNotReady, nil, diagnostics, err
	}

	result, err := jlink.ControlBreakpoint(jlink.BreakpointOptions{
		CommonOptions: commonOptions,
		Action:        action,
		Address:       address,
		Run:           *runAfter,
		Halt:          *halt,
	})
	diagnostics = appendDryRunDiagnostic(diagnostics, commonOptions.DryRun, "breakpoint.dry_run", "no breakpoint operation was performed; pass --yes to change target breakpoints")
	return finishJLinkResult(result, diagnostics, err)
}

func runCallStack(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("callstack")
	common := addCommonJLinkFlags(fs, "30s")
	elf := fs.String("elf", "", "ELF file with symbols")
	gdbPath := fs.String("gdb", "", "path to arm-none-eabi-gdb")
	gdbServerPath := fs.String("gdb-server", "", "path to JLinkGDBServerCL executable")
	port := fs.Uint("port", 2331, "GDB server TCP port")
	maxFrames := fs.Uint("max-frames", 16, "maximum backtrace frames")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	commonOptions, diagnostics, err := resolveCommonJLinkOptions(common)
	if err != nil {
		return ExitNotReady, nil, diagnostics, err
	}
	callStackOptions := jlink.CallStackOptions{
		CommonOptions: commonOptions,
		GDBPath:       *gdbPath,
		GDBServerPath: *gdbServerPath,
		ELF:           *elf,
		Port:          *port,
		MaxFrames:     *maxFrames,
	}
	result, err := jlink.CaptureCallStack(callStackOptions)
	diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "callstack.halted_target", Message: "call stack capture expects the MCU to already be halted at a breakpoint"})
	diagnostics = appendDryRunDiagnostic(diagnostics, callStackOptions.DryRun, "callstack.dry_run", "call stack execution was not started; pass --yes to start JLinkGDBServerCL and arm-none-eabi-gdb")
	if err != nil {
		status := ExitRuntime
		if result.TimedOut {
			status = ExitCancelled
		}
		return status, result, diagnostics, &cliError{Code: "callstack.failed", Message: err.Error(), Exit: status}
	}
	return ExitOK, result, diagnostics, nil
}

func addCommonJLinkFlags(fs interface {
	String(string, string, string) *string
	Bool(string, bool, string) *bool
}, defaultTimeout string) commonJLinkFlags {
	return commonJLinkFlags{
		device:       fs.String("device", "STM32H750VB", "J-Link device name"),
		iface:        fs.String("interface", "SWD", "target interface, for example SWD or JTAG"),
		speed:        fs.String("speed", "4000", "target interface speed in kHz or adaptive"),
		serial:       fs.String("serial", "", "optional J-Link serial number"),
		jlinkPath:    fs.String("jlink-path", "", "path to a SEGGER J-Link command-line executable"),
		timeoutValue: fs.String("timeout", defaultTimeout, "maximum time to wait"),
		dryRun:       fs.Bool("dry-run", false, "print command without opening a physical J-Link session"),
		yes:          fs.Bool("yes", false, "allow opening an exclusive J-Link session"),
	}
}

func resolveCommonJLinkOptions(flags commonJLinkFlags) (jlink.CommonOptions, []protocol.Diagnostic, error) {
	timeout, err := time.ParseDuration(*flags.timeoutValue)
	if err != nil {
		return jlink.CommonOptions{}, nil, &cliError{Code: "usage.invalid_timeout", Message: err.Error(), Exit: ExitUsage}
	}
	report := jlink.Discover(jlink.Options{ExplicitPath: *flags.jlinkPath})
	diagnostics := []protocol.Diagnostic{}
	if report.Selected == "" {
		diagnostics = append(diagnostics, protocol.Diagnostic{Level: "error", Code: "jlink.not_found", Message: "no executable SEGGER J-Link command-line tool was found"})
		return jlink.CommonOptions{}, diagnostics, &cliError{Code: "jlink.not_found", Message: "install SEGGER J-Link or pass --jlink-path", Exit: ExitNotReady}
	}
	diagnostics = append(diagnostics, protocol.Diagnostic{Level: "info", Code: "jlink.selected", Message: "selected J-Link command-line executable", Details: map[string]string{"path": report.Selected}})
	return jlink.CommonOptions{
		Executable: report.Selected,
		Device:     *flags.device,
		Interface:  *flags.iface,
		Speed:      *flags.speed,
		Serial:     *flags.serial,
		Timeout:    timeout,
		DryRun:     *flags.dryRun || !*flags.yes,
	}, diagnostics, nil
}

func appendDryRunDiagnostic(diagnostics []protocol.Diagnostic, dryRun bool, code string, message string) []protocol.Diagnostic {
	if dryRun {
		return append(diagnostics, protocol.Diagnostic{Level: "warning", Code: code, Message: message})
	}
	return diagnostics
}

func finishJLinkResult(result jlink.ConnectResult, diagnostics []protocol.Diagnostic, err error) (int, any, []protocol.Diagnostic, error) {
	if err != nil {
		status := ExitRuntime
		if result.TimedOut {
			status = ExitCancelled
		}
		return status, result, diagnostics, &cliError{Code: "jlink.operation_failed", Message: err.Error(), Exit: status}
	}
	return ExitOK, result, diagnostics, nil
}

func parseRequiredUint(name string, value string) (uint64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, &cliError{Code: "usage.missing_" + name, Message: fmt.Sprintf("%s is required", name), Exit: ExitUsage}
	}
	parsed, err := strconv.ParseUint(trimmed, 0, 64)
	if err != nil {
		return 0, &cliError{Code: "usage.invalid_" + name, Message: err.Error(), Exit: ExitUsage}
	}
	return parsed, nil
}
