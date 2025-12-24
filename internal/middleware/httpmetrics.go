package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zero-net-panel/zero-net-panel/pkg/metrics"
)

// HTTPMetricsMiddleware records request latency and status with normalized paths.
type HTTPMetricsMiddleware struct{}

// Handler wraps handlers with metrics instrumentation.
func (HTTPMetricsMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)

		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rec.status)
		metrics.ObserveHTTPRequest(path, r.Method, status, time.Since(start))
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func normalizePath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		if isNumeric(segment) {
			segments[i] = ":id"
		} else if len(segment) > 32 {
			segments[i] = segment[:32]
		}
	}
	normalized := strings.Join(segments, "/")
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return normalized
}

func isNumeric(segment string) bool {
	for _, r := range segment {
		if r < '0' || r > '9' {
			return false
		}
	}
	return segment != ""
}
