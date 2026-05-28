package jlink

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	ExplicitPath string
	Env          []string
}

type Candidate struct {
	Path       string `json:"path"`
	Source     string `json:"source"`
	Exists     bool   `json:"exists"`
	Executable bool   `json:"executable"`
	Reason     string `json:"reason,omitempty"`
}

type Report struct {
	Selected    string            `json:"selected,omitempty"`
	Candidates  []Candidate       `json:"candidates"`
	Environment map[string]string `json:"environment"`
	Platform    map[string]string `json:"platform"`
}

func Discover(opts Options) Report {
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}

	paths := candidatePaths(opts.ExplicitPath, env)
	candidates := make([]Candidate, 0, len(paths))
	seen := map[string]bool{}
	selected := ""

	for _, item := range paths {
		clean := filepath.Clean(item.path)
		key := strings.ToLower(clean)
		if seen[key] {
			continue
		}
		seen[key] = true

		candidate := probe(clean, item.source)
		if selected == "" && candidate.Exists && candidate.Executable {
			selected = candidate.Path
		}
		candidates = append(candidates, candidate)
	}

	return Report{
		Selected:   selected,
		Candidates: candidates,
		Environment: map[string]string{
			"JLINK_CLI_PATH": getenv(env, "JLINK_CLI_PATH"),
			"JLINK_PATH":     getenv(env, "JLINK_PATH"),
			"PATH":           getenv(env, "PATH"),
		},
		Platform: map[string]string{
			"goos":   runtime.GOOS,
			"goarch": runtime.GOARCH,
		},
	}
}

type pathSource struct {
	path   string
	source string
}

func candidatePaths(explicit string, env []string) []pathSource {
	paths := []pathSource{}
	if explicit != "" {
		paths = append(paths, pathSource{path: explicit, source: "flag"})
	}
	if value := getenv(env, "JLINK_CLI_PATH"); value != "" {
		paths = append(paths, pathSource{path: value, source: "env:JLINK_CLI_PATH"})
	}
	if value := getenv(env, "JLINK_PATH"); value != "" {
		paths = append(paths, pathSource{path: value, source: "env:JLINK_PATH"})
	}

	for _, dir := range knownDirs(env) {
		for _, name := range executableNames() {
			paths = append(paths, pathSource{path: filepath.Join(dir, name), source: "known:" + dir})
		}
	}

	for _, dir := range filepath.SplitList(getenv(env, "PATH")) {
		if dir == "" {
			continue
		}
		if !acceptPATHDir(dir) {
			continue
		}
		for _, name := range executableNames() {
			paths = append(paths, pathSource{path: filepath.Join(dir, name), source: "PATH"})
		}
	}

	return paths
}

func executableNames() []string {
	if runtime.GOOS == "windows" {
		return []string{"JLink.exe", "JLinkCL.exe", "JLinkExe.exe", "JLinkGDBServerCL.exe"}
	}
	return []string{"JLinkExe", "JLink", "JLinkGDBServerCLExe"}
}

func knownDirs(env []string) []string {
	switch runtime.GOOS {
	case "windows":
		dirs := []string{}
		for _, key := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
			if root := getenv(env, key); root != "" {
				dirs = append(dirs, filepath.Join(root, "SEGGER", "JLink"))
			}
		}
		return dirs
	case "darwin":
		return []string{"/Applications/SEGGER/JLink", "/opt/homebrew/bin", "/usr/local/bin"}
	default:
		return []string{"/opt/SEGGER/JLink", "/usr/local/bin", "/usr/bin"}
	}
}

func acceptPATHDir(dir string) bool {
	if runtime.GOOS != "windows" {
		return true
	}
	lower := strings.ToLower(filepath.Clean(dir))
	return strings.Contains(lower, "segger") || strings.Contains(lower, "jlink")
}

func probe(path string, source string) Candidate {
	candidate := Candidate{Path: path, Source: source}
	info, err := os.Stat(path)
	if err != nil {
		candidate.Reason = err.Error()
		return candidate
	}
	candidate.Exists = true
	if info.IsDir() {
		candidate.Reason = "path is a directory"
		return candidate
	}
	if runtime.GOOS == "windows" {
		candidate.Executable = true
		candidate.Reason = "ok"
		return candidate
	}
	candidate.Executable = info.Mode()&0111 != 0
	if candidate.Executable {
		candidate.Reason = "ok"
	} else {
		candidate.Reason = "file is not executable"
	}
	return candidate
}

func getenv(env []string, key string) string {
	for _, item := range env {
		name, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if name == key || (runtime.GOOS == "windows" && strings.EqualFold(name, key)) {
			return value
		}
	}
	return ""
}
