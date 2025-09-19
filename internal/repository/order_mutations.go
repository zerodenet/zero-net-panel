package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpdateOrderStatusParams describes fields that can be changed when updating an order state.
type UpdateOrderStatusParams struct {
	Status        string
	PaymentMethod *string
	PaidAt        *time.Time
	CancelledAt   *time.Time
	MetadataPatch map[string]any
}

// AddRefundParams captures arguments when applying a refund to an order.
type AddRefundParams struct {
	AmountCents   int64
	RefundAt      time.Time
	MetadataPatch map[string]any
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint64, params UpdateOrderStatusParams) (Order, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, err
	}

	status := strings.TrimSpace(strings.ToLower(params.Status))
	if status == "" {
		return Order{}, ErrInvalidArgument
	}

	var result Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			return translateError(err)
		}

		now := time.Now().UTC()
		order.Status = status
		order.UpdatedAt = now

		fields := []string{"Status", "UpdatedAt"}

		if params.PaymentMethod != nil {
			method := strings.TrimSpace(*params.PaymentMethod)
			if method != "" {
				order.PaymentMethod = method
				fields = append(fields, "PaymentMethod")
			}
		}

		if params.PaidAt != nil {
			paidAt := params.PaidAt.UTC()
			order.PaidAt = &paidAt
			fields = append(fields, "PaidAt")
		}

		if params.CancelledAt != nil {
			cancelledAt := params.CancelledAt.UTC()
			order.CancelledAt = &cancelledAt
			fields = append(fields, "CancelledAt")
		}

		metadataChanged := false
		if len(params.MetadataPatch) > 0 {
			order.Metadata = mergeMetadata(order.Metadata, params.MetadataPatch)
			metadataChanged = true
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&order).Select(fields).Updates(order).Error; err != nil {
			return translateError(err)
		}

		if !metadataChanged && order.Metadata == nil {
			order.Metadata = map[string]any{}
		}
		result = order
		return nil
	})
	if err != nil {
		return Order{}, err
	}

	return result, nil
}

func (r *orderRepository) AddRefund(ctx context.Context, id uint64, params AddRefundParams) (Order, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, err
	}
	if params.AmountCents <= 0 {
		return Order{}, ErrInvalidArgument
	}

	refundAt := params.RefundAt
	if refundAt.IsZero() {
		refundAt = time.Now().UTC()
	} else {
		refundAt = refundAt.UTC()
	}

	var result Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			return translateError(err)
		}

		remaining := order.TotalCents - order.RefundedCents
		if remaining <= 0 || params.AmountCents > remaining {
			return ErrInvalidArgument
		}

		updateTime := time.Now().UTC()
		order.RefundedCents += params.AmountCents
		order.RefundedAt = &refundAt
		order.UpdatedAt = updateTime

		fields := []string{"RefundedCents", "RefundedAt", "UpdatedAt"}
		if len(params.MetadataPatch) > 0 {
			order.Metadata = mergeMetadata(order.Metadata, params.MetadataPatch)
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&order).Select(fields).Updates(order).Error; err != nil {
			return translateError(err)
		}

		result = order
		return nil
	})
	if err != nil {
		return Order{}, err
	}

	return result, nil
}

func mergeMetadata(base map[string]any, patch map[string]any) map[string]any {
	if base == nil {
		base = make(map[string]any, len(patch))
	}
	for k, v := range patch {
		base[k] = v
	}
	return base
}
