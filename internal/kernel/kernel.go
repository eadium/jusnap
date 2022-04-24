package kernel

import (
	"bytes"
	"fmt"
	"jusnap/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Kernel struct {
	Locker    *utils.Locker
	Logger    *zap.SugaredLogger
	proc      *os.Process
	criu      *os.Process
	Snapshots []*Snapshot
	Name      string
	version   string
}

func Create(name string, l *zap.SugaredLogger) *Kernel {
	k := &Kernel{
		Logger: l,
		Name:   name,
		Locker: &utils.Locker{},
	}

	cmd := exec.Command("python3", "-m", "ipykernel_launcher")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()
	if err != nil {
		k.Logger.Fatalf("Error while creating kernel: %s", err)
		return nil
	}

	k.proc = cmd.Process
	k.proc.Pid = cmd.Process.Pid
	k.version = "vanilla"
	k.Logger.Infof("Kernel %d (%s): started", k.proc.Pid, name)
	return k
}

func (k *Kernel) Stop() {
	var err error
	if k.version == "vanilla" {
		err = k.proc.Signal(syscall.SIGTERM)
		if err != nil {
			k.Logger.Errorf("Error while sending signal to kernel %d: %s", k.proc.Pid, err)
			return
		}

		status, err1 := k.proc.Wait()
		k.Logger.Infof("Kernel exited with status: %s", status.String())

		if err1 != nil {
			k.Logger.Errorf("Error while waiting kernel PID %d: %s", k.proc.Pid, err1)
		}

	} else {
		err = syscall.Kill(-k.proc.Pid, syscall.SIGKILL)
		if err != nil {
			k.Logger.Errorf("Error while killing kernel PID %d: %s", k.proc.Pid, err)
			return
		}

		err = syscall.Kill(k.criu.Pid, syscall.SIGKILL)
		if err != nil {
			k.Logger.Errorf("Error while killing CRIU PID %d: %s", k.proc.Pid, err)
			return
		}

		status, err1 := k.criu.Wait()
		k.Logger.Infof("Kernel exited with status: %s", status.String())

		if err1 != nil {
			k.Logger.Errorf("Error while waiting CRIU PID %d: %s", k.proc.Pid, err1)
		}
	}

	k.Logger.Infof("Kernel %d: stopped", k.proc.Pid)
}

func (k *Kernel) CreateSnapshot() {
	k.Logger.Infow("Creating snapshot...")

	time := time.Now()
	nowStr := strconv.FormatInt(time.Unix(), 10)
	newpath := filepath.Join(".", "dumps", nowStr)
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		k.Logger.Errorf("Error while creating snapshot directory: %s", err)
	}

	cmd := exec.Command("criu", "dump",
		"-t", strconv.Itoa(k.proc.Pid),
		// "-o", filepath.Join(newpath, "dump.log"),
		"--images-dir", newpath,
		"--tcp-established",
		"--shell-job",
		"--leave-running")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err1 := cmd.Run()
	if err1 != nil {
		k.Logger.Errorf("%s: %s", fmt.Sprint(err1), stderr.String())
		return
	}
	if out.Len() != 0 {
		k.Logger.Infof("CRIU: %s", out.String())
	}

	snap := &Snapshot{
		kernel: k,
		date:   time,
		ID:     nowStr,
	}
	k.Snapshots = append(k.Snapshots, snap)
	k.Logger.Infof("Kernel %d: created snapshot %s", k.proc.Pid, nowStr)

}
