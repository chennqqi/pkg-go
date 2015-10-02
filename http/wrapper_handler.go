package pkghttp

import (
	"fmt"
	"net/http"
	"runtime"
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
	defer func() {
		call := &Call{
			Path:           request.URL.Path,
			RequestHeader:  headerMap(request.Header),
			ResponseHeader: headerMap(wrapperResponseWriter.Header()),
			StatusCode:     uint32(wrapperResponseWriter.statusCode),
			Duration:       prototime.DurationToProto(time.Since(start)),
			WriteError:     errorString(wrapperResponseWriter.writeError),
		}
		if recoverErr := recover(); recoverErr != nil {
			stack := make([]byte, 8192)
			stack = stack[:runtime.Stack(stack, false)]
			call.PanicError = fmt.Sprintf("%v", recoverErr)
			call.PanicStack = string(stack)
		}
		protolog.Info(call)
	}()
	h.Handler.ServeHTTP(wrapperResponseWriter, request)
}

// TODO(pedge): losing repeated fields, but seems cleaner for logging
// should we do repeated fields?
func headerMap(header http.Header) map[string]string {
	if header == nil {
		return nil
	}
	m := make(map[string]string)
	for key := range header {
		m[key] = header.Get(key)
	}
	return m
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
