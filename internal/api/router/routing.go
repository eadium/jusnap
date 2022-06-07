package router

import (
	"fmt"
	v1 "jusnap/internal/api/v1"
	"jusnap/internal/config"
	"jusnap/internal/service/snap"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/city-mobil/gobuns/handlers"
	"github.com/city-mobil/gobuns/zlog"
	"github.com/gorilla/mux"
)

const (
	messageTimeout = "HTTP: request timeout"
)

var proxyUrl *url.URL

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
	d.addProxyRoute(router)
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
		), d.cfg.Jusnap.HTTP.ReadTimeout,
		messageTimeout,
	)).Methods(http.MethodGet, http.MethodOptions)

	router.Handle("/snap/restore", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.RestoreSnapshot,
		), d.cfg.Jusnap.HTTP.ReadTimeout,
		messageTimeout,
	)).Methods(http.MethodPost)

	router.Handle("/snap/new", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.CreateSnapshot,
		), d.cfg.Jusnap.HTTP.ReadTimeout,
		messageTimeout,
	)).Methods(http.MethodPost)

	router.Handle("/snap/clear", http.TimeoutHandler(
		handlers.AccessLogHandler(
			accessLogger,
			snapHandler.ClearSnapshots,
		), d.cfg.Jusnap.HTTP.ReadTimeout,
		messageTimeout,
	)).Methods(http.MethodDelete, http.MethodOptions)
}

func (d *Dispatcher) addProxyRoute(router *mux.Router) {
	var err error
	proxyUrl, err = url.Parse(fmt.Sprintf("http://127.0.0.1:%d", d.cfg.Jusnap.JupyterConfig.Port))
	if err != nil {
		return
	}
	router.PathPrefix("/").HandlerFunc(proxyPass)
}

func proxyPass(res http.ResponseWriter, req *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	proxy.ServeHTTP(res, req)
}
