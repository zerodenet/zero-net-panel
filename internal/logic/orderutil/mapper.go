package orderutil

import (
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
func ToOrderDetail(order repository.Order, items []repository.OrderItem) types.OrderDetail {
	detail := types.OrderDetail{
		ID:            order.ID,
		Number:        order.Number,
		UserID:        order.UserID,
		Status:        order.Status,
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

	if len(items) == 0 {
		detail.Items = []types.OrderItem{}
		return detail
	}

	detail.Items = make([]types.OrderItem, 0, len(items))
	for _, item := range items {
		detail.Items = append(detail.Items, ToOrderItem(item))
	}
	return detail
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
		Type:              tx.Type,
		AmountCents:       tx.AmountCents,
		Currency:          tx.Currency,
		BalanceAfterCents: tx.BalanceAfterCents,
		Reference:         tx.Reference,
		Description:       tx.Description,
		Metadata:          tx.Metadata,
		CreatedAt:         tx.CreatedAt.UTC().Unix(),
	}
}
