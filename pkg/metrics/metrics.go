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

	// OrderCancelTotal counts order cancellation attempts grouped by actor.
	OrderCancelTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "order_cancel",
		Name:      "requests_total",
		Help:      "Total number of order cancellation attempts.",
	}, []string{"actor", "result"})

	// OrderCancelDurationSeconds records cancellation latency.
	OrderCancelDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "order_cancel",
		Name:      "duration_seconds",
		Help:      "Duration of order cancellation operations in seconds.",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2},
	}, []string{"actor", "result"})

	// OrderPaymentTotal counts manual payment confirmations.
	OrderPaymentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "order_payment",
		Name:      "requests_total",
		Help:      "Total number of manual order payment confirmations.",
	}, []string{"payment_method", "actor", "result"})

	// OrderPaymentDurationSeconds records manual payment latency.
	OrderPaymentDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "order_payment",
		Name:      "duration_seconds",
		Help:      "Duration of manual payment operations in seconds.",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2},
	}, []string{"payment_method", "actor", "result"})

	// OrderRefundTotal counts refund attempts.
	OrderRefundTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "order_refund",
		Name:      "requests_total",
		Help:      "Total number of order refund attempts.",
	}, []string{"actor", "result"})

	// OrderRefundDurationSeconds records refund latency.
	OrderRefundDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "order_refund",
		Name:      "duration_seconds",
		Help:      "Duration of order refund operations in seconds.",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2},
	}, []string{"actor", "result"})

	// OrderRefundAmount records the refund amount distribution (in currency units).
	OrderRefundAmount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "order_refund",
		Name:      "amount",
		Help:      "Distribution of refunded amount per operation (in currency units).",
		Buckets:   []float64{0.5, 1, 2, 5, 10, 20, 50, 100, 200},
	}, []string{"actor"})
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

// ObserveOrderCancel records an order cancellation attempt.
func ObserveOrderCancel(actor, result string, duration time.Duration) {
	sanitizedActor := normalizeActor(actor)
	sanitizedResult := normalizeResult(result)

	OrderCancelTotal.WithLabelValues(sanitizedActor, sanitizedResult).Inc()
	OrderCancelDurationSeconds.WithLabelValues(sanitizedActor, sanitizedResult).Observe(duration.Seconds())
}

// ObserveOrderPayment records a manual payment confirmation attempt.
func ObserveOrderPayment(paymentMethod, actor, result string, duration time.Duration) {
	sanitizedMethod := strings.ToLower(strings.TrimSpace(paymentMethod))
	if sanitizedMethod == "" {
		sanitizedMethod = "unknown"
	}
	sanitizedActor := normalizeActor(actor)
	sanitizedResult := normalizeResult(result)

	OrderPaymentTotal.WithLabelValues(sanitizedMethod, sanitizedActor, sanitizedResult).Inc()
	OrderPaymentDurationSeconds.WithLabelValues(sanitizedMethod, sanitizedActor, sanitizedResult).Observe(duration.Seconds())
}

// ObserveOrderRefund records a refund attempt, logging duration and amount.
func ObserveOrderRefund(actor string, amount float64, result string, duration time.Duration) {
	sanitizedActor := normalizeActor(actor)
	sanitizedResult := normalizeResult(result)

	if amount < 0 {
		amount = -amount
	}

	OrderRefundTotal.WithLabelValues(sanitizedActor, sanitizedResult).Inc()
	OrderRefundDurationSeconds.WithLabelValues(sanitizedActor, sanitizedResult).Observe(duration.Seconds())
	OrderRefundAmount.WithLabelValues(sanitizedActor).Observe(amount)
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

func normalizeActor(actor string) string {
	normalized := strings.ToLower(strings.TrimSpace(actor))
	if normalized == "" {
		return "unknown"
	}
	return normalized
}
