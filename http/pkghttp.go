/*
Package pkghttp defines common functionality for http.
*/
package pkghttp // import "go.pedge.io/pkg/http"

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"go.pedge.io/env"
	"go.pedge.io/pkg/app"
	"go.pedge.io/pkg/http"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

var (
	// ErrRequireHandler is the error returned if handler is not set.
	ErrRequireHandler = errors.New("pkghttp: handler must be set")
)

// HandlerEnv is the environment for a handler.
type HandlerEnv struct {
	// The port to serve on.
	Port uint16 `env:"HTTP_PORT,default=8080"`
	// HealthCheckPath is the path for health checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HTTP_EALTH_CHECK_PATH,default=/health"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"HTTP_SHUTDOWN_TIMEOUT_SEC,default=10"`
}

// GetHandlerEnv gets the HandlerEnv from the environment.
func GetHandlerEnv() (HandlerEnv, error) {
	handlerEnv := HandlerEnv{}
	if err := env.Populate(&handlerEnv); err != nil {
		return HandlerEnv{}, err
	}
	return handlerEnv, nil
}

// NewWrapperHandler returns a new wrapper handler.
func NewWrapperHandler(delegate http.Handler, handlerEnv HandlerEnv) http.Handler {
	return newWrapperHandler(delegate, handlerEnv)
}

// ListenAndServeSimple is ListenAndServe with environment variable setup.
func ListenAndServeSimple(handler http.Handler) error {
	appEnv, err := pkgapp.GetAppEnv()
	if err != nil {
		return err
	}
	if _, err := pkgapp.SetupAppEnv(appEnv); err != nil {
		return err
	}
	handlerEnv, err := pkghttp.GetHandlerEnv()
	if err != nil {
		return err
	}
	return pkghttp.ListenAndServe(handler, handlerEnv)
}

// ListenAndServe is the equivalent to http's method.
//
// Intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
//
// Uses a wrapper handler.
func ListenAndServe(handler http.Handler, handlerEnv HandlerEnv) error {
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
