package router

import (
	v1 "jusnap/internal/api/v1"
	"jusnap/internal/config"
	"jusnap/internal/service/snap"
	"net/http"
	"time"

	"github.com/city-mobil/gobuns/handlers"
	"github.com/city-mobil/gobuns/zlog"
	"github.com/gorilla/mux"
)

const (
	messageTimeout = "HTTP: request timeout"
)

type Dispatcher struct {
	cfg          *config.Config
	logger       zlog.Logger
	accessLogger zlog.Logger
	srv          snap.Service
}

func NewDispatcher(
	logger zlog.Logger,
	accessLogger zlog.Logger,
	cfg *config.Config,
	logSrv snap.Service,
) *Dispatcher {
	return &Dispatcher{
		logger:       logger,
		accessLogger: accessLogger,
		cfg:          cfg,
		srv:          logSrv,
	}
}

func (d *Dispatcher) Init() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	internal := d.addPrefixRouting(router)
	d.addRoutesV1(internal)

	return router
}

func (d *Dispatcher) addPrefixRouting(r *mux.Router) *mux.Router {
	router := r.PathPrefix("/api/").Subrouter()
	router.Use(jsonContentTypeHandler)

	return router
}

func (d *Dispatcher) addRoutesV1(router *mux.Router) {

	snapHandler := v1.NewSnapHandler(d.logger, d.srv)
	accessLogger := handlers.NewAccessLogger("/api/", handlers.WithLogger(d.accessLogger))

	router.Handle("/snap", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.GetSnapshots,
		), time.Second,
		messageTimeout,
	)).Methods(http.MethodGet)

	router.Handle("/snap/restore", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.RestoreSnapshot,
		), time.Second,
		messageTimeout,
	)).Methods(http.MethodPost)

	router.Handle("/snap/new", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.CreateSnapshot,
		), time.Second,
		messageTimeout,
	)).Methods(http.MethodPost)
}
