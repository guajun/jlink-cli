package app

import (
	"bytes"
	"encoding/json"
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
