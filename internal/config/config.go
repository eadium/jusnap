package config

import (
	"os"
	"time"

	"github.com/city-mobil/gobuns/config"
)

const (
	defaultHTTPPort              = "8000"
	defaultReadTimeout           = 5 * time.Second
	defaultWriteTimeout          = 5 * time.Second
	defaultipythonHistoryFile    = "~/.ipython/profile_default/history.sqlite"
	defaultRuntimePath           = "~/.local/share/jupyter/runtime/"
	defaultCooldownInterval      = 5 * time.Second
	defaultPythonInterpreter     = "python3"
	defaultCriuGhostLimit        = 2 // MB
	defaultIPythonHistoryEnabled = false
	defaultJupyterPort           = 8888
)

const (
	defaultLogLevel = "info"
)

var (
	httpPort          *string
	httpReadTimeout   *time.Duration
	httpWriteTimeout  *time.Duration
	logLevel          *string
	pythonInterpreter *string

	cooldownInterval      *time.Duration
	ipythonHistoryFile    *string
	ipythonHistoryEnabled *bool
	ipythonArgs           *[]string
	runtimePath           *string

	uid         *int
	gid         *int
	jupyterArgs *[]string
	jupyterPort *int

	criuGhostLimit *int
)

func init() { //nolint
	// http
	httpPort = config.String("jusnap.http.port", defaultHTTPPort, "HTTP port")
	httpReadTimeout = config.Duration("jusnap.http.read_timeout", defaultReadTimeout, "HTTP read timeout")
	httpWriteTimeout = config.Duration("jusnap.http.write_timeout", defaultWriteTimeout, "HTTP write timeout")

	// os
	uid = config.Int("jusnap.os.uid", os.Getuid(), "UID for created files")
	gid = config.Int("jusnap.os.gid", os.Getgid(), "GID for created files")

	// ipython
	pythonInterpreter = config.String("jusnap.ipython.python_interpreter", defaultPythonInterpreter, "Python interpreter to use")
	runtimePath = config.String("jusnap.ipython.runtime_path", defaultRuntimePath, "Path to Jupyter runtime dir")
	ipythonHistoryFile = config.String("jusnap.ipython.history_file", defaultipythonHistoryFile, "Path to history.sqlite")
	ipythonHistoryEnabled = config.Bool("jusnap.ipython.history_enabled", defaultIPythonHistoryEnabled, "Enables history file management (default false)")
	cooldownInterval = config.Duration("jusnap.ipython.cooldown", defaultCooldownInterval, "Snapshotting cooldown interval")
	ipythonArgs = config.StringSlice("jusnap.ipython.args", []string{}, "Launch arguments fot ipykernel")

	// jupyter notebook
	jupyterArgs = config.StringSlice("jusnap.jupyter.args", []string{}, "Launch arguments fot Jupyter Notebook")
	jupyterPort = config.Int("jusnap.jupyter.port", defaultJupyterPort, "TCP port for Jupyter Notebook")

	// app
	logLevel = config.String("jusnap.log_level", defaultLogLevel, "Logging level")

	// criu
	criuGhostLimit = config.Int("jusnap.criu.ghost_limit", defaultCriuGhostLimit, "Ghost file limit (MB)")
}

type AppConfig struct {
	HTTP              *HTTPServerConfig
	KernelConfig      *KernelConfig
	OsConfig          *OsConfig
	JupyterConfig     *JupyterConfig
	CriuConfig        *CriuConfig
	PythonInterpreter string
	LogLevel          string
	Version           string
}

type KernelConfig struct {
	HistoryFile      string
	RuntimePath      string
	IPythonArgs      []string
	CooldownInterval time.Duration
	HistoryEnabled   bool
}

type OsConfig struct {
	Uid int
	Gid int
}

type CriuConfig struct {
	GhostLimit int
}

type JupyterConfig struct {
	JupyterArgs []string
	Port        int
}

type HTTPServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Config is here for configure panopticon application
type Config struct {
	Jusnap *AppConfig
}

// Setup app configuration
func Setup(version string) *Config {
	return newDefaultConfig(version)
}

func newDefaultConfig(version string) *Config {
	return &Config{
		Jusnap: &AppConfig{
			HTTP:              newDefaultHTTPServerConfig(),
			KernelConfig:      newKernelConfig(),
			JupyterConfig:     newJupyterConfig(),
			CriuConfig:        newCriuConfig(),
			PythonInterpreter: *pythonInterpreter,
			LogLevel:          *logLevel,
			Version:           version,
			OsConfig: &OsConfig{
				Uid: *uid,
				Gid: *gid,
			},
		},
	}
}

func newKernelConfig() *KernelConfig {
	return &KernelConfig{
		HistoryFile:      *ipythonHistoryFile,
		RuntimePath:      *runtimePath,
		CooldownInterval: *cooldownInterval,
		IPythonArgs:      *ipythonArgs,
		HistoryEnabled:   *ipythonHistoryEnabled,
	}
}

func newJupyterConfig() *JupyterConfig {
	return &JupyterConfig{
		JupyterArgs: *jupyterArgs,
		Port:        *jupyterPort,
	}
}

func newCriuConfig() *CriuConfig {
	return &CriuConfig{
		GhostLimit: *criuGhostLimit,
	}
}

func newDefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port:         *httpPort,
		ReadTimeout:  *httpReadTimeout,
		WriteTimeout: *httpWriteTimeout,
	}
}
