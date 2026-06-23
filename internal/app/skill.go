package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guajun/jlink-cli/internal/protocol"
	bundledskills "github.com/guajun/jlink-cli/skills"
)

const defaultSkillName = "jlink-cli"

var builtinSkills = map[string]string{
	"jlink-cli":       "jlink-cli/SKILL.md",
	"jlink-commander": "jlink-commander/SKILL.md",
}

type skillInstallResult struct {
	Name    string               `json:"name"`
	Source  string               `json:"source"`
	Targets []skillInstallTarget `json:"targets"`
}

type skillInstallTarget struct {
	Agent       string `json:"agent"`
	Directory   string `json:"directory"`
	SkillFile   string `json:"skill_file"`
	Overwritten bool   `json:"overwritten"`
}

type skillTargetRoot struct {
	Agent string
	Root  string
}

func runSkill(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	if len(args) == 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.missing_subcommand", Message: "skill requires subcommand: install", Exit: ExitUsage}
	}
	if isHelpArg(args[0]) {
		return ExitOK, skillUsage(), nil, nil
	}
	if args[0] != "install" {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unknown_subcommand", Message: fmt.Sprintf("unknown skill subcommand %q", args[0]), Exit: ExitUsage}
	}
	if hasHelpArg(args[1:]) {
		return ExitOK, skillInstallUsage(), nil, nil
	}
	return runSkillInstall(args[1:], jsonOutput)
}

func runSkillInstall(args []string, jsonOutput *bool) (int, any, []protocol.Diagnostic, error) {
	fs := newFlagSet("skill install")
	skillName := fs.String("skill", defaultSkillName, "bundled skill to install: jlink-cli, jlink-commander, or all")
	agent := fs.String("agent", "all", "target agent: all, github-copilot/copilot, claude-code/claude, or codex")
	customDir := fs.String("dir", "", "custom skills directory; bypasses the built-in agent paths")
	fs.BoolVar(jsonOutput, "json", true, "write JSON output")
	if err := fs.Parse(args); err != nil {
		return ExitUsage, nil, nil, usageError(err)
	}
	if fs.NArg() != 0 {
		return ExitUsage, nil, nil, &cliError{Code: "usage.unexpected_argument", Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0)), Exit: ExitUsage}
	}

	skills, err := loadBuiltinSkills(*skillName)
	if err != nil {
		return ExitUsage, nil, nil, err
	}
	targets, err := skillTargetRoots(*agent, *customDir)
	if err != nil {
		return ExitUsage, nil, nil, err
	}

	var result []skillInstallResult
	for _, skill := range skills {
		skillResult := skillInstallResult{Name: skill.Name, Source: "builtin:" + skill.Name}
		for _, target := range targets {
			installed, err := installSkillToRoot(skill, target)
			if err != nil {
				return ExitRuntime, result, nil, err
			}
			skillResult.Targets = append(skillResult.Targets, installed)
		}
		result = append(result, skillResult)
	}

	return ExitOK, result, []protocol.Diagnostic{{Level: "info", Code: "skill.installed", Message: "installed bundled agent skill"}}, nil
}

type bundledSkill struct {
	Name    string
	Content []byte
}

func loadBuiltinSkills(name string) ([]bundledSkill, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		normalized = defaultSkillName
	}
	if normalized == "all" {
		return loadBuiltinSkillsByName([]string{"jlink-cli", "jlink-commander"})
	}
	if _, ok := builtinSkills[normalized]; !ok {
		return nil, &cliError{Code: "usage.unsupported_skill", Message: fmt.Sprintf("unsupported skill %q", name), Exit: ExitUsage}
	}
	return loadBuiltinSkillsByName([]string{normalized})
}

func loadBuiltinSkillsByName(names []string) ([]bundledSkill, error) {
	var skills []bundledSkill
	for _, name := range names {
		content, err := bundledskills.FS.ReadFile(builtinSkills[name])
		if err != nil {
			return nil, &cliError{Code: "skill.builtin_unavailable", Message: err.Error(), Exit: ExitRuntime}
		}
		skills = append(skills, bundledSkill{Name: name, Content: content})
	}
	return skills, nil
}

func skillTargetRoots(agent string, customDir string) ([]skillTargetRoot, error) {
	trimmedDir := strings.TrimSpace(customDir)
	if trimmedDir != "" {
		return []skillTargetRoot{{Agent: "custom", Root: trimmedDir}}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, &cliError{Code: "skill.home_unavailable", Message: err.Error(), Exit: ExitRuntime}
	}

	switch normalizeSkillAgent(agent) {
	case "all":
		return []skillTargetRoot{
			{Agent: "github-copilot", Root: filepath.Join(home, ".copilot", "skills")},
			{Agent: "claude-code", Root: filepath.Join(home, ".claude", "skills")},
			{Agent: "codex", Root: filepath.Join(home, ".codex", "skills")},
		}, nil
	case "github-copilot":
		return []skillTargetRoot{{Agent: "github-copilot", Root: filepath.Join(home, ".copilot", "skills")}}, nil
	case "claude-code":
		return []skillTargetRoot{{Agent: "claude-code", Root: filepath.Join(home, ".claude", "skills")}}, nil
	case "codex":
		return []skillTargetRoot{{Agent: "codex", Root: filepath.Join(home, ".codex", "skills")}}, nil
	default:
		return nil, &cliError{Code: "usage.unsupported_agent", Message: fmt.Sprintf("unsupported agent %q", agent), Exit: ExitUsage}
	}
}

func normalizeSkillAgent(agent string) string {
	switch strings.ToLower(strings.TrimSpace(agent)) {
	case "", "all":
		return "all"
	case "github-copilot", "copilot":
		return "github-copilot"
	case "claude-code", "claude":
		return "claude-code"
	case "codex":
		return "codex"
	default:
		return ""
	}
}

func installSkillToRoot(skill bundledSkill, target skillTargetRoot) (skillInstallTarget, error) {
	destinationDir := filepath.Join(target.Root, skill.Name)
	skillFile := filepath.Join(destinationDir, "SKILL.md")
	_, statErr := os.Stat(destinationDir)
	exists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return skillInstallTarget{}, &cliError{Code: "skill.stat_failed", Message: statErr.Error(), Exit: ExitRuntime, Details: map[string]string{"path": destinationDir}}
	}
	if exists {
		if err := os.RemoveAll(destinationDir); err != nil {
			return skillInstallTarget{}, &cliError{Code: "skill.remove_failed", Message: err.Error(), Exit: ExitRuntime, Details: map[string]string{"path": destinationDir}}
		}
	}
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return skillInstallTarget{}, &cliError{Code: "skill.mkdir_failed", Message: err.Error(), Exit: ExitRuntime, Details: map[string]string{"path": destinationDir}}
	}
	if err := os.WriteFile(skillFile, skill.Content, 0o644); err != nil {
		return skillInstallTarget{}, &cliError{Code: "skill.write_failed", Message: err.Error(), Exit: ExitRuntime, Details: map[string]string{"path": skillFile}}
	}

	return skillInstallTarget{Agent: target.Agent, Directory: destinationDir, SkillFile: skillFile, Overwritten: exists}, nil
}

func skillUsage() string {
	return "Usage:\n" +
		"  jlink-cli skill install [flags]\n" +
		"\n" +
		"Subcommands:\n" +
		"  install  Install bundled jlink-cli skills for local agent hosts\n"
}

func skillInstallUsage() string {
	return "Usage:\n" +
		"  jlink-cli skill install [flags]\n" +
		"\n" +
		"Flags:\n" +
		"  --skill <name>      Bundled skill: jlink-cli, jlink-commander, or all (default jlink-cli)\n" +
		"  --agent <name>      Target agent: all, github-copilot/copilot, claude-code/claude, or codex (default all)\n" +
		"  --dir <path>        Custom skills directory; bypasses the built-in agent paths\n" +
		"  --json              Write JSON output (default true)\n"
}
