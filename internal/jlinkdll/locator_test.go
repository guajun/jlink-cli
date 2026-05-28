package jlinkdll

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCandidatePathsPrefersExplicitPath(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), libraryName())
	candidates := CandidatePaths(Options{ExplicitPath: explicit, Env: []string{"PATH="}})
	if len(candidates) == 0 {
		t.Fatalf("expected at least one candidate")
	}
	if candidates[0].Path != filepath.Clean(explicit) || candidates[0].Source != "flag" {
		t.Fatalf("unexpected first candidate: %#v", candidates[0])
	}
}

func TestCandidatePathsUsesJLinkPathDirectory(t *testing.T) {
	dir := t.TempDir()
	candidates := CandidatePaths(Options{Env: []string{"JLINK_PATH=" + dir}})
	if len(candidates) == 0 {
		t.Fatalf("expected candidates")
	}
	want := filepath.Join(dir, libraryName())
	if candidates[0].Path != filepath.Clean(want) {
		t.Fatalf("expected %q, got %q", filepath.Clean(want), candidates[0].Path)
	}
}

func TestLibraryNameMatchesPlatform(t *testing.T) {
	name := libraryName()
	if runtime.GOOS == "windows" && name != "JLink_x64.dll" && name != "JLinkARM.dll" {
		t.Fatalf("unexpected windows library name %q", name)
	}
	if runtime.GOOS != "windows" && name == "" {
		t.Fatalf("expected non-empty library name")
	}
}

func TestRequiredSymbolsIncludesMemoryAndBreakpoints(t *testing.T) {
	for _, required := range []string{"JLINKARM_ReadMemEx", "JLINKARM_SetBPEx", "JLINKARM_ClrBPEx"} {
		if !hasSymbol(required) {
			t.Fatalf("missing required symbol %s", required)
		}
	}
}

func hasSymbol(name string) bool {
	for _, symbol := range RequiredSymbols {
		if symbol == name {
			return true
		}
	}
	return false
}
