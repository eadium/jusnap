package kernel

import (
	"bytes"
	"fmt"
	"jusnap/internal/utils"
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

	snapshotPath := filepath.Join(".", "dumps", s.ID)
	historyPath := filepath.Join(snapshotPath, "history.sqlite")
	pidPath := filepath.Join(".", "dumps", s.ID, "kernel.pid")
	historyBakPath := s.kernel.config.Jusnap.KernelConfig.HistoryFile + ".bak"

	errRename := os.Rename(s.kernel.config.Jusnap.KernelConfig.HistoryFile, historyBakPath)
	if errRename != nil {
		s.kernel.Logger.Warnf("Error while renaming %s: %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errRename.Error())
	}
	_, errCopy := utils.Copy(historyPath, s.kernel.config.Jusnap.KernelConfig.HistoryFile)
	if errCopy != nil {
		s.kernel.Logger.Errorf("Error while restoring ipython history: %s", errCopy.Error())
		if errRename == nil {
			errRename2 := os.Rename(historyBakPath, s.kernel.config.Jusnap.KernelConfig.HistoryFile)
			if errRename2 != nil {
				s.kernel.Logger.Warnf("Error while renaming %s: %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errRename2.Error())
			}
		}
	} else {
		errRm := os.Remove(historyBakPath)
		if errRm != nil {
			s.kernel.Logger.Errorf("Error while removing old ipython history (%s): %s", historyBakPath, errRm.Error())
		}
		errChmod := utils.SetFileMod(
			s.kernel.config.Jusnap.KernelConfig.HistoryFile,
			0775,
			s.kernel.config.Jusnap.OsConfig.Uid,
			s.kernel.config.Jusnap.OsConfig.Gid,
		)
		if errChmod != nil {
			s.kernel.Logger.Errorf("Error while SetFileMod ipython history (%s): %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errChmod.Error())
		}
	}

	cmd := exec.Command("criu", "restore",
		"--images-dir", snapshotPath,
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
