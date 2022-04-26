package main

import (
	"context"
	"jusnap/internal/api/router"
	"jusnap/internal/config"
	"jusnap/internal/kernel"
	"jusnap/internal/service/snap"
	"jusnap/internal/snapshot"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	bunsconf "github.com/city-mobil/gobuns/config"
	"github.com/city-mobil/gobuns/graceful"
	"github.com/city-mobil/gobuns/zlog"
	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

// var logger *zap.SugaredLogger

var (
	version     = "dev"
	commit      = "none"
	serviceName = "Jusnap"
)

func main() {
	ctx := context.Background()

	cfg, err := initConfig()
	if err != nil {
		log.Fatalln(err)
	}

	l, _ := zap.NewProduction()
	l.WithOptions()
	defer l.Sync()
	logger := l.Sugar()

	logger.Infow("Starting application")

	zlogger, accessLogger := initLogger(cfg)
	zlogger.Info().
		Str("version", version).
		Str("commit", commit).
		Msgf("Starting %s", serviceName)

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	k := kernel.Create("ipykernel", logger, cfg)
	defer k.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	snapshot.StartLoop(ctx, k, logger)

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
