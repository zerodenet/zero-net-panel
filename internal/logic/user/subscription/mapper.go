package subscription

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toUserSummary(sub repository.Subscription) types.UserSubscriptionSummary {
	summary := types.UserSubscriptionSummary{
		ID:                   sub.ID,
		Name:                 sub.Name,
		PlanName:             sub.PlanName,
		Status:               sub.Status,
		TemplateID:           sub.TemplateID,
		AvailableTemplateIDs: append([]uint64(nil), sub.AvailableTemplateIDs...),
		ExpiresAt:            sub.ExpiresAt.Unix(),
		TrafficTotalBytes:    sub.TrafficTotalBytes,
		TrafficUsedBytes:     sub.TrafficUsedBytes,
		DevicesLimit:         sub.DevicesLimit,
		LastRefreshedAt:      sub.LastRefreshedAt.Unix(),
	}
	return summary
}
