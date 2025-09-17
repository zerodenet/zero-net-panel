package templates

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// HistoryLogic 查询模板历史。
type HistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHistoryLogic 构造函数。
func NewHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryLogic {
	return &HistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// History 返回模板历史。
func (l *HistoryLogic) History(req *types.AdminSubscriptionTemplateHistoryRequest) (*types.AdminSubscriptionTemplateHistoryResponse, error) {
	history, err := l.svcCtx.Repositories.SubscriptionTemplate.History(l.ctx, req.TemplateID)
	if err != nil {
		return nil, err
	}

	entries := make([]types.SubscriptionTemplateHistoryEntry, 0, len(history))
	for _, h := range history {
		entries = append(entries, toHistoryEntry(h))
	}

	return &types.AdminSubscriptionTemplateHistoryResponse{
		TemplateID: req.TemplateID,
		History:    entries,
	}, nil
}
