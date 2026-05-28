//go:build windows

package jlinkdll

import (
	"fmt"
	"syscall"
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
