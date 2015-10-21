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
	// ErrRequireHandlerProvider is the error returned if handlerProvider is not set.
	ErrRequireHandlerProvider = errors.New("pkghttp: handlerProvider must be set")

	// DefaultEnv is the map of default environment variable values.
	DefaultEnv = map[string]string{
		"SHUTDOWN_TIMEOUT_SEC": "10",
		"HEALTH_CHECK_PATH":    "/health",
	}
)

// AppEnv is the struct that represents the environment variables used by ListenAndServe.
type AppEnv struct {
	// The port to serve on.
	// Required.
	Port uint16 `env:"PORT,required"`
	// HealthCheckPath is the path for healt checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HEALTH_CHECK_PATH"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC"`
	// See pkglog for the log environment variables.
	LogEnv pkglog.Env
	// See pkgmetrics for the metrics environment variables.
	MetricsEnv pkgmetrics.Env
}

// ListenAndServe is the equivalent to http's method.
//
// Sets up logging and metrics per the environment variables, intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
//
// Note that the metrics.Registry instance may be nil.
func ListenAndServe(appName string, handlerProvider func(metrics.Registry) (http.Handler, error)) {
	_ = listenAndServe(appName, handlerProvider)
}

func listenAndServe(appName string, handlerProvider func(metrics.Registry) (http.Handler, error)) error {
	if appName == "" {
		return handleErrorBeforeStart(ErrRequireAppName)
	}
	if handlerProvider == nil {
		return handleErrorBeforeStart(ErrRequireHandlerProvider)
	}
	appEnv := &AppEnv{}
	if err := env.Populate(appEnv, env.PopulateOptions{Defaults: DefaultEnv}); err != nil {
		return handleErrorBeforeStart(err)
	}
	if err := pkglog.SetupLogging(appName, appEnv.LogEnv); err != nil {
		return handleErrorBeforeStart(err)
	}
	registry, err := pkgmetrics.SetupMetrics(appName, appEnv.MetricsEnv)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	handler, err := handlerProvider(registry)
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	server := &graceful.Server{
		Timeout: time.Duration(appEnv.ShutdownTimeoutSec) * time.Second,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", appEnv.Port),
			Handler: newWrapperHandler(handler, appEnv.HealthCheckPath),
		},
	}
	protolog.Info(
		&ServerStarting{},
	)
	start := time.Now()
	err = server.ListenAndServe()
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
