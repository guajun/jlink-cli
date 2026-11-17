package scriptgen

import (
	"fmt"
	"path/filepath"
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

type FlashOptions struct {
	TargetOptions
	File    string
	Address *uint64 // nil uses embedded image addresses; raw images require an explicit address
	Verify  bool
	Reset   bool
}

type BreakpointOptions struct {
	TargetOptions
	Action  string
	Address uint64
	Halt    bool
	Run     bool
}

func ProbeScript() Script {
	return build(ProbeOnly, []string{"ExitOnError 1", "ShowEmuList", "q"})
}

func ConnectScript(opts TargetOptions) (Script, error) {
	commands, err := targetPrelude(opts)
	if err != nil {
		return Script{}, err
	}
	commands = append(commands, "q")
	return build(TargetSession, commands), nil
}

func TargetControlScript(opts TargetOptions, action string) (Script, error) {
	normalized := strings.ToLower(strings.TrimSpace(action))
	if normalized != "reset" && normalized != "halt" && normalized != "go" {
		return Script{}, fmt.Errorf("target action must be reset, halt, or go")
	}
	commands, err := targetPrelude(opts)
	if err != nil {
		return Script{}, err
	}
	commands = append(commands, normalized, "q")
	return build(TargetSession, commands), nil
}

func FlashScript(opts FlashOptions) (Script, error) {
	if strings.TrimSpace(opts.File) == "" {
		return Script{}, fmt.Errorf("missing firmware file")
	}
	commands, err := targetPrelude(opts.TargetOptions)
	if err != nil {
		return Script{}, err
	}
	if opts.Reset {
		commands = append(commands, "reset")
	}
	load := "loadfile " + opts.File
	if !HasEmbeddedAddresses(opts.File) {
		if opts.Address == nil {
			return Script{}, fmt.Errorf("address is required for raw images; resolve it from project configuration or provide --address")
		}
		load += ", " + formatHex(*opts.Address)
	}
	commands = append(commands, load)
	if opts.Verify && canVerifyBin(opts.File) {
		commands = append(commands, fmt.Sprintf("verifybin %s, %s", opts.File, formatHex(*opts.Address)))
	}
	commands = append(commands, "q")
	return build(Destructive, commands), nil
}

// HasEmbeddedAddresses identifies the addressed formats supported by Commander.
// Their load locations come from the image, not the optional loadfile address.
func HasEmbeddedAddresses(file string) bool {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(file))) {
	case ".elf", ".hex", ".mot", ".s", ".s19", ".s37", ".srec":
		return true
	default:
		return false
	}
}

func canVerifyBin(file string) bool {
	return strings.EqualFold(filepath.Ext(strings.TrimSpace(file)), ".bin")
}

func MemoryReadScript(opts MemoryReadOptions) (Script, error) {
	if opts.Length == 0 {
		return Script{}, fmt.Errorf("length must be greater than zero")
	}
	if opts.Width != 8 && opts.Width != 16 && opts.Width != 32 {
		return Script{}, fmt.Errorf("width must be one of 8, 16, or 32")
	}
	commands, err := targetPrelude(opts.TargetOptions)
	if err != nil {
		return Script{}, err
	}
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

func BreakpointScript(opts BreakpointOptions) (Script, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action != "set" && action != "clear" {
		return Script{}, fmt.Errorf("breakpoint action must be set or clear")
	}
	commands, err := targetPrelude(opts.TargetOptions)
	if err != nil {
		return Script{}, err
	}
	if opts.Halt {
		commands = append(commands, "halt")
	}
	if action == "set" {
		commands = append(commands, "SetBP "+formatHex(opts.Address))
	} else {
		commands = append(commands, "ClearBP "+formatHex(opts.Address))
	}
	if opts.Run {
		commands = append(commands, "go")
	}
	commands = append(commands, "q")
	return build(TargetSession, commands), nil
}

func targetPrelude(opts TargetOptions) ([]string, error) {
	if strings.TrimSpace(opts.Device) == "" {
		return nil, fmt.Errorf("device is required; specify --device using the project's J-Link device name")
	}
	return []string{
		"ExitOnError 1",
		"Device " + strings.TrimSpace(opts.Device),
		"SelectInterface " + strings.ToUpper(defaultString(opts.Interface, "SWD")),
		"Speed " + defaultString(opts.Speed, "4000"),
		"connect",
	}, nil
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
