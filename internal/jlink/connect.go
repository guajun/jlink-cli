package jlink

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ConnectOptions struct {
	Executable string
	Device     string
	Interface  string
	Speed      string
	Serial     string
	Timeout    time.Duration
	DryRun     bool
}

type CommandSpec struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
	Script     []string `json:"script"`
}

type ConnectResult struct {
	DryRun   bool        `json:"dry_run"`
	Command  CommandSpec `json:"command"`
	ExitCode int         `json:"exit_code,omitempty"`
	Stdout   string      `json:"stdout,omitempty"`
	Stderr   string      `json:"stderr,omitempty"`
	TimedOut bool        `json:"timed_out,omitempty"`
}

func Connect(opts ConnectOptions) (ConnectResult, error) {
	spec, err := BuildConnectSpec(opts, "<temp-script>")
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun, Command: spec}, err
	}
	if opts.DryRun {
		return ConnectResult{DryRun: true, Command: spec}, nil
	}

	tempDir, err := os.MkdirTemp("", "jlink-cli-*")
	if err != nil {
		return ConnectResult{DryRun: false, Command: spec}, err
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "connect.jlink")
	if err := os.WriteFile(scriptPath, []byte(strings.Join(spec.Script, "\n")+"\n"), 0600); err != nil {
		return ConnectResult{DryRun: false, Command: spec}, err
	}
	spec, err = BuildConnectSpec(opts, scriptPath)
	if err != nil {
		return ConnectResult{DryRun: false, Command: spec}, err
	}

	ctx := context.Background()
	cancel := func() {}
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, spec.Executable, spec.Args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	result := ConnectResult{DryRun: false, Command: spec, Stdout: stdout.String(), Stderr: stderr.String()}
	if ctx.Err() != nil {
		result.TimedOut = true
		result.ExitCode = -1
		return result, ctx.Err()
	}
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
		}
		return result, err
	}
	result.ExitCode = 0
	return result, nil
}

func BuildConnectSpec(opts ConnectOptions, scriptPath string) (CommandSpec, error) {
	if strings.TrimSpace(opts.Executable) == "" {
		return CommandSpec{}, fmt.Errorf("missing J-Link executable")
	}
	if strings.TrimSpace(opts.Device) == "" {
		return CommandSpec{}, fmt.Errorf("missing device")
	}
	if strings.TrimSpace(opts.Interface) == "" {
		return CommandSpec{}, fmt.Errorf("missing target interface")
	}
	if strings.TrimSpace(opts.Speed) == "" {
		return CommandSpec{}, fmt.Errorf("missing speed")
	}

	args := []string{
		"-device", opts.Device,
		"-if", strings.ToUpper(opts.Interface),
		"-speed", opts.Speed,
		"-autoconnect", "1",
		"-ExitOnError", "1",
	}
	if strings.TrimSpace(opts.Serial) != "" {
		args = append(args, "-SelectEmuBySN", opts.Serial)
	}
	args = append(args, "-CommanderScript", scriptPath)

	return CommandSpec{
		Executable: opts.Executable,
		Args:       args,
		Script:     []string{"connect", "q"},
	}, nil
}
