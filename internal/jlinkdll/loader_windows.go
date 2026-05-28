//go:build windows

package jlinkdll

import (
	"fmt"
	"syscall"
	"unsafe"
)

func ProbeSymbols(path string, names []string) ([]SymbolStatus, error) {
	dll, err := syscall.LoadDLL(path)
	if err != nil {
		return nil, err
	}
	defer dll.Release()

	statuses := make([]SymbolStatus, 0, len(names))
	for _, name := range names {
		_, err := dll.FindProc(name)
		status := SymbolStatus{Name: name, Present: err == nil}
		if err != nil {
			status.Reason = err.Error()
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func Load(path string) (*DLL, error) {
	dll, err := syscall.LoadDLL(path)
	if err != nil {
		return nil, err
	}
	return &DLL{path: path, dll: dll}, nil
}

type DLL struct {
	path string
	dll  *syscall.DLL
}

func (dll *DLL) Path() string {
	return dll.path
}

func (dll *DLL) Close() error {
	if dll == nil || dll.dll == nil {
		return nil
	}
	err := dll.dll.Release()
	dll.dll = nil
	return err
}

func (dll *DLL) proc(name string) (*syscall.Proc, error) {
	if dll == nil || dll.dll == nil {
		return nil, fmt.Errorf("J-Link DLL is not loaded")
	}
	return dll.dll.FindProc(name)
}

func (dll *DLL) call0(name string) (uintptr, error) {
	proc, err := dll.proc(name)
	if err != nil {
		return 0, err
	}
	ret, _, _ := proc.Call()
	return ret, nil
}

func (dll *DLL) call1(name string, arg uintptr) (uintptr, error) {
	proc, err := dll.proc(name)
	if err != nil {
		return 0, err
	}
	ret, _, _ := proc.Call(arg)
	return ret, nil
}

func (dll *DLL) call2(name string, first uintptr, second uintptr) (uintptr, error) {
	proc, err := dll.proc(name)
	if err != nil {
		return 0, err
	}
	ret, _, _ := proc.Call(first, second)
	return ret, nil
}

func (dll *DLL) execCommand(command string) (int32, string, error) {
	proc, err := dll.proc("JLINKARM_ExecCommand")
	if err != nil {
		return 0, "", err
	}
	commandBytes, err := syscall.BytePtrFromString(command)
	if err != nil {
		return 0, "", err
	}
	errBuffer := make([]byte, 4096)
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(commandBytes)), uintptr(unsafe.Pointer(&errBuffer[0])), uintptr(len(errBuffer)))
	message := cString(errBuffer)
	return int32(ret), message, nil
}

func cString(buffer []byte) string {
	for index, value := range buffer {
		if value == 0 {
			return string(buffer[:index])
		}
	}
	return string(buffer)
}
