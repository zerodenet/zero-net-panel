package repository

import (
	"context"
	"strings"
)

func (r *balanceRepository) RecordRefund(ctx context.Context, userID uint64, tx BalanceTransaction) (BalanceTransaction, UserBalance, error) {
	if err := ctx.Err(); err != nil {
		return BalanceTransaction{}, UserBalance{}, err
	}
	if tx.AmountCents <= 0 {
		return BalanceTransaction{}, UserBalance{}, ErrInvalidArgument
	}

	if strings.TrimSpace(strings.ToLower(tx.Type)) == "" {
		tx.Type = "refund"
	} else {
		tx.Type = strings.ToLower(strings.TrimSpace(tx.Type))
	}

	return r.ApplyTransaction(ctx, userID, tx)
}
