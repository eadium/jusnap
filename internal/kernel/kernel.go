package kernel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Kernel struct {
	logger *zap.SugaredLogger
	proc   *os.Process
	Name   string
	PID    int
}

func Create(name string, l *zap.SugaredLogger) *Kernel {
	k := &Kernel{
		logger: l,
		Name:   name,
	}

	cmd := exec.Command("python3", "-m", "ipykernel_launcher")
	err := cmd.Start()
	if err != nil {
		k.logger.Fatalf("Error while creating kernel: %s", err)
		return nil
	}

	k.proc = cmd.Process
	k.PID = cmd.Process.Pid
	k.logger.Infof("Created kernel %s with PID %d", name, k.PID)
	return k
}

func (k *Kernel) Kill() {

	err := k.proc.Kill()
	if err != nil {
		k.logger.Errorf("Error while killing kernel PID %d: %s", k.proc.Pid, err)
		return
	}
	k.logger.Infow("Stopped kernel")
}

func (k *Kernel) CreateSnapshot() {
	k.logger.Infow("Creating snapshot...")

	nowStr := strconv.FormatInt(time.Now().Unix(), 10)
	newpath := filepath.Join(".", "dumps", nowStr)
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		k.logger.Errorf("Error while creating snapshot directory: %s", err)
	}

	// criu dump -o dump.log --tcp-established --shell-job --leave-running -t
	cmd := exec.Command("/usr/local/sbin/criu", "dump",
		"-t", strconv.Itoa(k.proc.Pid),
		// "-o", filepath.Join(newpath, "dump.log"),
		"-D", newpath,
		"--tcp-established",
		"--shell-job",
		"--leave-running")
	// err1 := cmd.Run()
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err1 := cmd.Run()
	if err1 != nil {
		k.logger.Errorf("%s: %s", fmt.Sprint(err1), stderr.String())
		return
	}
	if out.Len() != 0 {
		k.logger.Infof("CRIU: %s", out.String())
	}

	if err1 != nil {
		k.logger.Errorf("Error while snappshotting: %v", err1)
		return
	}

	k.logger.Infow("Snapshot created")

}
