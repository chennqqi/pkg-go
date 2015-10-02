package pkghttp

import (
	"net/http"
	"time"

	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

type wrapperHandler struct {
	http.Handler
}

func newWrapperHandler(handler http.Handler) *wrapperHandler {
	return &wrapperHandler{handler}
}

func (h *wrapperHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	start := time.Now()
	wrapperResponseWriter := newWrapperResponseWriter(responseWriter)
	h.Handler.ServeHTTP(wrapperResponseWriter, request)
	protolog.Info(
		&Call{
			Path:       request.URL.Path,
			StatusCode: uint32(wrapperResponseWriter.statusCode),
			Duration:   prototime.DurationToProto(time.Since(start)),
		},
	)
}
