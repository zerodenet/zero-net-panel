package announcements

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toAnnouncementSummary(a repository.Announcement) types.AnnouncementSummary {
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
	return types.AnnouncementSummary{
		ID:          a.ID,
		Title:       a.Title,
		Content:     a.Content,
		Category:    a.Category,
		Status:      a.Status,
		Audience:    a.Audience,
		IsPinned:    a.IsPinned,
		Priority:    a.Priority,
		VisibleFrom: a.VisibleFrom.Unix(),
		VisibleTo:   visibleTo,
		PublishedAt: publishedAt,
		PublishedBy: a.PublishedBy,
		CreatedBy:   a.CreatedBy,
		UpdatedBy:   a.UpdatedBy,
		CreatedAt:   a.CreatedAt.Unix(),
		UpdatedAt:   a.UpdatedAt.Unix(),
	}
}
