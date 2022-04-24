package kernel

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type Snapshot struct {
	kernel *Kernel
	date   time.Time
	ID     string
	seqNum uint
}

func (s *Snapshot) Restore() {
	s.kernel.Logger.Infof("Restoring checkpoint %s", s.ID)
	s.kernel.Locker.Lock()
	defer s.kernel.Locker.Unlock()
	s.kernel.Stop()

	<-time.After(10 * time.Second)

	cmd := exec.Command("criu", "restore",
		"--images-dir", filepath.Join(".", "dumps", s.ID),
		"--tcp-established",
		"--shell-job")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		s.kernel.Logger.Errorf("%s: %s", fmt.Sprint(err), stderr.String())
		return
	}
	if out.Len() != 0 {
		s.kernel.Logger.Infof("CRIU: %s", out.String())
	}
	if err != nil {
		s.kernel.Logger.Infof("Error while restoring kernel ID %s", s.ID)
		return
	}
	s.kernel.criu = cmd.Process
	s.kernel.version = "criu"
	s.kernel.Logger.Infof("Restored checkpoint %s successfully", s.ID)
}
