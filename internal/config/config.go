package config

import (
	"time"

	"github.com/city-mobil/gobuns/config"
)

const (
	defaultHTTPPort     = "8000"
	defaultReadTimeout  = time.Second
	defaultWriteTimeout = time.Second
)

const (
	defaultLogLevel = "info"
)

var (
	httpPort         *string
	httpReadTimeout  *time.Duration
	httpWriteTimeout *time.Duration
	logLevel         *string
)

func init() { //nolint
	httpPort = config.String("jusnap.http.port", defaultHTTPPort, "HTTP port")
	httpReadTimeout = config.Duration("jusnap.http.read_timeout", defaultReadTimeout, "HTTP read timeout")
	httpWriteTimeout = config.Duration("jusnap.http.write_timeout", defaultWriteTimeout, "HTTP write timeout")

	logLevel = config.String("jusnap.log_level", defaultLogLevel, "Logging level")
}

type AppConfig struct {
	HTTP     *HTTPServerConfig
	LogLevel string
	Version  string
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
			HTTP:     newDefaultHTTPServerConfig(),
			LogLevel: *logLevel,
			Version:  version,
		},
	}
}

func newDefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port:         *httpPort,
		ReadTimeout:  *httpReadTimeout,
		WriteTimeout: *httpWriteTimeout,
	}
}
