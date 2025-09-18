package metrics

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "znp"

var (
	// NodeSyncTotal counts admin-triggered node synchronization attempts.
	NodeSyncTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "node_sync",
		Name:      "operations_total",
		Help:      "Total number of node synchronization attempts.",
	}, []string{"protocol", "result"})

	// NodeSyncDurationSeconds records the duration of node synchronization attempts.
	NodeSyncDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "node_sync",
		Name:      "duration_seconds",
		Help:      "Duration of node synchronization operations in seconds.",
		Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10},
	}, []string{"protocol", "result"})

	// OrderCreateTotal counts user order creation attempts grouped by payment method.
	OrderCreateTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "order_create",
		Name:      "requests_total",
		Help:      "Total number of user order creation attempts.",
	}, []string{"payment_method", "result"})

	// OrderCreateDurationSeconds records order creation latency grouped by payment method.
	OrderCreateDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "order_create",
		Name:      "duration_seconds",
		Help:      "Duration of user order creation flow in seconds.",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	}, []string{"payment_method", "result"})
)

// ObserveNodeSync records a node synchronization attempt with duration and outcome labels.
func ObserveNodeSync(protocol, result string, duration time.Duration) {
	sanitizedProtocol := strings.ToLower(strings.TrimSpace(protocol))
	if sanitizedProtocol == "" {
		sanitizedProtocol = "unknown"
	}
	sanitizedResult := normalizeResult(result)

	NodeSyncTotal.WithLabelValues(sanitizedProtocol, sanitizedResult).Inc()
	NodeSyncDurationSeconds.WithLabelValues(sanitizedProtocol, sanitizedResult).Observe(duration.Seconds())
}

// ObserveOrderCreate records an order creation attempt with duration, payment method and outcome labels.
func ObserveOrderCreate(paymentMethod, result string, duration time.Duration) {
	sanitizedMethod := strings.ToLower(strings.TrimSpace(paymentMethod))
	if sanitizedMethod == "" {
		sanitizedMethod = "unknown"
	}
	sanitizedResult := normalizeResult(result)

	OrderCreateTotal.WithLabelValues(sanitizedMethod, sanitizedResult).Inc()
	OrderCreateDurationSeconds.WithLabelValues(sanitizedMethod, sanitizedResult).Observe(duration.Seconds())
}

func normalizeResult(result string) string {
	normalized := strings.ToLower(strings.TrimSpace(result))
	if normalized == "" {
		return "unknown"
	}
	switch normalized {
	case "success", "error", "failure", "failed":
		if normalized == "failed" {
			return "error"
		}
		if normalized == "failure" {
			return "error"
		}
		return normalized
	default:
		return normalized
	}
}
