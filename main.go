package main

import (
	"context"
	"jusnap/internal/kernel"
	"jusnap/internal/snapshot"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	l, _ := zap.NewProduction()
	l.WithOptions()
	defer l.Sync()
	logger = l.Sugar()
	logger.Infow("Starting application")

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	k := kernel.Create("ipykernel", logger)
	defer k.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	snapshot.StartLoop(ctx, k, logger)

	select {
	case sig := <-cancelChan:
		logger.Infof("Caught signal %v", sig)
	case <-time.After(10 * time.Second):
		logger.Infof("time passed")
	}
	cancel()
	<-time.After(5 * time.Second)

	k.Snapshots[len(k.Snapshots)-1].Restore()
	<-time.After(10 * time.Second)

	k.Snapshots[len(k.Snapshots)-1].Restore()
	<-time.After(10 * time.Second)

	logger.Infof("Stopping application")

}
