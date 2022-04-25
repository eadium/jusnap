package snapshot

import (
	"context"
	"jusnap/internal/kernel"
	"jusnap/internal/utils"
	"time"

	"go.uber.org/zap"
)

type snapData struct {
	sig    chan struct{}
	locker *utils.Locker
	kernel *kernel.Kernel
	logger *zap.SugaredLogger
}

func snapshotLoop(ctx context.Context, d *snapData) {
	for {
		select {
		case <-ctx.Done():
			d.logger.Infof("Stopping snapshot loop")
			return
		case <-d.sig:
			if !d.locker.IsLocked() {
				d.kernel.CreateSnapshot()
			}
		}
	}
}

func ticker(ctx context.Context, tick chan struct{}) {
	ticker := time.NewTicker(300 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick <- struct{}{}
		}
	}
}

func StartLoop(ctx context.Context, k *kernel.Kernel, logger *zap.SugaredLogger) {
	d := &snapData{
		sig:    make(chan struct{}),
		locker: k.Locker,
		kernel: k,
		logger: logger,
	}
	go ticker(ctx, d.sig)
	go snapshotLoop(ctx, d)
}
