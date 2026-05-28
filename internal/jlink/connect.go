package jlink

import (
	"fmt"
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
	return RunCommanderScript(CommanderScriptOptions{
		Executable: opts.Executable,
		Device:     opts.Device,
		Interface:  opts.Interface,
		Speed:      opts.Speed,
		Serial:     opts.Serial,
		Timeout:    opts.Timeout,
		DryRun:     opts.DryRun,
		Script:     []string{"connect", "q"},
		ScriptName: "connect.jlink",
	})
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
