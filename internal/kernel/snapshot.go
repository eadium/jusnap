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
	s.kernel.Logger.Info().Msgf("Restoring checkpoint %s", s.ID)
	s.kernel.Locker.Lock()
	defer s.kernel.Locker.Unlock()
	s.kernel.Stop()

	snapshotPath := filepath.Join(".", "dumps", s.ID)

	dirExists, dirErr := utils.PathExists(snapshotPath)
	if dirErr != nil {
		s.kernel.Logger.Err(dirErr).Msgf("Error while checking %s: %s", snapshotPath, dirErr.Error())
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
			s.kernel.Logger.Warn().Msgf("Error while renaming %s: %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errRename.Error())
		}
		_, errCopy := utils.Copy(historyPath, s.kernel.config.Jusnap.KernelConfig.HistoryFile)
		if errCopy != nil {
			s.kernel.Logger.Err(errCopy).Msgf("Error while restoring ipython history: %s", errCopy.Error())
			if errRename == nil {
				errRename2 := os.Rename(historyBakPath, s.kernel.config.Jusnap.KernelConfig.HistoryFile)
				if errRename2 != nil {
					s.kernel.Logger.Warn().Msgf("Error while renaming %s: %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errRename2.Error())
				}
			}
		} else {
			errRm := os.Remove(historyBakPath)
			if errRm != nil {
				s.kernel.Logger.Err(errRm).Msgf("Error while removing old ipython history (%s): %s", historyBakPath, errRm.Error())
			}
			errChmod := utils.SetFileMod(
				s.kernel.config.Jusnap.KernelConfig.HistoryFile,
				0775,
				s.kernel.config.Jusnap.OsConfig.Uid,
				s.kernel.config.Jusnap.OsConfig.Gid,
			)
			if errChmod != nil {
				s.kernel.Logger.Err(errChmod).Msgf("Error while SetFileMod ipython history (%s): %s", s.kernel.config.Jusnap.KernelConfig.HistoryFile, errChmod.Error())
			}
		}
	}

	cmd := exec.Command(s.kernel.config.Jusnap.PythonInterpreter,
		"/usr/local/sbin/criu-ns", "restore",
		"--images-dir", snapshotPath,
		// "--pidfile", "kernel.pid",
		"-v4", "-o", "restore.log",
		"--tcp-established", // https://criu.org/Advanced_usage#TCP_connections
		"--tcp-close",       // https://criu.org/CLI/opt/--tcp-close
		"--shell-job")       // https://criu.org/Advanced_usage#Shell_jobs_C.2FR
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Start()

	if err != nil {
		s.kernel.Logger.Err(err).Msgf("%s: %s", fmt.Sprint(err), stderr.String())
		return err
	}

	if out.Len() != 0 {
		s.kernel.Logger.Info().Msgf("CRIU: %s", out.String())
	}

	<-time.After(100 * time.Millisecond)
	s.kernel.criuNs = cmd.Process
	s.kernel.version = "criu"
	errPID := s.updatePID()
	if errPID != nil {
		return errPID
	}

	go s.kernel.criu.Wait()
	go s.kernel.criuNs.Wait()

	s.kernel.Logger.Info().Msgf("Restored checkpoint %s with PID %d successfully", s.ID, s.kernel.proc.Pid)
	return nil
}

func (s *Snapshot) updatePID() error {
	criupid, pid, err := utils.FindChildProcessByExecName("criu")
	if err != nil {
		s.kernel.Logger.Err(err).Msgf("Error while looking for procces: %s", err.Error())
		return err
	}
	if criupid == 0 {
		str := "CRIU process was not found"
		s.kernel.Logger.Err(err).Msgf(str)
		return errors.New(str)
	}
	if pid == 0 {
		str := "kernel process was not found"
		s.kernel.Logger.Err(err).Msgf(str)
		return errors.New(str)
	}
	s.kernel.proc.Pid = pid

	var errFindCriu error
	s.kernel.criu, errFindCriu = os.FindProcess(criupid)
	if errFindCriu != nil {
		s.kernel.Logger.Err(err).Msgf("Error while attaching criu process: %s", errFindCriu.Error())
		return errFindCriu
	}

	return nil
}
