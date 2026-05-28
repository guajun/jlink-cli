package jlink

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildFlashScript(t *testing.T) {
	script, err := BuildFlashScript(FlashOptions{File: `C:\firmware.bin`, Address: 0x08000000, Verify: true, Reset: true})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"connect", "r", `loadfile C:\firmware.bin 0x8000000`, `verifybin C:\firmware.bin 0x8000000`, "r", "g", "q"}
	if !reflect.DeepEqual(script, expected) {
		t.Fatalf("unexpected script:\nwant %#v\n got %#v", expected, script)
	}
}

func TestBuildLiveMemoryReadScript(t *testing.T) {
	script, err := BuildMemoryReadScript(MemoryReadOptions{Address: 0x20000000, Length: 16, Width: 8, Halt: false})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"connect", "mem8 0x20000000 16", "q"}
	if !reflect.DeepEqual(script, expected) {
		t.Fatalf("unexpected script: %#v", script)
	}
}

func TestBuildHaltedMemoryReadScript(t *testing.T) {
	script, err := BuildMemoryReadScript(MemoryReadOptions{Address: 0x20000000, Length: 4, Width: 32, Halt: true})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"connect", "h", "mem32 0x20000000 4", "g", "q"}
	if !reflect.DeepEqual(script, expected) {
		t.Fatalf("unexpected script: %#v", script)
	}
}

func TestBuildBreakpointScript(t *testing.T) {
	script, err := BuildBreakpointScript(BreakpointOptions{Action: "set", Address: 0x08000100, Halt: true, Run: true})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"connect", "h", "SetBP 0x8000100", "g", "q"}
	if !reflect.DeepEqual(script, expected) {
		t.Fatalf("unexpected script: %#v", script)
	}
}

func TestBuildCallStackSpec(t *testing.T) {
	result, err := BuildCallStackSpec(CallStackOptions{
		CommonOptions: CommonOptions{Executable: `C:\Program Files\SEGGER\JLink\JLink.exe`, Device: "STM32H750VB", Interface: "SWD", Speed: "4000", Timeout: time.Second},
		ELF:           `C:\firmware.elf`,
		MaxFrames:     8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.GDBServer.Executable == "" || result.GDB.Executable == "" {
		t.Fatalf("expected GDB server and GDB command specs")
	}
	if !contains(result.GDB.Args, "bt 8") {
		t.Fatalf("expected bounded backtrace args, got %#v", result.GDB.Args)
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
