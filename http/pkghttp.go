package pkghttp // import "go.pedge.io/pkg/http"
import (
	"fmt"
	"log/syslog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/tylerb/graceful.v1"

	"go.pedge.io/env"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
	protosyslog "go.pedge.io/protolog/syslog"
)

var (
	// DefaultEnv is the default environment variable values.
	DefaultEnv = map[string]string{
		"SHUTDOWN_TIMEOUT_SEC": "10",
	}
)

// AppEnv has the environment variables that must be set.
type AppEnv struct {
	Port               uint16 `env:"PORT,required"`
	LogDir             string `env:"LOG_DIR"`
	SyslogNetwork      string `env:"SYSLOG_NETWORK"`
	SyslogAddress      string `env:"SYSLOG_ADDRESS"`
	ShutdownTimeoutSec uint64 `env:"SHUTDOWN_TIMEOUT_SEC"`
}

// Handler handles HTTP calls.
type Handler interface {
	// ServeHTTP is equivalent to http's method, but has a return value of the status code.
	ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) int
}

// ListenAndServe is the equivalent to http's method.
func ListenAndServe(appName string, f func() (Handler, error)) {
	_ = listenAndServe(appName, f)
}

func listenAndServe(appName string, f func() (Handler, error)) error {
	appEnv := &AppEnv{}
	if err := env.Populate(appEnv, env.PopulateOptions{Defaults: DefaultEnv}); err != nil {
		return handleErrorBeforeStart(err)
	}
	if err := setupLogging(appName, appEnv.LogDir, appEnv.SyslogNetwork, appEnv.SyslogAddress); err != nil {
		return handleErrorBeforeStart(err)
	}
	handler, err := f()
	if err != nil {
		return handleErrorBeforeStart(err)
	}
	server := &graceful.Server{
		Timeout: time.Duration(appEnv.ShutdownTimeoutSec) * time.Second,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", appEnv.Port),
			Handler: newInternalHandler(handler),
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

func handleErrorBeforeStart(err error) error {
	protolog.Error(
		&ServerCouldNotStart{
			Error: err.Error(),
		},
	)
	return err
}

type internalHandler struct {
	handler Handler
}

func newInternalHandler(handler Handler) *internalHandler {
	return &internalHandler{handler}
}

func (h *internalHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	start := time.Now()
	statusCode := h.handler.ServeHTTP(responseWriter, request)
	protolog.Info(
		&Call{
			Path:       request.URL.Path,
			StatusCode: uint32(statusCode),
			Duration:   prototime.DurationToProto(time.Since(start)),
		},
	)
}
