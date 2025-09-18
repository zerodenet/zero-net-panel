package plans

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UpdateLogic 处理套餐更新。
type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateLogic 构造函数。
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Update 更新套餐。
func (l *UpdateLogic) Update(req *types.AdminUpdatePlanRequest) (*types.PlanSummary, error) {
	plan, err := l.svcCtx.Repositories.Plan.Get(l.ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		plan.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil {
		plan.Slug = strings.TrimSpace(*req.Slug)
	}
	if req.Description != nil {
		plan.Description = strings.TrimSpace(*req.Description)
	}
	if req.Tags != nil {
		plan.Tags = append([]string(nil), req.Tags...)
	}
	if req.Features != nil {
		plan.Features = append([]string(nil), req.Features...)
	}
	if req.PriceCents != nil {
		plan.PriceCents = *req.PriceCents
	}
	if req.Currency != nil {
		plan.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.DurationDays != nil {
		plan.DurationDays = *req.DurationDays
	}
	if req.TrafficLimitBytes != nil {
		plan.TrafficLimitBytes = *req.TrafficLimitBytes
	}
	if req.DevicesLimit != nil {
		plan.DevicesLimit = *req.DevicesLimit
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		plan.Status = strings.TrimSpace(*req.Status)
	}
	if req.Visible != nil {
		plan.Visible = *req.Visible
	}

	updated, err := l.svcCtx.Repositories.Plan.Update(l.ctx, req.PlanID, plan)
	if err != nil {
		return nil, err
	}

	summary := toPlanSummary(updated)
	return &summary, nil
}
