package jlinkdll

import "os"

func environ() []string {
	return os.Environ()
}
