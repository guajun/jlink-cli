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
	script, err := MemoryReadScript(MemoryReadOptions{TargetOptions: TargetOptions{Device: "NRF52840_XXAA"}, Address: 0x20000000, Length: 4, Width: 32, Halt: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script.Text, "halt\nmem32 0x20000000, 0x4\ngo") {
		t.Fatalf("expected halt/read/go sequence:\n%s", script.Text)
	}
}

func TestFlashScriptSkipsVerifyBinForELF(t *testing.T) {
	script, err := FlashScript(FlashOptions{TargetOptions: TargetOptions{Device: "STM32H750VB"}, File: "build/app.elf", Address: flashAddress(0x08000000), Verify: true})
	if err != nil {
		t.Fatal(err)
	}
	if script.Classification != Destructive || !script.RequiresConfirmation {
		t.Fatalf("unexpected flash script metadata: %#v", script)
	}
	if !strings.Contains(script.Text, "loadfile build/app.elf\n") {
		t.Fatalf("expected loadfile in script:\n%s", script.Text)
	}
	if strings.Contains(script.Text, "verifybin") {
		t.Fatalf("did not expect verifybin for ELF script:\n%s", script.Text)
	}
}

func TestFlashScriptUsesVerifyBinForBIN(t *testing.T) {
	script, err := FlashScript(FlashOptions{TargetOptions: TargetOptions{Device: "STM32H750VB"}, File: "build/app.bin", Address: flashAddress(0x08000000), Verify: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script.Text, "verifybin build/app.bin, 0x8000000") {
		t.Fatalf("expected verifybin for BIN script:\n%s", script.Text)
	}
}

func TestBreakpointScript(t *testing.T) {
	script, err := BreakpointScript(BreakpointOptions{TargetOptions: TargetOptions{Device: "NRF52840_XXAA"}, Action: "set", Address: 0x08000100, Halt: true, Run: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script.Text, "halt\nSetBP 0x8000100\ngo") {
		t.Fatalf("expected halt/set/go sequence:\n%s", script.Text)
	}
}

func TestTargetScriptsRejectMissingDevice(t *testing.T) {
	generators := []func() (Script, error){
		func() (Script, error) { return ConnectScript(TargetOptions{}) },
		func() (Script, error) { return TargetControlScript(TargetOptions{}, "halt") },
		func() (Script, error) { return FlashScript(FlashOptions{File: "firmware.bin"}) },
		func() (Script, error) { return MemoryReadScript(MemoryReadOptions{Length: 1, Width: 8}) },
		func() (Script, error) { return BreakpointScript(BreakpointOptions{Action: "set"}) },
	}
	for _, generate := range generators {
		script, err := generate()
		if err == nil || !strings.Contains(err.Error(), "device") || len(script.Commands) != 0 {
			t.Fatalf("expected no target script without a device: %#v, %v", script, err)
		}
	}
	if _, err := ConnectScript(TargetOptions{Device: "   "}); err == nil {
		t.Fatal("blank device must be rejected")
	}
}

func TestConnectScriptUsesProvidedTarget(t *testing.T) {
	script, err := ConnectScript(TargetOptions{Device: "custom-device", Interface: "jtag", Speed: "1000"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Device custom-device", "SelectInterface JTAG", "Speed 1000"} {
		if !strings.Contains(script.Text, want) {
			t.Fatalf("missing %q: %s", want, script.Text)
		}
	}
}

func flashAddress(value uint64) *uint64 { return &value }

func TestFlashEmbeddedAddresses(t *testing.T) {
	for _, ext := range []string{"elf", "hex", "mot", "s", "s19", "s37", "srec", "ELF", "HEX"} {
		for _, address := range []*uint64{nil, flashAddress(0x12345678)} {
			file := "firmware." + ext
			script, err := FlashScript(FlashOptions{TargetOptions: TargetOptions{Device: "custom"}, File: file, Address: address, Verify: true})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(script.Text, "loadfile "+file+"\n") || strings.Contains(script.Text, "verifybin") {
				t.Fatalf("embedded locations must not be overridden: %s", script.Text)
			}
		}
	}
}

func TestFlashRawAddressPresence(t *testing.T) {
	opts := FlashOptions{TargetOptions: TargetOptions{Device: "custom"}, File: "firmware.bin", Verify: true}
	if script, err := FlashScript(opts); err == nil || len(script.Commands) != 0 {
		t.Fatalf("raw image without address must fail: %#v, %v", script, err)
	}
	opts.Address = flashAddress(0)
	script, err := FlashScript(opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"loadfile firmware.bin, 0x0", "verifybin firmware.bin, 0x0"} {
		if !strings.Contains(script.Text, want) {
			t.Fatalf("zero is a valid explicit address: %s", script.Text)
		}
	}
}
