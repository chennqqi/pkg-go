package pkghttp

import "net/http"

type wrapperResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newWrapperResponseWriter(responseWriter http.ResponseWriter) *wrapperResponseWriter {
	return &wrapperResponseWriter{responseWriter, 0}
}

func (w *wrapperResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
