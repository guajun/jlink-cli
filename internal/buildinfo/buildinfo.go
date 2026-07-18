package buildinfo

import "runtime"

var (
	Version = "0.2.0-dev"
	Commit  = "unknown"
	Date    = "unknown"
)

const ProtocolVersion = "1"

type Info struct {
	Version         string `json:"version"`
	Commit          string `json:"commit"`
	Date            string `json:"date"`
	ProtocolVersion string `json:"protocol_version"`
	GoVersion       string `json:"go_version"`
	GOOS            string `json:"goos"`
	GOARCH          string `json:"goarch"`
}

func Current() Info {
	return Info{
		Version:         Version,
		Commit:          Commit,
		Date:            Date,
		ProtocolVersion: ProtocolVersion,
		GoVersion:       runtime.Version(),
		GOOS:            runtime.GOOS,
		GOARCH:          runtime.GOARCH,
	}
}
