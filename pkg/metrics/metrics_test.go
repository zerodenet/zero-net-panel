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

func TestObserveOrderCancel(t *testing.T) {
	before := testutil.ToFloat64(OrderCancelTotal.WithLabelValues("admin", "success"))

	ObserveOrderCancel("Admin", "success", 50*time.Millisecond)

	after := testutil.ToFloat64(OrderCancelTotal.WithLabelValues("admin", "success"))
	if diff := after - before; diff != 1 {
		t.Fatalf("expected cancellation counter increase by 1, got %.0f", diff)
	}
}

func TestObserveOrderPayment(t *testing.T) {
	before := testutil.ToFloat64(OrderPaymentTotal.WithLabelValues("balance", "user", "success"))

	ObserveOrderPayment("BALANCE", "User", "success", 80*time.Millisecond)

	after := testutil.ToFloat64(OrderPaymentTotal.WithLabelValues("balance", "user", "success"))
	if diff := after - before; diff != 1 {
		t.Fatalf("expected payment counter increase by 1, got %.0f", diff)
	}
}

func TestObserveOrderRefund(t *testing.T) {
	before := testutil.ToFloat64(OrderRefundTotal.WithLabelValues("admin", "error"))
	amountFamiliesBefore := testutil.CollectAndCount(OrderRefundAmount)

	ObserveOrderRefund("ADMIN", 123.45, "error", 110*time.Millisecond)

	after := testutil.ToFloat64(OrderRefundTotal.WithLabelValues("admin", "error"))
	if diff := after - before; diff != 1 {
		t.Fatalf("expected refund counter increase by 1, got %.0f", diff)
	}

	amountFamiliesAfter := testutil.CollectAndCount(OrderRefundAmount)
	if amountFamiliesAfter <= amountFamiliesBefore {
		t.Fatalf("expected refund amount histogram to collect samples")
	}
}
