package scriptgen

import (
	"strings"
	"testing"
)

func TestProbeScript(t *testing.T) {
	script := ProbeScript()
	if script.Classification != ProbeOnly || script.RequiresConfirmation {
		t.Fatalf("unexpected probe script metadata: %#v", script)
	}
	if !strings.Contains(script.Text, "ShowEmuList") {
		t.Fatalf("expected ShowEmuList command: %q", script.Text)
	}
}

func TestMemoryReadScript(t *testing.T) {
	script, err := MemoryReadScript(MemoryReadOptions{TargetOptions: TargetOptions{Device: "STM32H750VB", Interface: "swd", Speed: "4000"}, Address: 0x20000000, Length: 16, Width: 8})
	if err != nil {
		t.Fatal(err)
	}
	if script.Classification != TargetSession || !script.RequiresConfirmation {
		t.Fatalf("unexpected memory script metadata: %#v", script)
	}
	for _, want := range []string{"Device STM32H750VB", "SelectInterface SWD", "Speed 4000", "connect", "mem8 0x20000000, 0x10", "q"} {
		if !strings.Contains(script.Text, want) {
			t.Fatalf("expected %q in script:\n%s", want, script.Text)
		}
	}
}

func TestMemoryReadScriptHaltResume(t *testing.T) {
	script, err := MemoryReadScript(MemoryReadOptions{Address: 0x20000000, Length: 4, Width: 32, Halt: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script.Text, "halt\nmem32 0x20000000, 0x4\ngo") {
		t.Fatalf("expected halt/read/go sequence:\n%s", script.Text)
	}
}
