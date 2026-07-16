package jlink

import (
	"reflect"
	"testing"
)

func TestRunScriptDryRunSelectsSerial(t *testing.T) {
	result, err := RunScript(ScriptOptions{
		Executable: "JLink.exe",
		Script:     []string{"connect", "mem8 0x20000000 16", "q"},
		Serial:     "123456789",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("RunScript returned error: %v", err)
	}

	want := []string{"-NoGui", "1", "-SelectEmuBySN", "123456789", "-CommanderScript", "<temp-script>"}
	if !reflect.DeepEqual(result.Command.Args, want) {
		t.Fatalf("unexpected args:\n got: %#v\nwant: %#v", result.Command.Args, want)
	}
}
