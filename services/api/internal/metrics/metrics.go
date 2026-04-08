package metrics

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type Metrics struct {
	requests uint64
}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&m.requests, 1)
		next.ServeHTTP(w, r)
	})
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("# metrics scaffolding\n"))
		_, _ = w.Write([]byte("http_requests_total "))
		_, _ = w.Write([]byte(strconv.FormatUint(atomic.LoadUint64(&m.requests), 10)))
		_, _ = w.Write([]byte("\n"))
	})
}
