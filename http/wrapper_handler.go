package pkghttp

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
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
			RequestHeader:  valuesMap(request.Header),
			RequestForm:    valuesMap(request.Form),
			ResponseHeader: valuesMap(wrapperResponseWriter.Header()),
			StatusCode:     uint32(statusCode(wrapperResponseWriter.statusCode)),
			Error:          errorString(wrapperResponseWriter.writeError),
		}
		if request.URL != nil {
			call.Path = request.URL.Path
			call.Query = valuesMap(request.URL.Query())
		}
		if recoverErr := recover(); recoverErr != nil {
			// TODO(pedge): should we write anything at all?
			responseWriter.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, 8192)
			stack = stack[:runtime.Stack(stack, false)]
			call.Error = fmt.Sprintf("panic: %v\n%s", recoverErr, string(stack))
		}
		call.Duration = prototime.DurationToProto(time.Since(start))
		protolog.Info(call)
	}()
	h.Handler.ServeHTTP(wrapperResponseWriter, request)
}

func valuesMap(values map[string][]string) map[string]string {
	if values == nil {
		return nil
	}
	m := make(map[string]string)
	for key, value := range values {
		m[key] = strings.Join(value, " ")
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
