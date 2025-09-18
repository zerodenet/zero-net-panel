package announcement

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toUserAnnouncementSummary(a repository.Announcement) types.UserAnnouncementSummary {
	var visibleTo *int64
	if a.VisibleTo != nil {
		ts := a.VisibleTo.Unix()
		visibleTo = &ts
	}
	var publishedAt *int64
	if a.PublishedAt != nil {
		ts := a.PublishedAt.Unix()
		publishedAt = &ts
	}
	return types.UserAnnouncementSummary{
		ID:          a.ID,
		Title:       a.Title,
		Content:     a.Content,
		Category:    a.Category,
		Audience:    a.Audience,
		IsPinned:    a.IsPinned,
		Priority:    a.Priority,
		VisibleFrom: a.VisibleFrom.Unix(),
		VisibleTo:   visibleTo,
		PublishedAt: publishedAt,
	}
}
