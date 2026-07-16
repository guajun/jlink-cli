package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
			contains: []string{"jlink-cli script memory-read --address <addr> --length <n>", "--halt", "--serial"},
		},
		{
			name:     "script help",
			args:     []string{"script", "--help"},
			contains: []string{"jlink-cli script <subcommand>", "memory-read", "breakpoint-clear"},
		},
		{
			name:     "skill help",
			args:     []string{"skill", "--help"},
			contains: []string{"jlink-cli skill install", "bundled jlink-cli skills"},
		},
		{
			name:     "skill install help",
			args:     []string{"skill", "install", "--help"},
			contains: []string{"jlink-cli skill install [flags]", "--skill", "--agent", "--dir"},
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

func TestSkillInstallDefaultsToKnownUserDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Main([]string{"skill", "install", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	expectedFiles := []string{
		filepath.Join(home, ".copilot", "skills", "jlink-cli", "SKILL.md"),
		filepath.Join(home, ".claude", "skills", "jlink-cli", "SKILL.md"),
		filepath.Join(home, ".codex", "skills", "jlink-cli", "SKILL.md"),
	}
	for _, expectedFile := range expectedFiles {
		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("expected installed skill at %s: %v", expectedFile, err)
		}
		if !strings.Contains(string(content), "name: jlink-cli") {
			t.Fatalf("unexpected installed content at %s: %s", expectedFile, content)
		}
		if strings.Count(string(content), "name: jlink-cli") != 1 || !strings.Contains(string(content), "--serial") {
			t.Fatalf("expected one serial-aware jlink-cli skill at %s", expectedFile)
		}
	}

	var response struct {
		OK     bool `json:"ok"`
		Result []struct {
			Name    string `json:"name"`
			Targets []struct {
				Agent string `json:"agent"`
			} `json:"targets"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if !response.OK || len(response.Result) != 1 || response.Result[0].Name != "jlink-cli" || len(response.Result[0].Targets) != 3 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestSkillInstallAgentAndCustomDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Main([]string{"skill", "install", "--agent", "copilot", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".copilot", "skills", "jlink-cli", "SKILL.md")); err != nil {
		t.Fatalf("expected copilot install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "jlink-cli", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("expected claude install to be skipped, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "skills", "jlink-cli", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("expected codex install to be skipped, got %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Main([]string{"skill", "install", "--agent", "codex", "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "skills", "jlink-cli", "SKILL.md")); err != nil {
		t.Fatalf("expected codex install: %v", err)
	}

	targetDir := t.TempDir()

	stdout.Reset()
	stderr.Reset()
	exitCode = Main([]string{"skill", "install", "--dir", targetDir, "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "jlink-cli", "SKILL.md"))
	if err != nil {
		t.Fatalf("expected custom dir install: %v", err)
	}
	if !strings.Contains(string(content), "name: jlink-cli") {
		t.Fatalf("unexpected custom dir content: %s", content)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Main([]string{"skill", "install", "--skill", "jlink-commander", "--dir", targetDir, "--json"}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, stderr.String())
	}
	content, err = os.ReadFile(filepath.Join(targetDir, "jlink-commander", "SKILL.md"))
	if err != nil {
		t.Fatalf("expected commander custom dir install: %v", err)
	}
	if !strings.Contains(string(content), "name: jlink-commander") {
		t.Fatalf("unexpected commander custom dir content: %s", content)
	}
	if strings.Count(string(content), "name: jlink-commander") != 1 || !strings.Contains(string(content), "-SelectEmuBySN") {
		t.Fatalf("expected one serial-aware jlink-commander skill: %s", content)
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

func TestTargetCommandsAcceptSerial(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "memory read",
			args: []string{"memory", "read", "--address", "0x20000000", "--length", "16", "--serial", "123456789", "--json"},
		},
		{
			name: "breakpoint set",
			args: []string{"script", "breakpoint-set", "--address", "0x08000100", "--serial", "123456789", "--json"},
		},
		{
			name: "breakpoint clear",
			args: []string{"script", "breakpoint-clear", "--address", "0x08000100", "--serial", "123456789", "--json"},
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
