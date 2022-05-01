package main

import (
	"context"
	"jusnap/internal/api/router"
	"jusnap/internal/config"
	"jusnap/internal/jupyter"
	"jusnap/internal/kernel"
	"jusnap/internal/service/snap"
	"log"
	"net/http"
	"os"

	bunsconf "github.com/city-mobil/gobuns/config"
	"github.com/city-mobil/gobuns/graceful"
	"github.com/city-mobil/gobuns/zlog"
	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

var (
	version     = "dev"
	commit      = "none"
	serviceName = "Jusnap"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cfg, err := initConfig()
	if err != nil {
		log.Fatalln(err)
	}

	l, _ := zap.NewProduction()
	defer l.Sync()
	logger := l.Sugar()

	zlogger, accessLogger := initLogger(cfg)
	zlogger.Info().
		Str("version", version).
		Str("commit", commit).
		Msgf("Starting %s", serviceName)

	k := kernel.Create("ipykernel", logger, cfg, ctx)
	defer k.Stop()
	defer cancel()

	n := jupyter.Create(ctx, logger, cfg)
	defer n.Stop()

	snapService := snap.NewService(logger, k)
	dispatcher := router.NewDispatcher(zlogger, accessLogger, cfg, snapService)
	router := dispatcher.Init()

	httpServer := initHTTPServer(router, cfg)

	graceful.AddCallback(func() error {
		return httpServer.Shutdown(ctx)
	})

	graceful.ExecOnError(func(err error) {
		zlogger.Err(err).Msg("graceful got an error")
	})

	go func() {
		zlogger.Info().Msgf("starting HTTP server: listening on %s", cfg.Jusnap.HTTP.Port)

		err = httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			zlogger.Fatal().Err(err).Msg("failed to start the server")
		}
	}()

	err = graceful.WaitShutdown()
	zlogger.Info().Msgf("Stopping application")
	if err != nil {
		zlogger.Fatal().Err(err).Msg("failed to gracefully shutdown server")
	}
}

func initHTTPServer(router *mux.Router, cfg *config.Config) *http.Server {
	httpCfg := cfg.Jusnap.HTTP
	return &http.Server{
		Addr:         ":" + httpCfg.Port,
		Handler:      router,
		ReadTimeout:  httpCfg.ReadTimeout,
		WriteTimeout: httpCfg.WriteTimeout,
	}
}

func initConfig() (*config.Config, error) {
	err := bunsconf.InitOnce()
	if err != nil {
		return nil, err
	}
	return config.Setup(version), nil
}

func initLogger(cfg *config.Config) (logger, accessLogger zlog.Logger) {
	level, err := zlog.ParseLevel(cfg.Jusnap.LogLevel)
	if err != nil {
		level = zlog.DebugLevel
	}
	zlog.SetGlobalLevel(level)

	f, err := os.OpenFile("access.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	graceful.AddCallback(func() error {
		return f.Close()
	})
	return zlog.New(os.Stdout), zlog.New(os.Stdout)
}
