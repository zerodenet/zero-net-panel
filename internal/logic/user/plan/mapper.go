package plan

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toUserPlanSummary(plan repository.Plan) types.UserPlanSummary {
	return types.UserPlanSummary{
		ID:                plan.ID,
		Name:              plan.Name,
		Description:       plan.Description,
		Features:          append([]string(nil), plan.Features...),
		PriceCents:        plan.PriceCents,
		Currency:          plan.Currency,
		DurationDays:      plan.DurationDays,
		TrafficLimitBytes: plan.TrafficLimitBytes,
		DevicesLimit:      plan.DevicesLimit,
		Tags:              append([]string(nil), plan.Tags...),
	}
}
