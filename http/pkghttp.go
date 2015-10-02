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

	defaultEnv = map[string]string{
		"SHUTDOWN_TIMEOUT_SEC": "10",
	}
)

type appEnv struct {
	Port                uint16 `env:"PORT,required"`
	LogDir              string `env:"LOG_DIR"`
	SyslogNetwork       string `env:"SYSLOG_NETWORK"`
	SyslogAddress       string `env:"SYSLOG_ADDRESS"`
	ShutdownTimeoutSec  uint64 `env:"SHUTDOWN_TIMEOUT_SEC"`
	LibratoEmailAddress string `env:"LIBRATO_EMAIL_ADDRESS"`
	LibratoAPIToken     string `env:"LIBRATO_API_TOKEN"`
	StathatUserKey      string `env:"STATHAT_USER_KEY"`
}

// ListenAndServe is the equivalent to http's method. Note that the metrics.Registry instance may be nil.
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
	appEnv := &appEnv{}
	if err := env.Populate(appEnv, env.PopulateOptions{Defaults: defaultEnv}); err != nil {
		return handleErrorBeforeStart(err)
	}
	if err := setupLogging(appName, appEnv.LogDir, appEnv.SyslogNetwork, appEnv.SyslogAddress); err != nil {
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
			Handler: newWrapperHandler(handler),
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

func setupLogging(appName string, logDir string, syslogNetwork string, syslogAddress string) error {
	pushers := []protolog.Pusher{
		protolog.NewStandardWritePusher(
			protolog.NewFileFlusher(
				os.Stderr,
			),
		),
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
	protolog.SetLogger(
		protolog.NewStandardLogger(
			protolog.NewMultiPusher(
				pushers...,
			),
		),
	)
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
