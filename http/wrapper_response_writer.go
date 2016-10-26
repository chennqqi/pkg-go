package pkghttp

import "net/http"

type wrapResponseWriter interface {
	http.ResponseWriter
	StatusCode() int
	WriteError() error
}

type wrapperResponseWriter struct {
	http.ResponseWriter
	statusCode int
	writeError error
}

func newWrapperResponseWriter(responseWriter http.ResponseWriter) *wrapperResponseWriter {
	return &wrapperResponseWriter{responseWriter, 0, nil}
}

func (w *wrapperResponseWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.writeError = err
	return n, err
}

func (w *wrapperResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrapperResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *wrapperResponseWriter) WriteError() error {
	return w.writeError
}

type wrapperResponseWriteFlusher struct {
	http.ResponseWriter
	http.Flusher
	statusCode int
	writeError error
}

func newWrapperResponseWriteFlusher(responseWriter http.ResponseWriter, flusher http.Flusher) *wrapperResponseWriteFlusher {
	return &wrapperResponseWriteFlusher{responseWriter, flusher, 0, nil}
}

func (w *wrapperResponseWriteFlusher) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.writeError = err
	return n, err
}

func (w *wrapperResponseWriteFlusher) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrapperResponseWriteFlusher) StatusCode() int {
	return w.statusCode
}

func (w *wrapperResponseWriteFlusher) WriteError() error {
	return w.writeError
}
