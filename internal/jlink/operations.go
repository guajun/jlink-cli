package jlink

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CommonOptions struct {
	Executable string
	Device     string
	Interface  string
	Speed      string
	Serial     string
	Timeout    time.Duration
	DryRun     bool
}

type FlashOptions struct {
	CommonOptions
	File    string
	Address uint64
	Verify  bool
	Reset   bool
}

type MemoryReadOptions struct {
	CommonOptions
	Address uint64
	Length  uint64
	Width   uint
	Halt    bool
}

type BreakpointOptions struct {
	CommonOptions
	Action  string
	Address uint64
	Run     bool
	Halt    bool
}

type CallStackOptions struct {
	CommonOptions
	GDBPath       string
	GDBServerPath string
	ELF           string
	Port          uint
	MaxFrames     uint
}

type CallStackResult struct {
	DryRun          bool        `json:"dry_run"`
	GDBServer       CommandSpec `json:"gdb_server"`
	GDB             CommandSpec `json:"gdb"`
	GDBExitCode     int         `json:"gdb_exit_code,omitempty"`
	GDBStdout       string      `json:"gdb_stdout,omitempty"`
	GDBStderr       string      `json:"gdb_stderr,omitempty"`
	GDBServerStdout string      `json:"gdb_server_stdout,omitempty"`
	GDBServerStderr string      `json:"gdb_server_stderr,omitempty"`
	TimedOut        bool        `json:"timed_out,omitempty"`
	Notes           []string    `json:"notes,omitempty"`
}

func ProgramFlash(opts FlashOptions) (ConnectResult, error) {
	script, err := BuildFlashScript(opts)
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun}, err
	}
	return RunCommanderScript(commanderOptions(opts.CommonOptions, script, "flash.jlink"))
}

func ReadMemory(opts MemoryReadOptions) (ConnectResult, error) {
	script, err := BuildMemoryReadScript(opts)
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun}, err
	}
	return RunCommanderScript(commanderOptions(opts.CommonOptions, script, "memory-read.jlink"))
}

func ControlBreakpoint(opts BreakpointOptions) (ConnectResult, error) {
	script, err := BuildBreakpointScript(opts)
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun}, err
	}
	return RunCommanderScript(commanderOptions(opts.CommonOptions, script, "breakpoint.jlink"))
}

func BuildFlashScript(opts FlashOptions) ([]string, error) {
	if strings.TrimSpace(opts.File) == "" {
		return nil, fmt.Errorf("missing firmware file")
	}
	address := formatAddress(opts.Address)
	script := []string{"connect"}
	if opts.Reset {
		script = append(script, "r")
	}
	script = append(script, fmt.Sprintf("loadfile %s %s", quoteCommanderPath(opts.File), address))
	if opts.Verify {
		script = append(script, fmt.Sprintf("verifybin %s %s", quoteCommanderPath(opts.File), address))
	}
	if opts.Reset {
		script = append(script, "r", "g")
	}
	script = append(script, "q")
	return script, nil
}

func BuildMemoryReadScript(opts MemoryReadOptions) ([]string, error) {
	if opts.Length == 0 {
		return nil, fmt.Errorf("length must be greater than zero")
	}
	if opts.Width != 8 && opts.Width != 16 && opts.Width != 32 {
		return nil, fmt.Errorf("width must be one of 8, 16, or 32")
	}
	script := []string{"connect"}
	if opts.Halt {
		script = append(script, "h")
	}
	script = append(script, fmt.Sprintf("mem%s %s %d", strconv.FormatUint(uint64(opts.Width), 10), formatAddress(opts.Address), opts.Length))
	if opts.Halt {
		script = append(script, "g")
	}
	script = append(script, "q")
	return script, nil
}

func BuildBreakpointScript(opts BreakpointOptions) ([]string, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action != "set" && action != "clear" {
		return nil, fmt.Errorf("breakpoint action must be set or clear")
	}
	script := []string{"connect"}
	if opts.Halt {
		script = append(script, "h")
	}
	command := "SetBP"
	if action == "clear" {
		command = "ClrBP"
	}
	script = append(script, fmt.Sprintf("%s %s", command, formatAddress(opts.Address)))
	if opts.Run {
		script = append(script, "g")
	}
	script = append(script, "q")
	return script, nil
}

func BuildCallStackSpec(opts CallStackOptions) (CallStackResult, error) {
	gdbPath := strings.TrimSpace(opts.GDBPath)
	if gdbPath == "" {
		gdbPath = "arm-none-eabi-gdb"
	}
	serverPath := strings.TrimSpace(opts.GDBServerPath)
	if serverPath == "" {
		serverPath = defaultGDBServerPath(opts.Executable)
	}
	port := opts.Port
	if port == 0 {
		port = 2331
	}
	maxFrames := opts.MaxFrames
	if maxFrames == 0 {
		maxFrames = 16
	}
	elf := strings.TrimSpace(opts.ELF)
	if elf == "" {
		elf = "<firmware.elf>"
	}
	if strings.TrimSpace(opts.Device) == "" {
		return CallStackResult{DryRun: true}, fmt.Errorf("missing device")
	}

	serverArgs := []string{"-device", opts.Device, "-if", strings.ToUpper(opts.Interface), "-speed", opts.Speed, "-port", strconv.FormatUint(uint64(port), 10), "-singlerun", "-halt"}
	if strings.TrimSpace(opts.Serial) != "" {
		serverArgs = append(serverArgs, "-select", "USB="+opts.Serial)
	}
	gdbArgs := []string{"--batch", "-ex", fmt.Sprintf("target remote :%d", port), "-ex", "set pagination off", "-ex", fmt.Sprintf("bt %d", maxFrames), elf}

	return CallStackResult{
		DryRun: opts.DryRun,
		GDBServer: CommandSpec{
			Executable: serverPath,
			Args:       serverArgs,
		},
		GDB: CommandSpec{
			Executable: gdbPath,
			Args:       gdbArgs,
		},
		Notes: []string{"call stack capture expects the MCU to already be halted at a breakpoint"},
	}, nil
}

func CaptureCallStack(opts CallStackOptions) (CallStackResult, error) {
	result, err := BuildCallStackSpec(opts)
	if err != nil {
		return result, err
	}
	if opts.DryRun {
		result.DryRun = true
		return result, nil
	}
	if strings.TrimSpace(opts.ELF) == "" {
		return result, fmt.Errorf("--elf is required when executing callstack capture")
	}

	ctx := context.Background()
	cancel := func() {}
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}
	defer cancel()

	serverCmd := exec.CommandContext(ctx, result.GDBServer.Executable, result.GDBServer.Args...)
	var serverStdout bytes.Buffer
	var serverStderr bytes.Buffer
	serverCmd.Stdout = &serverStdout
	serverCmd.Stderr = &serverStderr
	if err := serverCmd.Start(); err != nil {
		result.GDBServerStdout = serverStdout.String()
		result.GDBServerStderr = serverStderr.String()
		return result, err
	}

	serverDone := make(chan error, 1)
	go func() { serverDone <- serverCmd.Wait() }()
	defer func() {
		cancel()
		select {
		case <-serverDone:
		case <-time.After(2 * time.Second):
		}
	}()

	if err := waitForTCP(ctx, opts.Port); err != nil {
		result.GDBServerStdout = serverStdout.String()
		result.GDBServerStderr = serverStderr.String()
		if ctx.Err() != nil {
			result.TimedOut = true
		}
		return result, err
	}

	gdbCmd := exec.CommandContext(ctx, result.GDB.Executable, result.GDB.Args...)
	var gdbStdout bytes.Buffer
	var gdbStderr bytes.Buffer
	gdbCmd.Stdout = &gdbStdout
	gdbCmd.Stderr = &gdbStderr
	gdbErr := gdbCmd.Run()
	result.GDBStdout = gdbStdout.String()
	result.GDBStderr = gdbStderr.String()
	result.GDBServerStdout = serverStdout.String()
	result.GDBServerStderr = serverStderr.String()
	if ctx.Err() != nil {
		result.TimedOut = true
		return result, ctx.Err()
	}
	if gdbErr != nil {
		var exitError *exec.ExitError
		if errors.As(gdbErr, &exitError) {
			result.GDBExitCode = exitError.ExitCode()
		}
		return result, gdbErr
	}
	result.GDBExitCode = 0
	return result, nil
}

func waitForTCP(ctx context.Context, port uint) error {
	address := net.JoinHostPort("127.0.0.1", strconv.FormatUint(uint64(port), 10))
	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		conn, err := net.DialTimeout("tcp", address, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for GDB server on %s: %w", address, lastErr)
}

func commanderOptions(common CommonOptions, script []string, scriptName string) CommanderScriptOptions {
	return CommanderScriptOptions{
		Executable: common.Executable,
		Device:     common.Device,
		Interface:  common.Interface,
		Speed:      common.Speed,
		Serial:     common.Serial,
		Timeout:    common.Timeout,
		DryRun:     common.DryRun,
		Script:     script,
		ScriptName: scriptName,
	}
}

func formatAddress(address uint64) string {
	return fmt.Sprintf("0x%X", address)
}

func quoteCommanderPath(path string) string {
	if strings.ContainsAny(path, " \t") {
		return strconv.Quote(path)
	}
	return path
}

func defaultGDBServerPath(jlinkPath string) string {
	if strings.TrimSpace(jlinkPath) == "" {
		return "JLinkGDBServerCL"
	}
	dir := filepath.Dir(jlinkPath)
	if strings.HasSuffix(strings.ToLower(jlinkPath), ".exe") {
		return filepath.Join(dir, "JLinkGDBServerCL.exe")
	}
	return filepath.Join(dir, "JLinkGDBServerCLExe")
}
