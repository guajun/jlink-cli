package jlink

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildConnectSpec(t *testing.T) {
	spec, err := BuildConnectSpec(ConnectOptions{
		Executable: `C:\Program Files\SEGGER\JLink\JLink.exe`,
		Device:     "STM32H750VB",
		Interface:  "swd",
		Speed:      "4000",
		Serial:     "123456789",
	}, `C:\Temp\connect.jlink`)
	if err != nil {
		t.Fatal(err)
	}

	expectedArgs := []string{"-device", "STM32H750VB", "-if", "SWD", "-speed", "4000", "-autoconnect", "1", "-ExitOnError", "1", "-SelectEmuBySN", "123456789", "-CommanderScript", `C:\Temp\connect.jlink`}
	if !reflect.DeepEqual(spec.Args, expectedArgs) {
		t.Fatalf("unexpected args:\nwant %#v\n got %#v", expectedArgs, spec.Args)
	}
	expectedScript := []string{"connect", "q"}
	if !reflect.DeepEqual(spec.Script, expectedScript) {
		t.Fatalf("unexpected script: %#v", spec.Script)
	}
}

func TestConnectDryRunDoesNotExecute(t *testing.T) {
	result, err := Connect(ConnectOptions{
		Executable: "definitely-not-a-real-jlink-executable",
		Device:     "STM32H750VB",
		Interface:  "SWD",
		Speed:      "4000",
		Timeout:    time.Second,
		DryRun:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun {
		t.Fatalf("expected dry run result")
	}
	if result.Command.Executable == "" {
		t.Fatalf("expected command spec")
	}
}
