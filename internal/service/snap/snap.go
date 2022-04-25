package snap

import (
	"context"
	"jusnap/internal/kernel"

	"go.uber.org/zap"
)

type Service interface {
	GetSnapshots(context.Context) (*GetHistoryResponse, error)
	RestoreSnapshot(context.Context, *Request) (*RestoreSnapResponse, error)
	CreateSnapshot(context.Context) (*RestoreSnapResponse, error)
}

type service struct {
	logger *zap.SugaredLogger
	kernel *kernel.Kernel
}

func NewService(logger *zap.SugaredLogger, k *kernel.Kernel) Service {
	return &service{
		logger: logger,
		kernel: k,
	}
}

func (s *service) GetSnapshots(ctx context.Context) (*GetHistoryResponse, error) {
	var snaps []SnapshotData

	for _, s := range s.kernel.GetSnapshots() {
		snaps = append(snaps, SnapshotData{
			ID:   s.ID,
			Date: s.Date,
		})
	}

	return &GetHistoryResponse{
		Snapshots: snaps,
	}, nil
}

func (s *service) RestoreSnapshot(ctx context.Context, req *Request) (*RestoreSnapResponse, error) {
	snap := s.kernel.FindSnapshot(req.SnapID)
	if snap == nil {
		return nil, nil
	}
	err := snap.Restore()
	if err != nil {
		return nil, err
	}
	res := &RestoreSnapResponse{
		Status: "restored",
		ID:     snap.ID,
		Date:   snap.Date,
	}

	return res, nil
}

func (s *service) CreateSnapshot(ctx context.Context) (*RestoreSnapResponse, error) {

	snap, err := s.kernel.CreateSnapshot()
	if err != nil {
		return nil, err
	}

	resp := &RestoreSnapResponse{
		Status: "snapshotted",
		ID:     snap.ID,
		Date:   snap.Date,
	}

	return resp, nil
}
