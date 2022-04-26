package config

import (
	"os"
	"time"

	"github.com/city-mobil/gobuns/config"
)

const (
	defaultHTTPPort           = "8000"
	defaultReadTimeout        = time.Second
	defaultWriteTimeout       = time.Second
	defaultipythonHistoryFile = "~/.ipython/profile_default/history.sqlite"
)

const (
	defaultLogLevel = "info"
)

var (
	httpPort         *string
	httpReadTimeout  *time.Duration
	httpWriteTimeout *time.Duration
	logLevel         *string

	ipythonHistoryFile *string
	uid                *int
	gid                *int
)

func init() { //nolint
	httpPort = config.String("jusnap.http.port", defaultHTTPPort, "HTTP port")
	httpReadTimeout = config.Duration("jusnap.http.read_timeout", defaultReadTimeout, "HTTP read timeout")
	httpWriteTimeout = config.Duration("jusnap.http.write_timeout", defaultWriteTimeout, "HTTP write timeout")

	ipythonHistoryFile = config.String("jusnap.ipython.history_file", defaultipythonHistoryFile, "Path to history.sqlite")
	uid = config.Int("jusnap.os.uid", os.Getuid(), "UID for created files")
	gid = config.Int("jusnap.os.gid", os.Getgid(), "GID for created files")

	logLevel = config.String("jusnap.log_level", defaultLogLevel, "Logging level")
}

type AppConfig struct {
	HTTP         *HTTPServerConfig
	KernelConfig *KernelConfig
	OsConfig     *OsConfig
	LogLevel     string
	Version      string
}

type KernelConfig struct {
	HistoryFile string
}

type OsConfig struct {
	Uid int
	Gid int
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
			HTTP:         newDefaultHTTPServerConfig(),
			LogLevel:     *logLevel,
			Version:      version,
			KernelConfig: newKernelConfig(),
			OsConfig: &OsConfig{
				Uid: *uid,
				Gid: *gid,
			},
		},
	}
}

func newKernelConfig() *KernelConfig {
	return &KernelConfig{
		HistoryFile: *ipythonHistoryFile,
	}
}

func newDefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port:         *httpPort,
		ReadTimeout:  *httpReadTimeout,
		WriteTimeout: *httpWriteTimeout,
	}
}
