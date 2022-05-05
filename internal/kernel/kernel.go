package kernel

import (
	"bytes"
	"context"
	"fmt"
	"jusnap/internal/config"
	"jusnap/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/city-mobil/gobuns/zlog"
)

type Kernel struct {
	Locker    *utils.Locker
	Logger    zlog.Logger
	proc      *os.Process
	criu      *os.Process
	criuNs    *os.Process
	snapshots []*Snapshot
	config    *config.Config
	control   chan struct{}
	ctx       context.Context
	JsonPath  string
	Name      string
	version   string
}

func Create(name string, l zlog.Logger, cfg *config.Config, ctx context.Context) *Kernel {
	k := &Kernel{
		Logger:   l,
		Name:     name,
		Locker:   &utils.Locker{},
		config:   cfg,
		JsonPath: filepath.Join(cfg.Jusnap.KernelConfig.RuntimePath, "kernel-persist.json"),
		control:  make(chan struct{}),
		ctx:      ctx,
	}
	args := []string{"-m", "ipykernel_launcher", "-f", k.JsonPath}
	args = append(args, cfg.Jusnap.KernelConfig.IPythonArgs...)
	cmd := exec.Command(cfg.Jusnap.PythonInterpreter, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Credential: &syscall.Credential{
			Uid: uint32(cfg.Jusnap.OsConfig.Uid),
			Gid: uint32(cfg.Jusnap.OsConfig.Gid),
		},
	}

	err := cmd.Start()
	if err != nil {
		k.Logger.Fatal().Msgf("Error while creating kernel: %s", err)
		return nil
	}

	k.proc = cmd.Process
	k.proc.Pid = cmd.Process.Pid
	k.version = "vanilla"
	k.Logger.Info().Msgf("Kernel %d (%s): started", k.proc.Pid, name)

	if e := k.LoadSnapshots(); e != nil {
		k.Logger.Warn().Msgf("Error while loading existing snapshots: %s", e.Error())
	} else {
		k.Logger.Info().Msgf("Loaded %d existing snapshots from disk", len(k.snapshots))
	}

	go k.CooldownLoop()

	return k
}

func (k *Kernel) Stop() {
	var err error
	if k.version == "vanilla" {
		err = k.proc.Signal(syscall.SIGTERM)
		if err != nil {
			k.Logger.Err(err).Msgf("Error while sending signal to kernel %d: %s", k.proc.Pid, err)
			return
		}

		status, err1 := k.proc.Wait()
		k.Logger.Info().Msgf("Kernel exited with status: %s", status.String())

		if err1 != nil {
			k.Logger.Err(err).Msgf("Error while waiting for kernel PID %d: %s", k.proc.Pid, err1)
		}

	} else {
		err = syscall.Kill(-k.proc.Pid, syscall.SIGTERM)
		if err != nil {
			k.Logger.Err(err).Msgf("Error while killing kernel PID %d: %s", k.proc.Pid, err)
			return
		}

		// err = syscall.Kill(k.criu.Pid, syscall.SIGKILL)
		err = k.criu.Kill()
		if err != nil {
			k.Logger.Err(err).Msgf("Error while killing CRIU PID %d: %s", k.criu.Pid, err)
			return
		}

		status, err1 := k.criu.Wait()
		if err1 != nil {
			k.Logger.Err(err).Msgf("Error while waiting CRIU PID %d: %s", k.criu.Pid, err1)
		}
		k.Logger.Info().Msgf("CRIU exited with status: %s", status.String())
	}

	k.Logger.Info().Msgf("Kernel %d: stopped", k.proc.Pid)
}

func (k *Kernel) CreateSnapshot() (*Snapshot, error) {
	// cooldown for snapshotting
	select {
	case <-k.control:
		k.Logger.Info().Msg("Creating snapshot...")
	default:
		return nil, nil
	}

	time := time.Now()
	nowStr := strconv.FormatInt(time.Unix(), 10)
	snapshotPath := filepath.Join(".", "dumps", nowStr)
	historyPath := filepath.Join(snapshotPath, "history.sqlite")

	err := os.MkdirAll(snapshotPath, os.ModePerm)
	if err != nil {
		k.Logger.Err(err).Msgf("Error while creating snapshot directory: %s", err)
		return nil, err
	}

	cmd := exec.Command(k.config.Jusnap.PythonInterpreter,
		"/usr/local/sbin/criu-ns", "dump",
		"-t", strconv.Itoa(k.proc.Pid),
		// "-o", filepath.Join(snapshotPath, "dump.log"),
		"--images-dir", snapshotPath,
		"--tcp-established",
		"--shell-job",
		// "--file-locks",
		"--leave-running")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err1 := cmd.Run()
	if err1 != nil {
		k.Logger.Err(err).Msgf("%s: %s", fmt.Sprint(err1.Error()), stderr.String())
		oserr := os.RemoveAll(snapshotPath)
		if oserr != nil {
			k.Logger.Err(err).Msgf("Error while removing directory %s", snapshotPath)
		}
		return nil, err1
	}
	if out.Len() != 0 {
		k.Logger.Info().Msgf("CRIU: %s", out.String())
	}

	if k.config.Jusnap.KernelConfig.HistoryEnabled {
		_, errCopy := utils.Copy(k.config.Jusnap.KernelConfig.HistoryFile, historyPath)
		if errCopy != nil {
			k.Logger.Err(err).Msgf("Error while copying ipython history: %s", errCopy.Error())
		}
	}

	snap := &Snapshot{
		kernel: k,
		Date:   time,
		ID:     nowStr,
	}
	k.snapshots = append(k.snapshots, snap)
	k.Logger.Info().Msgf("Kernel %d: created snapshot %s", k.proc.Pid, nowStr)
	return snap, nil
}

func (k *Kernel) GetSnapshots() []*Snapshot {
	return k.snapshots
}

func (k *Kernel) ClearSnapshots() error {
	k.snapshots = nil
	return nil
}

func (k *Kernel) GetSnapshotsIDs() []string {
	var ids []string
	for _, v := range k.snapshots {
		ids = append(ids, v.ID)
	}
	return ids
}

func (k *Kernel) FindSnapshot(id string) *Snapshot {
	for _, s := range k.snapshots {
		if s.ID == id {
			path := filepath.Join(".", "dumps", id)
			pathExists, dirErr := utils.PathExists(path)
			if dirErr != nil {
				k.Logger.Err(dirErr).Msgf("Error while checking %s: %s", path, dirErr.Error())
				return nil
			}
			if !pathExists {
				return nil
			}
			return s
		}
	}

	return nil
}

func (k *Kernel) LoadSnapshots() error {
	f, err := os.Open("./dumps/")
	if err != nil {
		return err
	}
	snapNames, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return err
	}

	var snaps []*Snapshot
	for _, v := range snapNames {
		t, err1 := strconv.ParseInt(v, 10, 64)
		if err1 != nil {
			k.Logger.Warn().Msgf("Error while loading snapshot %s: bad name", v)
		}
		s := Snapshot{
			kernel: k,
			Date:   time.Unix(t, 0),
			ID:     v,
		}
		snaps = append(snaps, &s)
	}

	k.snapshots = snaps
	return nil
}

func (k *Kernel) CooldownLoop() {
	for {
		select {
		case <-time.After(k.config.Jusnap.KernelConfig.CooldownInterval):
			k.control <- struct{}{}
		case <-k.ctx.Done():
			return
		}
	}
}
