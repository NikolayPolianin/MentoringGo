package metrics

import (
	"net/http"
	"time"
)

type Writer struct {
	http.ResponseWriter
	wroteCode bool
	code      int
}

func WriterWrapper(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(code int) {
	if !w.wroteCode {
		w.setCode(code)
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *Writer) Write(data []byte) (int, error) {
	if !w.wroteCode {
		w.setCode(http.StatusOK)
	}

	return w.ResponseWriter.Write(data) //nolint:wrapcheck
}

func (w *Writer) Code() int {
	return w.code
}

func (w *Writer) setCode(code int) {
	w.wroteCode = true
	w.code = code
}

func NewMiddleware(metrics *HTTPServer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()

			ww := WriterWrapper(w)
			next.ServeHTTP(ww, r)
			method := r.Method + " " + r.URL.Path
			// Metrics
			metrics.Duration(method, now)
			metrics.TotalInc(method, ww.Code())

		}

		return http.HandlerFunc(fn)
	}
}
