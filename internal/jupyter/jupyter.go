package jupyter

import (
	"bytes"
	"context"
	"fmt"
	"jusnap/internal/config"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/city-mobil/gobuns/zlog"
)

type Notebook struct {
	Logger zlog.Logger
	proc   *os.Process
	config *config.Config
	ctx    context.Context
}

func Create(ctx context.Context, l zlog.Logger, cfg *config.Config) *Notebook {
	n := &Notebook{
		Logger: l,
		config: cfg,
		ctx:    ctx,
	}
	args := []string{"-m", "notebook", "--port", strconv.Itoa(cfg.Jusnap.JupyterConfig.Port)}
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
		n.Logger.Error().Msgf("Error while creating log file for notebook: %s", err)
	}
	defer outfile.Close()
	outfile.Chown(cfg.Jusnap.OsConfig.Uid, cfg.Jusnap.OsConfig.Gid)

	cmd.Stdout = outfile
	cmd.Stderr = outfile

	err = cmd.Start()
	if err != nil {
		n.Logger.Fatal().Msgf("Error while starting notebook: %s", err)
		return nil
	}

	n.proc = cmd.Process
	n.Logger.Info().Msgf("Notebook %d: started", n.proc.Pid)

	go func(n *Notebook) {
		<-time.After(5 * time.Second)
		cmd := exec.Command("jupyter", "notebook", "list")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err1 := cmd.Run()
		if err1 != nil {
			n.Logger.Err(err).Msgf("%s: %s", fmt.Sprint(err1.Error()), stderr.String())
			return
		}
		if out.Len() != 0 {
			urlRegex, errRegex := regexp.Compile(`http://((((\d+)\.*){4})|((\w+)\.*)+):(?P<port>\d+)/\?token=\w{48}`)
			if errRegex != nil {
				n.Logger.Err(err).Msgf("Error while compiling Jupyter URL regexp")
			}
			urlSlice := urlRegex.FindStringSubmatch(out.String())
			if len(urlSlice) == 0 {
				return
			}
			urlMap := reSubMatchMap(urlRegex, out.String())
			if len(urlMap) != 0 {
				url := strings.Replace(urlSlice[0], urlMap["port"], cfg.Jusnap.HTTP.Port, 1)
				n.Logger.Info().Msgf("Jupyter Notebook URL: %s", url)
			}
		}
	}(n)

	return n
}

func (n *Notebook) Stop() {
	if err := n.proc.Signal(syscall.SIGTERM); err != nil {
		n.Logger.Err(err).Msgf("Error while sending signal to notebook server %d: %s", n.proc.Pid, err)
		if err = n.proc.Kill(); err != nil {
			n.Logger.Err(err).Msgf("Error while killing notebook server PID %d: %s", n.proc.Pid, err)
		}
	}

	status, err1 := n.proc.Wait()
	if err1 != nil {
		n.Logger.Err(err1).Msgf("Error while waiting for Notebook PID %d: %s", n.proc.Pid, err1)
	}
	n.Logger.Info().Msgf("Notebook exited with status: %s", status.String())

}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
