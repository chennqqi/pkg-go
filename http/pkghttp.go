/*
Package pkghttp defines common functionality for http.
*/
package pkghttp // import "go.pedge.io/pkg/http"

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rcrowley/go-metrics"

	"gopkg.in/tylerb/graceful.v1"

	"go.pedge.io/env"
	"go.pedge.io/pkg/log"
	"go.pedge.io/pkg/metrics"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

var (
	// ErrRequireAppName is the error returned if appName is not set.
	ErrRequireAppName = errors.New("pkghttp: appName must be set")
	// ErrRequireHandler is the error returned if handler is not set.
	ErrRequireHandler = errors.New("pkghttp: handler must be set")
)

// HandlerEnv is the environment for a handler.
type HandlerEnv struct {
	// The port to serve on.
	Port uint16 `env:"PORT,default=8080"`
	// HealthCheckPath is the path for health checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HEALTH_CHECK_PATH,default=/health"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC,default=10"`
}

// AppEnv is the struct that represents the environment variables used by ListenAndServe.
type AppEnv struct {
	// See pkglog for the log environment variables.
	LogEnv pkglog.Env
	// See pkgmetrics for the metrics environment variables.
	MetricsEnv pkgmetrics.Env
}

// HandlerOptions are options for a new http.Handler.
type HandlerOptions struct {
	// The registry to use.
	// Can be nil
	MetricsRegistry metrics.Registry
}

// GetHandlerEnv gets the HandlerEnv from the environment.
func GetHandlerEnv() (HandlerEnv, error) {
	handlerEnv := HandlerEnv{}
	if err := env.Populate(&handlerEnv); err != nil {
		return HandlerEnv{}, err
	}
	return handlerEnv, nil
}

// GetAppEnv gets the AppEnv from the environment.
func GetAppEnv() (AppEnv, error) {
	appEnv := AppEnv{}
	if err := env.Populate(&appEnv); err != nil {
		return AppEnv{}, err
	}
	return appEnv, nil
}

// SetupAppEnv does the setup for AppEnv.
func SetupAppEnv(appName string, appEnv AppEnv) (HandlerOptions, error) {
	if appName == "" {
		return HandlerOptions{}, ErrRequireAppName
	}
	if err := pkglog.SetupLogging(appName, appEnv.LogEnv); err != nil {
		return HandlerOptions{}, err
	}
	registry, err := pkgmetrics.SetupMetrics(appName, appEnv.MetricsEnv)
	if err != nil {
		return HandlerOptions{}, err
	}
	return HandlerOptions{
		MetricsRegistry: registry,
	}, nil
}

// NewWrapperHandler returns a new wrapper handler.
func NewWrapperHandler(delegate http.Handler, handlerEnv HandlerEnv) http.Handler {
	return newWrapperHandler(delegate, handlerEnv)
}

// ListenAndServe is the equivalent to http's method.
//
// Intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
// Calls SetupAppEnv.
//
// Uses a wrapper handler.
func ListenAndServe(appName string, handlerProvider func(HandlerOptions) (http.Handler, error)) error {
	appEnv, err := GetAppEnv()
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	handlerEnv, err := GetHandlerEnv()
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	options, err := SetupAppEnv(appName, appEnv)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	handler, err := handlerProvider(options)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	return ListenAndServeHandler(handler, handlerEnv)
}

// ListenAndServeHandler is the equivalent to http's method.
//
// Intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
//
// Uses a wrapper handler.
func ListenAndServeHandler(handler http.Handler, handlerEnv HandlerEnv) error {
	if handler == nil {
		return handleErrorBeforeStart(ErrRequireHandler)
	}
	if handlerEnv.Port == 0 {
		handlerEnv.Port = 8080
	}
	if handlerEnv.HealthCheckPath == "" {
		handlerEnv.HealthCheckPath = "/health"
	}
	if handlerEnv.ShutdownTimeoutSec == 0 {
		handlerEnv.ShutdownTimeoutSec = 10
	}
	server := &graceful.Server{
		Timeout: time.Duration(handlerEnv.ShutdownTimeoutSec) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", handlerEnv.Port),
			Handler: NewWrapperHandler(
				handler,
				handlerEnv,
			),
		},
	}
	protolog.Info(
		&ServerStarting{
			Port: uint32(handlerEnv.Port),
		},
	)
	start := time.Now()
	err := server.ListenAndServe()
	serverFinished := &ServerFinished{
		Duration: prototime.DurationToProto(time.Since(start)),
	}
	if err != nil {
		serverFinished.Error = err.Error()
		protolog.Error(serverFinished)
		return err
	}
	protolog.Info(serverFinished)
	return nil
}

func handleErrorBeforeStart(err error) error {
	protolog.Error(
		&ServerCouldNotStart{
			Error: err.Error(),
		},
	)
	return err
}
