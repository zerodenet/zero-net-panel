package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// BalanceLogic 用户余额逻辑。
type BalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewBalanceLogic 构造函数。
func NewBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BalanceLogic {
	return &BalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Balance 获取余额及流水。
func (l *BalanceLogic) Balance(req *types.UserBalanceRequest) (*types.UserBalanceResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	balance, err := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
	if err != nil {
		return nil, err
	}

	opts := repository.ListBalanceTransactionsOptions{
		Page:    page,
		PerPage: perPage,
		Type:    req.Type,
	}
	transactions, total, err := l.svcCtx.Repositories.Balance.ListTransactions(l.ctx, user.ID, opts)
	if err != nil {
		return nil, err
	}

	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	resp := toBalanceResponse(balance, transactions, pagination)
	return &resp, nil
}
