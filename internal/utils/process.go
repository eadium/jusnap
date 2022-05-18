package utils

import (
	"strings"

	ps "github.com/mitchellh/go-ps"
)

func FindChildProcessByExecName(name string) (ppid, pid int, err error) {
	processList, err := ps.Processes()
	if err != nil {
		return 0, 0, err
	}

	// O(n^2) - subject to refactor
	for _, p := range processList {
		if strings.Contains(p.Executable(), name) {
			ppid := p.Pid()
			for _, c := range processList {
				if c.PPid() == ppid {
					return ppid, c.Pid(), nil
				}
			}
			return ppid, 0, nil
		}
	}

	return 0, 0, nil
}
