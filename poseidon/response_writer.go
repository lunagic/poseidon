package poseidon

import (
	"net/http"
	"time"
)

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
		startTime:      time.Now(),
	}
}

type ResponseWriter struct {
	http.ResponseWriter
	written    bool
	statusCode int
	startTime  time.Time
}

func (w *ResponseWriter) Written() bool {
	return w.written
}

func (w *ResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *ResponseWriter) Duration() time.Duration {
	return time.Since(w.startTime)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}

	return w.ResponseWriter.Write(b)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if !w.written {
		w.written = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
