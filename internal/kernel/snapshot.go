package kernel

import (
	"bytes"
	"errors"
	"fmt"
	"jusnap/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
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

	dirExists, dirErr := utils.PathExists(snapshotPath)
	if dirErr != nil {
		s.kernel.Logger.Errorf("Error while checking %s: %s", snapshotPath, dirErr.Error())
		return dirErr
	}
	if !dirExists {
		return errors.New("snapshot directory not found")
	}

	if s.kernel.config.Jusnap.KernelConfig.HistoryEnabled {
		historyBakPath := s.kernel.config.Jusnap.KernelConfig.HistoryFile + ".bak"
		historyPath := filepath.Join(snapshotPath, "history.sqlite")
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
	}

	cmd := exec.Command(s.kernel.config.Jusnap.PythonInterpreter,
		"/usr/local/sbin/criu-ns", "restore",
		"--images-dir", snapshotPath,
		// "--pidfile", "kernel.pid",
		"-v4", "-o", "restore.log",
		"--tcp-established",
		"--tcp-close",
		// "--file-locks",
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

	<-time.After(100 * time.Millisecond)
	s.kernel.criuNs = cmd.Process
	s.kernel.version = "criu"
	errPID := s.updatePID()

	go s.kernel.criu.Wait()
	go s.kernel.criuNs.Wait()

	if errPID != nil {
		return errPID
	}

	s.kernel.Logger.Infof("Restored checkpoint %s with PID %d successfully", s.ID, s.kernel.proc.Pid)
	return nil
}

func (s *Snapshot) updatePID() error {
	criupid, pid, err := utils.FindChildProcessByExecName("criu")
	if err != nil {
		s.kernel.Logger.Errorf("Error while looking for procces: %s", err.Error())
		return err
	}
	if pid == 0 {
		str := "CRIU process was not found"
		s.kernel.Logger.Errorf(str)
		return errors.New(str)
	}
	s.kernel.proc.Pid = pid

	var errFindCriu error
	s.kernel.criu, errFindCriu = os.FindProcess(criupid)
	if errFindCriu != nil {
		s.kernel.Logger.Errorf("Error while attaching criu process: %s", errFindCriu.Error())
		return errFindCriu
	}

	fmt.Printf("proc pid: %d, criu pid: %d\n", pid, criupid)
	return nil
}
