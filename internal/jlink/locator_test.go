package jlink

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscoverSelectsExplicitExecutable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, executableNames()[0])
	if err := os.WriteFile(path, []byte("test"), 0755); err != nil {
		t.Fatal(err)
	}

	report := Discover(Options{ExplicitPath: path, Env: []string{"PATH="}})
	if report.Selected != path {
		t.Fatalf("expected selected %q, got %q", path, report.Selected)
	}
	if len(report.Candidates) == 0 || !report.Candidates[0].Executable {
		t.Fatalf("expected executable candidate, got %#v", report.Candidates)
	}
}

func TestDiscoverReportsMissingCandidate(t *testing.T) {
	missing := filepath.Join(t.TempDir(), executableNames()[0])
	report := Discover(Options{ExplicitPath: missing, Env: []string{"PATH="}})
	if report.Selected != "" {
		t.Fatalf("expected no selected path, got %q", report.Selected)
	}
	if len(report.Candidates) == 0 || report.Candidates[0].Exists {
		t.Fatalf("expected missing candidate, got %#v", report.Candidates)
	}
}

func TestDiscoverRejectsNonExecutableOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows treats existing .exe files as executable for this locator")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, executableNames()[0])
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	report := Discover(Options{ExplicitPath: path, Env: []string{"PATH="}})
	if report.Selected != "" {
		t.Fatalf("expected no selected path, got %q", report.Selected)
	}
	if report.Candidates[0].Executable {
		t.Fatalf("expected candidate not executable")
	}
}

func TestAcceptPATHDirRejectsPlainWindowsJDKPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific PATH filtering")
	}
	if acceptPATHDir(`C:\Program Files\Microsoft\jdk-21\bin`) {
		t.Fatalf("expected JDK bin path to be ignored")
	}
	if !acceptPATHDir(`C:\Program Files\SEGGER\JLink`) {
		t.Fatalf("expected SEGGER JLink path to be accepted")
	}
}
