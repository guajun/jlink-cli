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

	var response map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if response["ok"] != true {
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
			contains: []string{"jlink-cli script memory-read --address <addr> --length <n>", "--halt"},
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

	exitCode := Main([]string{"script", "memory-read", "--address", "0x20000000", "--length", "16", "--width", "8", "--json"}, &stdout, &stderr)
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

func TestScriptConnectJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"script", "connect", "--json"}, &stdout, &stderr)
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

	exitCode := Main([]string{"script", "flash", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
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

func TestRunAcceptsYesForFlashPlan(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Main([]string{"run", "--json", "--input", `{"request":"flash build/app.elf to STM32H750VB"}`}, &stdout, &stderr)
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

	exitCode := Main([]string{"flash", "program", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
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

	exitCode := Main([]string{"flash", "--file", "build/app.elf", "--json"}, &stdout, &stderr)
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

	exitCode := Main([]string{"memory", "read", "--address", "0x20000000", "--length", "16", "--width", "8", "--json"}, &stdout, &stderr)
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

	exitCode := Main([]string{"target", "run", "--json"}, &stdout, &stderr)
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
