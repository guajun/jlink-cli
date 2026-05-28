package jlinkdll

import (
	"fmt"
	"strconv"
	"strings"
)

type ConnectOptions struct {
	DLLPath   string
	Device    string
	Interface string
	Speed     string
	DryRun    bool
}

type ConnectPlan struct {
	DLLPath   string   `json:"dll_path"`
	Device    string   `json:"device"`
	Interface string   `json:"interface"`
	Speed     string   `json:"speed"`
	Commands  []string `json:"commands"`
}

type ConnectResult struct {
	DryRun          bool        `json:"dry_run"`
	Plan            ConnectPlan `json:"plan"`
	Connected       bool        `json:"connected"`
	Halted          *bool       `json:"halted,omitempty"`
	ConnectionTouch bool        `json:"connection_touched"`
}

func BuildConnectPlan(opts ConnectOptions) (ConnectPlan, error) {
	if strings.TrimSpace(opts.DLLPath) == "" {
		return ConnectPlan{}, fmt.Errorf("missing J-Link DLL path")
	}
	if strings.TrimSpace(opts.Device) == "" {
		return ConnectPlan{}, fmt.Errorf("missing device")
	}
	if _, err := interfaceID(opts.Interface); err != nil {
		return ConnectPlan{}, err
	}
	if _, err := speedValue(opts.Speed); err != nil {
		return ConnectPlan{}, err
	}
	return ConnectPlan{
		DLLPath:   opts.DLLPath,
		Device:    opts.Device,
		Interface: normalizedInterface(opts.Interface),
		Speed:     normalizedSpeed(opts.Speed),
		Commands: []string{
			"JLINKARM_OpenEx",
			"JLINKARM_ExecCommand(Device = " + opts.Device + ")",
			"JLINKARM_TIF_Select(" + normalizedInterface(opts.Interface) + ")",
			"JLINKARM_SetSpeed(" + normalizedSpeed(opts.Speed) + ")",
			"JLINKARM_Connect",
			"JLINKARM_IsConnected",
			"JLINKARM_IsHalted",
			"JLINKARM_Close",
		},
	}, nil
}

func Connect(opts ConnectOptions) (ConnectResult, error) {
	plan, err := BuildConnectPlan(opts)
	if err != nil {
		return ConnectResult{DryRun: opts.DryRun, Plan: plan}, err
	}
	if opts.DryRun {
		return ConnectResult{DryRun: true, Plan: plan}, nil
	}

	dll, err := Load(opts.DLLPath)
	if err != nil {
		return ConnectResult{Plan: plan}, err
	}
	defer dll.Close()

	result := ConnectResult{DryRun: false, Plan: plan, ConnectionTouch: true}
	if err := dll.open(); err != nil {
		return result, err
	}
	defer dll.closeSession()

	if err := dll.execCommandChecked("Device = " + opts.Device); err != nil {
		return result, err
	}
	interfaceValue, _ := interfaceID(opts.Interface)
	if err := dll.selectInterface(interfaceValue); err != nil {
		return result, err
	}
	speed, _ := speedValue(opts.Speed)
	if err := dll.setSpeed(speed); err != nil {
		return result, err
	}
	if err := dll.connectTarget(); err != nil {
		return result, err
	}
	connected, err := dll.isConnected()
	if err != nil {
		return result, err
	}
	result.Connected = connected
	halted, err := dll.isHalted()
	if err == nil {
		result.Halted = &halted
	}
	return result, nil
}

func (dll *DLL) open() error {
	ret, err := dll.call2("JLINKARM_OpenEx", 0, 0)
	if err != nil {
		return err
	}
	if ret != 0 {
		return fmt.Errorf("JLINKARM_OpenEx returned error pointer 0x%X", ret)
	}
	return nil
}

func (dll *DLL) closeSession() error {
	_, err := dll.call0("JLINKARM_Close")
	return err
}

func (dll *DLL) execCommandChecked(command string) error {
	_, message, err := dll.execCommand(command)
	if err != nil {
		return err
	}
	if strings.TrimSpace(message) != "" {
		return fmt.Errorf(message)
	}
	return nil
}

func (dll *DLL) selectInterface(value uintptr) error {
	ret, err := dll.call1("JLINKARM_TIF_Select", value)
	if err != nil {
		return err
	}
	if int32(ret) != 0 {
		return fmt.Errorf("JLINKARM_TIF_Select returned %d", int32(ret))
	}
	return nil
}

func (dll *DLL) setSpeed(speed uintptr) error {
	_, err := dll.call1("JLINKARM_SetSpeed", speed)
	return err
}

func (dll *DLL) connectTarget() error {
	ret, err := dll.call0("JLINKARM_Connect")
	if err != nil {
		return err
	}
	if int32(ret) < 0 {
		return fmt.Errorf("JLINKARM_Connect returned %d", int32(ret))
	}
	return nil
}

func (dll *DLL) isConnected() (bool, error) {
	ret, err := dll.call0("JLINKARM_IsConnected")
	if err != nil {
		return false, err
	}
	return int32(ret) > 0, nil
}

func (dll *DLL) isHalted() (bool, error) {
	ret, err := dll.call0("JLINKARM_IsHalted")
	if err != nil {
		return false, err
	}
	if int32(ret) < 0 {
		return false, fmt.Errorf("JLINKARM_IsHalted returned %d", int32(ret))
	}
	return int32(ret) > 0, nil
}

func interfaceID(value string) (uintptr, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "SWD":
		return 1, nil
	case "JTAG":
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported target interface %q", value)
	}
}

func normalizedInterface(value string) string {
	if strings.TrimSpace(value) == "" {
		return "SWD"
	}
	return strings.ToUpper(strings.TrimSpace(value))
}

func speedValue(value string) (uintptr, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "auto") {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 32)
	if err != nil {
		return 0, err
	}
	return uintptr(parsed), nil
}

func normalizedSpeed(value string) string {
	if strings.TrimSpace(value) == "" {
		return "auto"
	}
	return strings.TrimSpace(value)
}
