package account

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toBalanceResponse(balance repository.UserBalance, transactions []repository.BalanceTransaction, pagination types.PaginationMeta) types.UserBalanceResponse {
	result := types.UserBalanceResponse{
		UserID:       balance.UserID,
		BalanceCents: balance.BalanceCents,
		Currency:     balance.Currency,
		UpdatedAt:    balance.UpdatedAt.Unix(),
		Transactions: make([]types.BalanceTransactionSummary, 0, len(transactions)),
		Pagination:   pagination,
	}

	for _, tx := range transactions {
		result.Transactions = append(result.Transactions, toBalanceTransactionSummary(tx))
	}

	return result
}

func toBalanceTransactionSummary(tx repository.BalanceTransaction) types.BalanceTransactionSummary {
	return types.BalanceTransactionSummary{
		ID:                tx.ID,
		EntryType:         tx.Type,
		AmountCents:       tx.AmountCents,
		Currency:          tx.Currency,
		BalanceAfterCents: tx.BalanceAfterCents,
		Reference:         tx.Reference,
		Description:       tx.Description,
		Metadata:          tx.Metadata,
		CreatedAt:         tx.CreatedAt.Unix(),
	}
}
