package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"version", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Version string `json:"version"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Version != "0.3.0-dev" {
		t.Fatalf("expected ok response, got %#v", response)
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"nope", "--json"}, &stdout, &stderr)
	if exitCode != ExitUsage {
		t.Fatalf("expected usage exit, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected non-json unknown command to write no stdout, got %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatalf("expected stderr error")
	}
}

func TestNestedCommandHelp(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "target long help",
			args:     []string{"target", "--help"},
			contains: []string{"jlink-cli target <reset|halt|run>", "reset", "halt", "run"},
		},
		{
			name:     "target short help",
			args:     []string{"target", "-h"},
			contains: []string{"jlink-cli target <reset|halt|run>", "--device", "--yes"},
		},
		{
			name:     "target action help",
			args:     []string{"target", "run", "--help"},
			contains: []string{"jlink-cli target run [flags]", "--interface", "--timeout"},
		},
		{
			name:     "flash help",
			args:     []string{"flash", "--help"},
			contains: []string{"jlink-cli flash --file <path>", "jlink-cli flash program --file <path>", "--verify"},
		},
		{
			name:     "flash program help",
			args:     []string{"flash", "program", "-h"},
			contains: []string{"jlink-cli script flash --file <path>", "--address", "--reset"},
		},
		{
			name:     "memory help",
			args:     []string{"memory", "--help"},
			contains: []string{"jlink-cli memory read --address <addr> --length <n>", "--width"},
		},
		{
			name:     "memory read help",
			args:     []string{"memory", "read", "--help"},
			contains: []string{"jlink-cli script memory-read --address <addr> --length <n>", "--halt", "--serial"},
		},
		{
			name:     "script help",
			args:     []string{"script", "--help"},
			contains: []string{"jlink-cli script <subcommand>", "memory-read", "breakpoint-clear"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := Main(tt.args, &stdout, &stderr)
			if exitCode != ExitOK {
				t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got %q", stderr.String())
			}
			for _, want := range tt.contains {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("expected stdout to contain %q, got:\n%s", want, stdout.String())
				}
			}
		})
	}
}

func TestSeparateSkillInstallation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--help"}, &stdout, &stderr); code != ExitOK {
		t.Fatalf("help failed: %s", &stderr)
	}
	for _, want := range []string{"installed separately", "gh skill install guajun/jlink-cli jlink-cli --agent codex --scope user", "jlink-commander"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("missing %q in help: %s", want, &stdout)
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := Main([]string{"skill", "install"}, &stdout, &stderr); code != ExitUsage {
		t.Fatalf("removed command returned %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("unexpected error: %s", &stderr)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Main([]string{"inspect", "--json"}, &stdout, &stderr); code != ExitOK {
		t.Fatal(&stderr)
	}
	var response struct {
		Result struct {
			Commands []string `json:"commands"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	for _, command := range response.Result.Commands {
		if command == "skill" {
			t.Fatal("inspect still advertises skill")
		}
	}
}

func TestRunParsesInputWithoutExclusiveDeviceUse(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"run", "--input", `{"action":"ping"}`}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
			Note   string `json:"note"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Status != "accepted" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Result.Note == "" {
		t.Fatalf("expected non-exclusive prototype note")
	}
}

func TestScriptProbeJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"script", "probe", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Classification       string   `json:"classification"`
			RequiresConfirmation bool     `json:"requires_confirmation"`
			Commands             []string `json:"commands"`
			Text                 string   `json:"text"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Classification != "probe-only" || response.Result.RequiresConfirmation {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Result.Text == "" || len(response.Result.Commands) == 0 {
		t.Fatalf("expected generated script")
	}
}

func TestScriptMemoryReadRequiresConfirmation(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"script", "memory-read", "--device", "NRF52840_XXAA", "--address", "0x20000000", "--length", "16", "--width", "8", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Classification       string `json:"classification"`
				RequiresConfirmation bool   `json:"requires_confirmation"`
				Text                 string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Script.Classification != "target-session" || !response.Result.Script.RequiresConfirmation {
		t.Fatalf("unexpected response: %#v", response)
	}
	if !strings.Contains(response.Result.Script.Text, "mem8 0x20000000, 0x10") {
		t.Fatalf("expected mem8 command: %s", response.Result.Script.Text)
	}
}

func TestTargetCommandsAcceptSerial(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "memory read",
			args: []string{"memory", "read", "--device", "NRF52840_XXAA", "--address", "0x20000000", "--length", "16", "--serial", "123456789", "--json"},
		},
		{
			name: "breakpoint set",
			args: []string{"script", "breakpoint-set", "--device", "NRF52840_XXAA", "--address", "0x08000100", "--serial", "123456789", "--json"},
		},
		{
			name: "breakpoint clear",
			args: []string{"script", "breakpoint-clear", "--device", "NRF52840_XXAA", "--address", "0x08000100", "--serial", "123456789", "--json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := Main(tt.args, &stdout, &stderr)
			if exitCode != ExitOK {
				t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
			}
			if !strings.Contains(stdout.String(), `"ok": true`) {
				t.Fatalf("unexpected response: %s", stdout.String())
			}
		})
	}
}

func TestScriptConnectJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"script", "connect", "--device", "NRF52840_XXAA", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				RequiresConfirmation bool   `json:"requires_confirmation"`
				Text                 string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || !response.Result.Script.RequiresConfirmation || !strings.Contains(response.Result.Script.Text, "connect") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestScriptFlashJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"script", "flash", "--device", "NRF52840_XXAA", "--address", "0", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Classification string `json:"classification"`
				Text           string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Script.Classification != "destructive" || !strings.Contains(response.Result.Script.Text, "loadfile build/app.elf") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestRunStructuredFlashPlan(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"run", "--json", "--input", `{"action":"flash","arguments":{"device":"NRF52840_XXAA","file":"build/app.elf","address":"0","interface":"JTAG","speed":"1000","serial":"987654321"}}`}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Text string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || !strings.Contains(response.Result.Script.Text, "loadfile build/app.elf") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestFlashProgramAliasJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"flash", "program", "--device", "NRF52840_XXAA", "--address", "0", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Classification string `json:"classification"`
				Text           string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Script.Classification != "destructive" || !strings.Contains(response.Result.Script.Text, "loadfile build/app.elf") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestFlashDirectJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"flash", "--device", "NRF52840_XXAA", "--address", "0", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Classification string `json:"classification"`
				Text           string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || response.Result.Script.Classification != "destructive" || !strings.Contains(response.Result.Script.Text, "loadfile build/app.elf") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestMemoryReadAliasJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"memory", "read", "--device", "NRF52840_XXAA", "--address", "0x20000000", "--length", "16", "--width", "8", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Text string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || !strings.Contains(response.Result.Script.Text, "mem8 0x20000000, 0x10") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestTargetRunAliasJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"target", "run", "--device", "NRF52840_XXAA", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Script struct {
				Text string `json:"text"`
			} `json:"script"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || !strings.Contains(response.Result.Script.Text, "go\nq") {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestTargetCommandsRequireDevice(t *testing.T) {
	commands := [][]string{
		{"connect", "--dry-run"}, {"script", "connect"},
		{"script", "flash", "--file", "firmware.bin", "--address", "0"},
		{"flash", "--file", "firmware.bin", "--address", "0"},
		{"memory", "read", "--address", "0", "--length", "1"},
		{"script", "memory-read", "--address", "0", "--length", "1"},
		{"target", "halt"}, {"target", "run"}, {"target", "reset"},
		{"script", "breakpoint-set", "--address", "0"},
		{"script", "breakpoint-clear", "--address", "0"},
	}
	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := Main(append(args, "--json"), &out, &errOut); code != ExitUsage {
				t.Fatalf("expected usage error, got %d: %s %s", code, &out, &errOut)
			}
			if !strings.Contains(out.String(), "device") || strings.Contains(out.String(), "STM32H750VB") {
				t.Fatalf("expected missing-device diagnostic without fallback: %s", &out)
			}
		})
	}
}

func TestRawFlashRequiresExplicitAddress(t *testing.T) {
	for _, prefix := range [][]string{{"flash"}, {"flash", "program"}, {"script", "flash"}} {
		var out, errOut bytes.Buffer
		args := append(prefix, "--device", "NRF52840_XXAA", "--file", "firmware.bin", "--json")
		if code := Main(args, &out, &errOut); code != ExitUsage || !strings.Contains(out.String(), "usage.missing_address") {
			t.Fatalf("expected missing address, got %d: %s %s", code, &out, &errOut)
		}
	}
}

func TestRunDoesNotInferFlashSettings(t *testing.T) {
	for _, input := range []string{
		`{"request":"flash firmware.bin to another board"}`,
		`{"request":"do not flash firmware.bin"}`,
		`{"action":"flash","arguments":{"file":"firmware.bin","address":"0"}}`,
		`{"action":"flash","arguments":{"file":"firmware.bin","device":"NRF52840_XXAA"}}`,
	} {
		var out, errOut bytes.Buffer
		if code := Main([]string{"run", "--input", input, "--yes", "--json"}, &out, &errOut); code != ExitUsage {
			t.Fatalf("expected usage error before execution, got %d: %s %s", code, &out, &errOut)
		}
	}
}

func TestFlashAddressedImagesWithoutAddress(t *testing.T) {
	for _, ext := range []string{"elf", "hex", "mot", "s", "s19", "s37", "srec", "ELF"} {
		file := "build/firmware." + ext
		commands := [][]string{
			{"flash", "--device", "custom", "--file", file},
			{"flash", "program", "--device", "custom", "--file", file},
			{"script", "flash", "--device", "custom", "--file", file},
			{"run", "--input", `{"action":"flash","arguments":{"device":"custom","file":"` + file + `"}}`},
		}
		for _, args := range commands {
			var out, errOut bytes.Buffer
			if code := Main(append(args, "--json"), &out, &errOut); code != ExitOK {
				t.Fatalf("expected embedded-address plan: %v: %d %s %s", args, code, &out, &errOut)
			}
			var response struct {
				Result struct {
					Script struct {
						Commands []string `json:"commands"`
					} `json:"script"`
				} `json:"result"`
			}
			if err := json.Unmarshal(out.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, command := range response.Result.Script.Commands {
				if command == "loadfile "+file {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected loadfile without an address: %s", &out)
			}
		}
	}
}

func TestFlashIgnoredAddressDiagnostic(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Main([]string{"flash", "--device", "custom", "--file", "firmware.elf", "--address", "0x1234", "--json"}, &out, &errOut)
	if code != ExitOK || !strings.Contains(out.String(), "flash.address_ignored") || strings.Contains(out.String(), "loadfile firmware.elf,") {
		t.Fatalf("expected embedded locations and ignored-address diagnostic: %d %s %s", code, &out, &errOut)
	}
}

func TestRawFlashZeroAndInvalidAddress(t *testing.T) {
	for _, address := range []string{"0", "invalid"} {
		var out, errOut bytes.Buffer
		code := Main([]string{"flash", "--device", "custom", "--file", "firmware.bin", "--address", address, "--json"}, &out, &errOut)
		if address == "0" {
			if code != ExitOK || !strings.Contains(out.String(), "loadfile firmware.bin, 0x0") {
				t.Fatalf("zero address failed: %s %s", &out, &errOut)
			}
		} else if code != ExitUsage {
			t.Fatalf("invalid address accepted: %s", &out)
		}
	}
}
