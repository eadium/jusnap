package v1

import (
	"encoding/json"
	"fmt"
	"jusnap/internal/service/snap"
	"net/http"

	"github.com/city-mobil/gobuns/handlers"
	"github.com/city-mobil/gobuns/zlog"
)

type Handler interface {
	GetSnapshots(handlers.Context) (interface{}, error)
	RestoreSnapshot(handlers.Context) (interface{}, error)
	CreateSnapshot(ctx handlers.Context) (interface{}, error)
	ClearSnapshots(ctx handlers.Context) (interface{}, error)
}

type snapHandler struct {
	logger  zlog.Logger
	Service snap.Service
}

func (s *snapHandler) RestoreSnapshot(ctx handlers.Context) (interface{}, error) {
	req, err := parseRequest(ctx.HTTPRequest())
	if err != nil {
		s.logger.Warn().Msgf((fmt.Sprintf("[RestoreSnapshot] bad request: %s", err)))
		return nil, &handlers.RequestError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	res, err := s.Service.RestoreSnapshot(ctx, req)
	if err != nil {
		s.logger.Error().Msgf((fmt.Sprintf("[RestoreSnapshot] internal error: %s", err)))
		return nil, &handlers.RequestError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	if res == nil {
		return nil, &handlers.RequestError{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Snapshot %s not found", req.SnapID),
		}
	}
	return res, nil
}

func (s *snapHandler) CreateSnapshot(ctx handlers.Context) (interface{}, error) {

	res, err := s.Service.CreateSnapshot(ctx)
	if err != nil {
		s.logger.Error().Msgf((fmt.Sprintf("[CreateSnapshot] internal error: %s", err)))
		return nil, &handlers.RequestError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return res, nil
}

func (s *snapHandler) GetSnapshots(ctx handlers.Context) (interface{}, error) {

	res, err := s.Service.GetSnapshots(ctx)
	if err != nil {
		return nil, &handlers.RequestError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return res, nil
}

func (s *snapHandler) ClearSnapshots(ctx handlers.Context) (interface{}, error) {

	res, err := s.Service.ClearSnapshots()
	if err != nil {
		return nil, &handlers.RequestError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return res, nil
}

func parseRequest(req *http.Request) (*snap.Request, error) {
	var o snap.Request
	err := json.NewDecoder(req.Body).Decode(&o)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func NewSnapHandler(logger zlog.Logger, srv snap.Service) Handler {
	return &snapHandler{
		logger:  logger,
		Service: srv,
	}
}
