package orderutil

import (
	"strings"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// ToOrderItem converts repository order item into API representation.
func ToOrderItem(item repository.OrderItem) types.OrderItem {
	return types.OrderItem{
		ID:             item.ID,
		OrderID:        item.OrderID,
		ItemType:       item.ItemType,
		ItemID:         item.ItemID,
		Name:           item.Name,
		Quantity:       item.Quantity,
		UnitPriceCents: item.UnitPriceCents,
		Currency:       item.Currency,
		SubtotalCents:  item.SubtotalCents,
		Metadata:       item.Metadata,
		CreatedAt:      item.CreatedAt.UTC().Unix(),
	}
}

// ToOrderDetail hydrates an order with its items.
func ToOrderDetail(order repository.Order, items []repository.OrderItem, refunds []repository.OrderRefund, payments []repository.OrderPayment) types.OrderDetail {
	detail := types.OrderDetail{
		ID:            order.ID,
		Number:        order.Number,
		UserID:        order.UserID,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		TotalCents:    order.TotalCents,
		RefundedCents: order.RefundedCents,
		Currency:      order.Currency,
		PaymentMethod: order.PaymentMethod,
		PlanSnapshot:  order.PlanSnapshot,
		Metadata:      order.Metadata,
		CreatedAt:     order.CreatedAt.UTC().Unix(),
		UpdatedAt:     order.UpdatedAt.UTC().Unix(),
	}

	if order.PlanID != nil {
		detail.PlanID = order.PlanID
	}

	if strings.TrimSpace(order.PaymentIntentID) != "" {
		detail.PaymentIntentID = order.PaymentIntentID
	}

	if strings.TrimSpace(order.PaymentReference) != "" {
		detail.PaymentReference = order.PaymentReference
	}

	if strings.TrimSpace(order.PaymentFailureCode) != "" {
		detail.PaymentFailureCode = order.PaymentFailureCode
	}

	if strings.TrimSpace(order.PaymentFailureReason) != "" {
		detail.PaymentFailureMessage = order.PaymentFailureReason
	}

	if order.PaidAt != nil {
		paid := order.PaidAt.UTC().Unix()
		detail.PaidAt = &paid
	}

	if order.CancelledAt != nil {
		cancelled := order.CancelledAt.UTC().Unix()
		detail.CancelledAt = &cancelled
	}

	if order.RefundedAt != nil {
		refunded := order.RefundedAt.UTC().Unix()
		detail.RefundedAt = &refunded
	}

	detail.Items = make([]types.OrderItem, 0, len(items))
	for _, item := range items {
		detail.Items = append(detail.Items, ToOrderItem(item))
	}

	detail.Refunds = make([]types.OrderRefund, 0, len(refunds))
	for _, refund := range refunds {
		detail.Refunds = append(detail.Refunds, ToOrderRefund(refund))
	}

	detail.Payments = make([]types.OrderPayment, 0, len(payments))
	for _, payment := range payments {
		detail.Payments = append(detail.Payments, ToOrderPayment(payment))
	}

	return detail
}

// ToOrderRefund converts repository refund into API representation.
func ToOrderRefund(refund repository.OrderRefund) types.OrderRefund {
	return types.OrderRefund{
		ID:          refund.ID,
		OrderID:     refund.OrderID,
		AmountCents: refund.AmountCents,
		Reason:      refund.Reason,
		Reference:   refund.Reference,
		Metadata:    refund.Metadata,
		CreatedAt:   refund.CreatedAt.UTC().Unix(),
	}
}

// ToBalanceSnapshot converts repository balance into API snapshot.
func ToBalanceSnapshot(balance repository.UserBalance) types.BalanceSnapshot {
	return types.BalanceSnapshot{
		UserID:       balance.UserID,
		BalanceCents: balance.BalanceCents,
		Currency:     balance.Currency,
		UpdatedAt:    balance.UpdatedAt.UTC().Unix(),
	}
}

// ToBalanceTransactionView converts repository transaction for API responses.
func ToBalanceTransactionView(tx repository.BalanceTransaction) types.BalanceTransactionSummary {
	return types.BalanceTransactionSummary{
		ID:                tx.ID,
		EntryType:         tx.Type,
		AmountCents:       tx.AmountCents,
		Currency:          tx.Currency,
		BalanceAfterCents: tx.BalanceAfterCents,
		Reference:         tx.Reference,
		Description:       tx.Description,
		Metadata:          tx.Metadata,
		CreatedAt:         tx.CreatedAt.UTC().Unix(),
	}
}

func ToOrderPayment(payment repository.OrderPayment) types.OrderPayment {
	return types.OrderPayment{
		ID:             payment.ID,
		OrderID:        payment.OrderID,
		Provider:       payment.Provider,
		Method:         payment.Method,
		IntentID:       payment.IntentID,
		Reference:      payment.Reference,
		Status:         payment.Status,
		AmountCents:    payment.AmountCents,
		Currency:       payment.Currency,
		FailureCode:    payment.FailureCode,
		FailureMessage: payment.FailureMessage,
		Metadata:       payment.Metadata,
		CreatedAt:      payment.CreatedAt.UTC().Unix(),
		UpdatedAt:      payment.UpdatedAt.UTC().Unix(),
	}
}
