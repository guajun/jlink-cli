package scriptgen

import (
	"fmt"
	"strconv"
	"strings"
)

type Classification string

const (
	ProbeOnly     Classification = "probe-only"
	TargetSession Classification = "target-session"
	Destructive   Classification = "destructive"
)

type Script struct {
	Classification       Classification `json:"classification"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
	Commands             []string       `json:"commands"`
	Text                 string         `json:"text"`
}

type TargetOptions struct {
	Device    string
	Interface string
	Speed     string
}

type MemoryReadOptions struct {
	TargetOptions
	Address uint64
	Length  uint64
	Width   uint
	Halt    bool
}

func ProbeScript() Script {
	return build(ProbeOnly, []string{"ExitOnError 1", "ShowEmuList", "q"})
}

func MemoryReadScript(opts MemoryReadOptions) (Script, error) {
	if opts.Length == 0 {
		return Script{}, fmt.Errorf("length must be greater than zero")
	}
	if opts.Width != 8 && opts.Width != 16 && opts.Width != 32 {
		return Script{}, fmt.Errorf("width must be one of 8, 16, or 32")
	}
	commands := targetPrelude(opts.TargetOptions)
	if opts.Halt {
		commands = append(commands, "halt")
	}
	commands = append(commands, fmt.Sprintf("mem%d %s, %s", opts.Width, formatHex(opts.Address), formatHex(opts.Length)))
	if opts.Halt {
		commands = append(commands, "go")
	}
	commands = append(commands, "q")
	return build(TargetSession, commands), nil
}

func targetPrelude(opts TargetOptions) []string {
	return []string{
		"ExitOnError 1",
		"Device " + defaultString(opts.Device, "STM32H750VB"),
		"SelectInterface " + strings.ToUpper(defaultString(opts.Interface, "SWD")),
		"Speed " + defaultString(opts.Speed, "4000"),
		"connect",
	}
}

func build(classification Classification, commands []string) Script {
	return Script{
		Classification:       classification,
		RequiresConfirmation: classification != ProbeOnly,
		Commands:             commands,
		Text:                 strings.Join(commands, "\n") + "\n",
	}
}

func formatHex(value uint64) string {
	return "0x" + strings.ToUpper(strconv.FormatUint(value, 16))
}

func defaultString(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
