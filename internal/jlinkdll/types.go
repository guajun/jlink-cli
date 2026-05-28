package jlinkdll

import "runtime"

type Options struct {
	ExplicitPath string
	Env          []string
}

type Candidate struct {
	Path   string `json:"path"`
	Source string `json:"source"`
	Exists bool   `json:"exists"`
	LoadOK bool   `json:"load_ok"`
	Reason string `json:"reason,omitempty"`
}

type SymbolStatus struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
	Reason  string `json:"reason,omitempty"`
}

type DoctorReport struct {
	Selected          string            `json:"selected,omitempty"`
	Candidates        []Candidate       `json:"candidates"`
	Symbols           []SymbolStatus    `json:"symbols,omitempty"`
	MissingRequired   []string          `json:"missing_required,omitempty"`
	Platform          map[string]string `json:"platform"`
	Environment       map[string]string `json:"environment"`
	RuntimeLoad       bool              `json:"runtime_load"`
	ConnectionTouched bool              `json:"connection_touched"`
}

func platform() map[string]string {
	return map[string]string{"goos": runtime.GOOS, "goarch": runtime.GOARCH}
}
