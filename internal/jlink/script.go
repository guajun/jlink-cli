package jlink

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ScriptOptions struct {
	Executable string
	Script     []string
	Serial     string
	Timeout    time.Duration
	DryRun     bool
}

func RunScript(opts ScriptOptions) (ConnectResult, error) {
	spec := CommandSpec{Executable: opts.Executable, Args: buildScriptArgs(opts.Serial, "<temp-script>"), Script: opts.Script}
	if opts.DryRun {
		return ConnectResult{DryRun: true, Command: spec}, nil
	}
	if opts.Executable == "" {
		return ConnectResult{Command: spec}, errors.New("missing J-Link executable")
	}
	tempDir, err := os.MkdirTemp("", "jlink-cli-*")
	if err != nil {
		return ConnectResult{Command: spec}, err
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "operation.jlink")
	if err := os.WriteFile(scriptPath, []byte(joinScript(opts.Script)), 0600); err != nil {
		return ConnectResult{Command: spec}, err
	}
	spec.Args = buildScriptArgs(opts.Serial, scriptPath)

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

func buildScriptArgs(serial string, scriptPath string) []string {
	args := []string{"-NoGui", "1"}
	if strings.TrimSpace(serial) != "" {
		args = append(args, "-SelectEmuBySN", serial)
	}
	return append(args, "-CommanderScript", scriptPath)
}

func joinScript(commands []string) string {
	text := ""
	for _, command := range commands {
		text += command + "\n"
	}
	return text
}
