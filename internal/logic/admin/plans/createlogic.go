package plans

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// CreateLogic 处理套餐创建。
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 构造函数。
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建套餐。
func (l *CreateLogic) Create(req *types.AdminCreatePlanRequest) (*types.PlanSummary, error) {
	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "CNY"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}

	plan := repository.Plan{
		Name:              strings.TrimSpace(req.Name),
		Slug:              strings.TrimSpace(req.Slug),
		Description:       strings.TrimSpace(req.Description),
		Tags:              append([]string(nil), req.Tags...),
		Features:          append([]string(nil), req.Features...),
		PriceCents:        req.PriceCents,
		Currency:          strings.ToUpper(currency),
		DurationDays:      req.DurationDays,
		TrafficLimitBytes: req.TrafficLimitBytes,
		DevicesLimit:      req.DevicesLimit,
		SortOrder:         req.SortOrder,
		Status:            status,
		Visible:           req.Visible,
	}

	created, err := l.svcCtx.Repositories.Plan.Create(l.ctx, plan)
	if err != nil {
		return nil, err
	}

	summary := toPlanSummary(created)
	return &summary, nil
}
