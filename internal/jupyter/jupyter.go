package jupyter

import (
	"context"
	"jusnap/internal/config"
	"os"
	"os/exec"
	"syscall"

	"go.uber.org/zap"
)

type Notebook struct {
	Logger *zap.SugaredLogger
	proc   *os.Process
	config *config.Config
	ctx    context.Context
}

func Create(ctx context.Context, l *zap.SugaredLogger, cfg *config.Config) *Notebook {
	n := &Notebook{
		Logger: l,
		config: cfg,
		ctx:    ctx,
	}
	args := []string{"-m", "notebook"}
	args = append(args, cfg.Jusnap.JupyterConfig.JupyterArgs...)
	cmd := exec.Command(cfg.Jusnap.PythonInterpreter, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Credential: &syscall.Credential{
			Uid: uint32(cfg.Jusnap.OsConfig.Uid),
			Gid: uint32(cfg.Jusnap.OsConfig.Gid),
		},
	}

	outfile, err := os.Create("jupyter.log")
	if err != nil {
		n.Logger.Fatalf("Error while creating log file for notebook: %s", err)
	}
	defer outfile.Close()
	outfile.Chown(cfg.Jusnap.OsConfig.Uid, cfg.Jusnap.OsConfig.Gid)

	cmd.Stdout = outfile
	cmd.Stderr = outfile

	err = cmd.Start()
	if err != nil {
		n.Logger.Fatalf("Error while starting notebook: %s", err)
		return nil
	}

	n.proc = cmd.Process
	n.Logger.Infof("Notebook %d: started", n.proc.Pid)

	go func(n *Notebook) {

	}(n)

	return n
}

func (n *Notebook) Stop() {
	if err := n.proc.Signal(syscall.SIGTERM); err != nil {
		n.Logger.Errorf("Error while sending signal to notebook server %d: %s", n.proc.Pid, err)
		if err = n.proc.Kill(); err != nil {
			n.Logger.Errorf("Error while killing notebook server PID %d: %s", n.proc.Pid, err)
		}
	}

	status, err1 := n.proc.Wait()
	if err1 != nil {
		n.Logger.Errorf("Error while waiting for Notebook PID %d: %s", n.proc.Pid, err1)
	}
	n.Logger.Infof("Notebook exited with status: %s", status.String())

}
