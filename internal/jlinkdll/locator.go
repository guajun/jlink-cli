package jlinkdll

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CandidatePaths(opts Options) []Candidate {
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}
	paths := rawCandidatePaths(opts.ExplicitPath, env)
	candidates := make([]Candidate, 0, len(paths))
	seen := map[string]bool{}
	for _, source := range paths {
		path := filepath.Clean(source.path)
		key := strings.ToLower(path)
		if seen[key] {
			continue
		}
		seen[key] = true
		candidate := Candidate{Path: path, Source: source.source}
		info, err := os.Stat(path)
		if err != nil {
			candidate.Reason = err.Error()
			candidates = append(candidates, candidate)
			continue
		}
		candidate.Exists = !info.IsDir()
		if info.IsDir() {
			candidate.Reason = "path is a directory"
		}
		candidates = append(candidates, candidate)
	}
	return candidates
}

type rawPath struct {
	path   string
	source string
}

func rawCandidatePaths(explicit string, env []string) []rawPath {
	paths := []rawPath{}
	if explicit != "" {
		paths = append(paths, rawPath{path: explicit, source: "flag"})
	}
	if value := getenv(env, "JLINK_DLL_PATH"); value != "" {
		paths = append(paths, rawPath{path: value, source: "env:JLINK_DLL_PATH"})
	}
	if value := getenv(env, "JLINK_PATH"); value != "" {
		paths = append(paths, rawPath{path: libraryPathFromDirOrFile(value), source: "env:JLINK_PATH"})
	}
	for _, dir := range knownDirs(env) {
		paths = append(paths, rawPath{path: filepath.Join(dir, libraryName()), source: "known:" + dir})
	}
	return paths
}

func libraryPathFromDirOrFile(value string) string {
	if strings.HasSuffix(strings.ToLower(value), strings.ToLower(libraryName())) {
		return value
	}
	return filepath.Join(value, libraryName())
}

func libraryName() string {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "386" {
			return "JLinkARM.dll"
		}
		return "JLink_x64.dll"
	case "darwin":
		return "libjlinkarm.dylib"
	default:
		return "libjlinkarm.so"
	}
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
		return []string{"/Applications/SEGGER/JLink", "/usr/local/lib", "/opt/homebrew/lib"}
	default:
		return []string{"/opt/SEGGER/JLink", "/usr/local/lib", "/usr/lib"}
	}
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
