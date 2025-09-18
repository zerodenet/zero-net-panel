package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestObserveNodeSync(t *testing.T) {
	const protocol = "unit-protocol"
	const result = "success"

	before := testutil.ToFloat64(NodeSyncTotal.WithLabelValues(protocol, result))

	ObserveNodeSync(protocol, result, 250*time.Millisecond)

	after := testutil.ToFloat64(NodeSyncTotal.WithLabelValues(protocol, result))

	if diff := after - before; diff != 1 {
		t.Fatalf("expected counter to increase by 1, got %.0f", diff)
	}
}

func TestObserveOrderCreate(t *testing.T) {
	const method = "test-method"
	const result = "error"

	before := testutil.ToFloat64(OrderCreateTotal.WithLabelValues(method, "error"))

	ObserveOrderCreate(method, result, 120*time.Millisecond)

	after := testutil.ToFloat64(OrderCreateTotal.WithLabelValues(method, "error"))

	if diff := after - before; diff != 1 {
		t.Fatalf("expected counter to increase by 1, got %.0f", diff)
	}
}
