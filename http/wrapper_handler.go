package pkghttp

import (
	"fmt"
	"net/http"
	"net/url"
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
			Method:         request.Method,
			Path:           request.URL.Path,
			RequestHeader:  headerMap(request.Header),
			RequestForm:    valuesMap(request.Form),
			ResponseHeader: headerMap(wrapperResponseWriter.Header()),
			StatusCode:     uint32(statusCode(wrapperResponseWriter.statusCode)),
			Duration:       prototime.DurationToProto(time.Since(start)),
			WriteError:     errorString(wrapperResponseWriter.writeError),
		}
		if recoverErr := recover(); recoverErr != nil {
			// TODO(pedge): should we write anything at all?
			responseWriter.WriteHeader(http.StatusInternalServerError)
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

// TODO(pedge): losing repeated fields, but seems cleaner for logging
// should we do repeated fields?
func valuesMap(values url.Values) map[string]string {
	if values == nil {
		return nil
	}
	m := make(map[string]string)
	for key := range values {
		m[key] = values.Get(key)
	}
	return m
}

func statusCode(code int) int {
	if code == 0 {
		return http.StatusOK
	}
	return code
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
