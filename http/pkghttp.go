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

// AppEnv is the struct that represents the environment variables used by ListenAndServe.
type AppEnv struct {
	// The port to serve on.
	Port uint16 `env:"PORT,default=8080"`
	// HealthCheckPath is the path for health checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HEALTH_CHECK_PATH,default=/health"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC,default=10"`
	// See pkglog for the log environment variables.
	LogEnv pkglog.Env
	// See pkgmetrics for the metrics environment variables.
	MetricsEnv pkgmetrics.Env
}

// HandlerOptions are options for a new http.Handler.
type HandlerOptions struct {
	// The port to serve on.
	// Default value is 8080.
	Port uint16 `env:"PORT,default=8080"`
	// HealthCheckPath is the path for health checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HEALTH_CHECK_PATH,default=/health"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC,default=10"`
	// The registry to use.
	// Can be nil
	MetricsRegistry metrics.Registry
}

// SetupAppEnv reads the environment variables for AppEnv and does the setup.
func SetupAppEnv(appName string) (HandlerOptions, error) {
	if appName == "" {
		return HandlerOptions{}, ErrRequireAppName
	}
	appEnv := &AppEnv{}
	if err := env.Populate(appEnv); err != nil {
		return HandlerOptions{}, err
	}
	if err := pkglog.SetupLogging(appName, appEnv.LogEnv); err != nil {
		return HandlerOptions{}, err
	}
	registry, err := pkgmetrics.SetupMetrics(appName, appEnv.MetricsEnv)
	if err != nil {
		return HandlerOptions{}, err
	}
	return HandlerOptions{
		Port:               appEnv.Port,
		HealthCheckPath:    appEnv.HealthCheckPath,
		ShutdownTimeoutSec: appEnv.ShutdownTimeoutSec,
		MetricsRegistry:    registry,
	}, nil
}

// NewWrapperHandler returns a new wrapper handler.
func NewWrapperHandler(delegate http.Handler, options HandlerOptions) http.Handler {
	return newWrapperHandler(delegate, options)
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
	options, err := SetupAppEnv(appName)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	handler, err := handlerProvider(options)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	return ListenAndServeHandler(handler, options)
}

// ListenAndServeHandler is the equivalent to http's method.
//
// Intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
//
// Uses a wrapper handler.
func ListenAndServeHandler(handler http.Handler, options HandlerOptions) error {
	if handler == nil {
		return handleErrorBeforeStart(ErrRequireHandler)
	}
	if options.Port == 0 {
		options.Port = 8080
	}
	if options.HealthCheckPath == "" {
		options.HealthCheckPath = "/health"
	}
	if options.ShutdownTimeoutSec == 0 {
		options.ShutdownTimeoutSec = 10
	}
	server := &graceful.Server{
		Timeout: time.Duration(options.ShutdownTimeoutSec) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", options.Port),
			Handler: NewWrapperHandler(
				handler,
				options,
			),
		},
	}
	protolog.Info(
		&ServerStarting{
			Port: uint32(options.Port),
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
