package kernel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Snapshot struct {
	kernel *Kernel
	Date   time.Time
	ID     string
}

func (s *Snapshot) Restore() error {
	s.kernel.Logger.Infof("Restoring checkpoint %s", s.ID)
	s.kernel.Locker.Lock()
	defer s.kernel.Locker.Unlock()
	s.kernel.Stop()

	dirPath := filepath.Join(".", "dumps", s.ID)
	pidPath := filepath.Join(".", "dumps", s.ID, "kernel.pid")

	cmd := exec.Command("criu", "restore",
		"--images-dir", dirPath,
		"--pidfile", "kernel.pid",
		"-v", "-o", "restore.log",
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
		return err
	}

	if out.Len() != 0 {
		s.kernel.Logger.Infof("CRIU: %s", out.String())
	}

	<-time.After(200 * time.Millisecond)
	s.kernel.criu = cmd.Process
	errPID := s.updatePID(pidPath)

	s.kernel.version = "criu"
	go s.kernel.criu.Wait()

	if errPID != nil {
		return errPID
	}

	s.kernel.Logger.Infof("Restored checkpoint %s with PID %d successfully", s.ID, s.kernel.proc.Pid)
	return nil
}

func (s *Snapshot) updatePID(fname string) error {
	dat, err := os.ReadFile(fname)
	if err != nil {
		s.kernel.Logger.Errorf("Error while reading PID: %s", err)
		return err
	}
	pid, err := strconv.Atoi(strings.Trim(string(dat), "\n"))
	if err != nil {
		s.kernel.Logger.Errorf("Error while convering PID: %s", err)
		return err
	}
	s.kernel.proc.Pid = pid
	err1 := os.Remove(fname)
	if err1 != nil {
		s.kernel.Logger.Errorf("Error while removing %s: %s", fname, err1)
		return err1
	}
	return nil
}
