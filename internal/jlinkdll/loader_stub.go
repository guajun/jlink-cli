//go:build !windows

package jlinkdll

import "fmt"

func ProbeSymbols(path string, names []string) ([]SymbolStatus, error) {
	statuses := make([]SymbolStatus, 0, len(names))
	for _, name := range names {
		statuses = append(statuses, SymbolStatus{Name: name, Present: false, Reason: "runtime symbol probing is not implemented on this platform yet"})
	}
	return statuses, fmt.Errorf("runtime J-Link DLL probing is not implemented on this platform yet")
}

func Load(path string) (*DLL, error) {
	return nil, fmt.Errorf("runtime J-Link DLL loading is not implemented on this platform yet")
}

type DLL struct{}

func (dll *DLL) Path() string {
	return ""
}

func (dll *DLL) Close() error {
	return nil
}
