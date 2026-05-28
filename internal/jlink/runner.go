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

type CommanderScriptOptions struct {
	Executable string
	Device     string
	Interface  string
	Speed      string
	Serial     string
	Timeout    time.Duration
	DryRun     bool
	Script     []string
	ScriptName string
}

func RunCommanderScript(opts CommanderScriptOptions) (ConnectResult, error) {
	spec, err := BuildCommanderScriptSpec(opts, "<temp-script>")
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

	scriptName := opts.ScriptName
	if strings.TrimSpace(scriptName) == "" {
		scriptName = "script.jlink"
	}
	scriptPath := filepath.Join(tempDir, filepath.Base(scriptName))
	return RunCommanderScriptAt(opts, scriptPath)
}

func RunCommanderScriptAt(opts CommanderScriptOptions, scriptPath string) (ConnectResult, error) {
	spec, err := BuildCommanderScriptSpec(opts, scriptPath)
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun, Command: spec}, err
	}
	if opts.DryRun {
		return ConnectResult{DryRun: true, Command: spec}, nil
	}
	if err := os.WriteFile(scriptPath, []byte(strings.Join(spec.Script, "\n")+"\n"), 0600); err != nil {
		return ConnectResult{DryRun: false, Command: spec}, err
	}
	return runCommand(spec, opts.Timeout)
}

func BuildCommanderScriptSpec(opts CommanderScriptOptions, scriptPath string) (CommandSpec, error) {
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
	if len(opts.Script) == 0 {
		return CommandSpec{}, fmt.Errorf("missing J-Link Commander script")
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

	return CommandSpec{Executable: opts.Executable, Args: args, Script: opts.Script}, nil
}

func runCommand(spec CommandSpec, timeout time.Duration) (ConnectResult, error) {
	ctx := context.Background()
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, spec.Executable, spec.Args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

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
