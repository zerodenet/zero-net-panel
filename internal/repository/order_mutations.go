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

// UpdateOrderPaymentStateParams describes fields applicable when updating payment state on an order.
type UpdateOrderPaymentStateParams struct {
	PaymentStatus    string
	OrderStatus      *string
	PaymentIntentID  *string
	PaymentReference *string
	FailureCode      *string
	FailureMessage   *string
	PaidAt           *time.Time
	MetadataPatch    map[string]any
}

// UpdateOrderPaymentParams defines allowed modifications on an order payment record.
type UpdateOrderPaymentParams struct {
	Status         string
	Reference      *string
	FailureCode    *string
	FailureMessage *string
	MetadataPatch  map[string]any
	ProcessedAt    *time.Time
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

func (r *orderRepository) UpdatePaymentState(ctx context.Context, id uint64, params UpdateOrderPaymentStateParams) (Order, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, err
	}

	status := strings.TrimSpace(strings.ToLower(params.PaymentStatus))
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
		order.PaymentStatus = status
		order.UpdatedAt = now

		fields := []string{"PaymentStatus", "UpdatedAt"}

		if params.OrderStatus != nil {
			orderStatus := strings.TrimSpace(strings.ToLower(*params.OrderStatus))
			if orderStatus != "" {
				order.Status = orderStatus
				fields = append(fields, "Status")
			}
		}

		if params.PaymentIntentID != nil {
			order.PaymentIntentID = strings.TrimSpace(*params.PaymentIntentID)
			fields = append(fields, "PaymentIntentID")
		}

		if params.PaymentReference != nil {
			order.PaymentReference = strings.TrimSpace(*params.PaymentReference)
			fields = append(fields, "PaymentReference")
		}

		if params.FailureCode != nil {
			order.PaymentFailureCode = strings.TrimSpace(*params.FailureCode)
			fields = append(fields, "PaymentFailureCode")
		}

		if params.FailureMessage != nil {
			order.PaymentFailureReason = strings.TrimSpace(*params.FailureMessage)
			fields = append(fields, "PaymentFailureReason")
		}

		if params.PaidAt != nil {
			paidAt := params.PaidAt.UTC()
			order.PaidAt = &paidAt
			fields = append(fields, "PaidAt")
		}

		if len(params.MetadataPatch) > 0 {
			order.Metadata = mergeMetadata(order.Metadata, params.MetadataPatch)
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&order).Select(fields).Updates(order).Error; err != nil {
			return translateError(err)
		}

		if order.Metadata == nil {
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

func (r *orderRepository) UpdatePaymentRecord(ctx context.Context, id uint64, params UpdateOrderPaymentParams) (OrderPayment, error) {
	if err := ctx.Err(); err != nil {
		return OrderPayment{}, err
	}

	status := strings.TrimSpace(strings.ToLower(params.Status))
	if status == "" {
		return OrderPayment{}, ErrInvalidArgument
	}

	var result OrderPayment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment OrderPayment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, id).Error; err != nil {
			return translateError(err)
		}

		updateTime := time.Now().UTC()
		if params.ProcessedAt != nil {
			updateTime = params.ProcessedAt.UTC()
		}

		payment.Status = status
		payment.UpdatedAt = updateTime

		fields := []string{"Status", "UpdatedAt"}

		if params.Reference != nil {
			payment.Reference = strings.TrimSpace(*params.Reference)
			fields = append(fields, "Reference")
		}

		if params.FailureCode != nil {
			payment.FailureCode = strings.TrimSpace(*params.FailureCode)
			fields = append(fields, "FailureCode")
		}

		if params.FailureMessage != nil {
			payment.FailureMessage = strings.TrimSpace(*params.FailureMessage)
			fields = append(fields, "FailureMessage")
		}

		if len(params.MetadataPatch) > 0 {
			payment.Metadata = mergeMetadata(payment.Metadata, params.MetadataPatch)
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&payment).Select(fields).Updates(payment).Error; err != nil {
			return translateError(err)
		}

		if payment.Metadata == nil {
			payment.Metadata = map[string]any{}
		}
		result = payment
		return nil
	})
	if err != nil {
		return OrderPayment{}, err
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
