/*
Package pkghttp defines common functionality for http.
*/
package pkghttp // import "go.pedge.io/pkg/http"

import (
	"errors"
	"fmt"
	"log/syslog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mihasya/go-metrics-librato"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/stathat"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/tylerb/graceful.v1"

	"go.pedge.io/env"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
	protosyslog "go.pedge.io/protolog/syslog"
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
	// DisableStderrLog says to disable logging to stderr.
	DisableStderrLog bool `env:"DISABLE_STDERR_LOG"`
	// The directory to write rotating logs to.
	// If not set and SyslogNetwork and SyslogAddress not set, logs will be to stderr.
	LogDir string `env:"LOG_DIR"`
	// The syslog network, either udp or tcp.
	// Must be set with SyslogAddress.
	// If not set and LogDir not set, logs will be to stderr.
	SyslogNetwork string `env:"SYSLOG_NETWORK"`
	// The syslog host:port.
	// Must be set with SyslogNetwork.
	// If not set and LogDir not set, logs will be to stderr.
	SyslogAddress string `env:"SYSLOG_ADDRESS"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC"`
	// The email address for the Librato account to send stats to.
	// Must be set with LibratoAPIToken.
	// If not set and StathatUserKey not set, no metrics.Registry for stats will be created.
	LibratoEmailAddress string `env:"LIBRATO_EMAIL_ADDRESS"`
	// The API Token for the Librato account to send stats to.
	// Must be set with LibratoEmailAddress.
	// If not set and StathatUserKey not set, no metrics.Registry for stats will be created.
	LibratoAPIToken string `env:"LIBRATO_API_TOKEN"`
	// The StatHat user key to send stats to.
	// If not set and LibratoEmailAddress and LibratoAPIToken not set, no metrics.Registry for stats will be created.
	StathatUserKey string `env:"STATHAT_USER_KEY"`
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
	protolog.RedirectStdLogger()
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
	if err := setupLogging(appName, appEnv.DisableStderrLog, appEnv.LogDir, appEnv.SyslogNetwork, appEnv.SyslogAddress); err != nil {
		return handleErrorBeforeStart(err)
	}
	registry, err := setupMetrics(appName, appEnv.StathatUserKey, appEnv.LibratoEmailAddress, appEnv.LibratoAPIToken)
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

func setupLogging(appName string, disableStderrLog bool, logDir string, syslogNetwork string, syslogAddress string) error {
	var pushers []protolog.Pusher
	if !disableStderrLog {
		pushers = append(
			pushers,
			protolog.NewStandardWritePusher(
				protolog.NewFileFlusher(
					os.Stderr,
				),
			),
		)
	}
	if logDir != "" {
		pushers = append(
			pushers,
			protolog.NewStandardWritePusher(
				protolog.NewWriterFlusher(
					&lumberjack.Logger{
						Filename:   filepath.Join(logDir, fmt.Sprintf("%s.log", appName)),
						MaxBackups: 3,
					},
				),
			),
		)
	}
	if syslogNetwork != "" && syslogAddress != "" {
		writer, err := syslog.Dial(
			syslogNetwork,
			syslogAddress,
			syslog.LOG_INFO,
			appName,
		)
		if err != nil {
			return err
		}
		pushers = append(
			pushers,
			protosyslog.NewPusher(
				writer,
				protosyslog.PusherOptions{},
			),
		)
	}
	if len(pushers) > 0 {
		protolog.SetLogger(
			protolog.NewStandardLogger(
				protolog.NewMultiPusher(
					pushers...,
				),
			),
		)
	} else {
		protolog.SetLogger(
			protolog.DiscardLogger,
		)
	}
	return nil
}

func setupMetrics(appName string, stathatUserKey string, libratoEmailAddress string, libratoAPIToken string) (metrics.Registry, error) {
	if stathatUserKey == "" && libratoEmailAddress == "" && libratoAPIToken == "" {
		return nil, nil
	}
	registry := metrics.NewPrefixedRegistry(appName)
	if stathatUserKey != "" {
		go stathat.Stathat(
			registry,
			time.Hour,
			stathatUserKey,
		)
	}
	if libratoEmailAddress != "" && libratoAPIToken != "" {
		go librato.Librato(
			registry,
			5*time.Minute,
			libratoEmailAddress,
			libratoAPIToken,
			appName,
			[]float64{0.95},
			time.Millisecond,
		)
	}
	return registry, nil
}

func handleErrorBeforeStart(err error) error {
	protolog.Error(
		&ServerCouldNotStart{
			Error: err.Error(),
		},
	)
	return err
}
