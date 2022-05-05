package snap

import (
	"context"
	"jusnap/internal/kernel"
	"os"

	"github.com/city-mobil/gobuns/zlog"
)

type Service interface {
	GetSnapshots(context.Context) (*GetHistoryResponse, error)
	RestoreSnapshot(context.Context, *Request) (*RestoreSnapResponse, error)
	CreateSnapshot(context.Context) (*RestoreSnapResponse, error)
	ClearSnapshots() (*ClearSnapResponse, error)
}

type service struct {
	logger zlog.Logger
	kernel *kernel.Kernel
}

func NewService(logger zlog.Logger, k *kernel.Kernel) Service {
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

	if len(snaps) == 0 {
		snaps = make([]SnapshotData, 0)
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
	if snap == nil {
		resp := &RestoreSnapResponse{
			Status: "skipped",
		}
		return resp, nil
	}

	resp := &RestoreSnapResponse{
		Status: "snapshotted",
		ID:     snap.ID,
		Date:   snap.Date,
	}

	return resp, nil
}

func (s *service) ClearSnapshots() (*ClearSnapResponse, error) {
	errClear := s.kernel.ClearSnapshots()
	if errClear != nil {
		s.logger.Err(errClear).Msgf("Error while cleaning snapshots in memory: %s", errClear.Error())
	}
	err := os.RemoveAll("./dumps/")
	if err != nil {
		s.logger.Err(err).Msgf("Error while removing snapshots directory: %s", err.Error())
		return nil, err
	}
	err = os.MkdirAll("./dumps/", 0755)
	if err != nil {
		s.logger.Err(err).Msgf("Error while creating snapshots directory: %s", err.Error())
		return nil, err
	}
	return &ClearSnapResponse{
		Status: "cleared",
	}, nil
}
