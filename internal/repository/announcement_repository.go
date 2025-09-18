package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Announcement represents platform level notices aligned with xboard 公告体系。
type Announcement struct {
	ID          uint64     `gorm:"primaryKey"`
	Title       string     `gorm:"size:255"`
	Content     string     `gorm:"type:text"`
	Category    string     `gorm:"size:32"`
	Status      string     `gorm:"size:32"`
	Audience    string     `gorm:"size:32"`
	IsPinned    bool       `gorm:"column:is_pinned"`
	Priority    int        `gorm:"column:priority"`
	VisibleFrom time.Time  `gorm:"column:visible_from"`
	VisibleTo   *time.Time `gorm:"column:visible_to"`
	PublishedAt *time.Time `gorm:"column:published_at"`
	PublishedBy string     `gorm:"size:128"`
	CreatedBy   string     `gorm:"size:128"`
	UpdatedBy   string     `gorm:"size:128"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName returns custom table name.
func (Announcement) TableName() string { return "announcements" }

// ListAnnouncementsOptions describes filters for admin listing.
type ListAnnouncementsOptions struct {
	Page      int
	PerPage   int
	Status    string
	Category  string
	Audience  string
	Query     string
	Sort      string
	Direction string
}

// AnnouncementRepository exposes CRUD helpers.
type AnnouncementRepository interface {
	List(ctx context.Context, opts ListAnnouncementsOptions) ([]Announcement, int64, error)
	ListActive(ctx context.Context, audience string, limit int) ([]Announcement, error)
	Create(ctx context.Context, announcement Announcement) (Announcement, error)
	Publish(ctx context.Context, id uint64, publishAt time.Time, visibleTo *time.Time, operator string) (Announcement, error)
	Get(ctx context.Context, id uint64) (Announcement, error)
}

type announcementRepository struct {
	db *gorm.DB
}

// NewAnnouncementRepository wires repository dependencies.
func NewAnnouncementRepository(db *gorm.DB) (AnnouncementRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &announcementRepository{db: db}, nil
}

func (r *announcementRepository) List(ctx context.Context, opts ListAnnouncementsOptions) ([]Announcement, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListAnnouncementsOptions(opts)
	base := r.db.WithContext(ctx).Model(&Announcement{})

	if status := strings.TrimSpace(strings.ToLower(opts.Status)); status != "" {
		base = base.Where("LOWER(status) = ?", status)
	}
	if category := strings.TrimSpace(strings.ToLower(opts.Category)); category != "" {
		base = base.Where("LOWER(category) = ?", category)
	}
	if audience := strings.TrimSpace(strings.ToLower(opts.Audience)); audience != "" {
		base = base.Where("LOWER(audience) = ?", audience)
	}
	if query := strings.TrimSpace(strings.ToLower(opts.Query)); query != "" {
		like := fmt.Sprintf("%%%s%%", query)
		base = base.Where("(LOWER(title) LIKE ? OR LOWER(content) LIKE ?)", like, like)
	}

	countQuery := base.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Announcement{}, 0, nil
	}

	orderClause := buildAnnouncementOrderClause(opts.Sort, opts.Direction)
	offset := (opts.Page - 1) * opts.PerPage
	listQuery := base.Session(&gorm.Session{}).Order(orderClause).Limit(opts.PerPage).Offset(offset)

	var announcements []Announcement
	if err := listQuery.Find(&announcements).Error; err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

func (r *announcementRepository) ListActive(ctx context.Context, audience string, limit int) ([]Announcement, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := r.db.WithContext(ctx).Model(&Announcement{}).
		Where("status = ?", "published").
		Where("visible_from <= ?", now).
		Where("visible_to IS NULL OR visible_to >= ?", now).
		Order("is_pinned DESC, priority DESC, published_at DESC, id DESC")

	if audience := strings.TrimSpace(strings.ToLower(audience)); audience != "" && audience != "all" {
		query = query.Where("LOWER(audience) IN ?", []string{audience, "all"})
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var announcements []Announcement
	if err := query.Find(&announcements).Error; err != nil {
		return nil, err
	}
	return announcements, nil
}

func (r *announcementRepository) Create(ctx context.Context, announcement Announcement) (Announcement, error) {
	if err := ctx.Err(); err != nil {
		return Announcement{}, err
	}

	now := time.Now().UTC()
	if announcement.Status == "" {
		announcement.Status = "draft"
	}
	if announcement.Audience == "" {
		announcement.Audience = "all"
	}
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(&announcement).Error; err != nil {
		return Announcement{}, translateError(err)
	}

	return announcement, nil
}

func (r *announcementRepository) Publish(ctx context.Context, id uint64, publishAt time.Time, visibleTo *time.Time, operator string) (Announcement, error) {
	if err := ctx.Err(); err != nil {
		return Announcement{}, err
	}

	updates := map[string]any{
		"status":       "published",
		"published_at": publishAt,
		"visible_from": publishAt,
		"updated_at":   publishAt,
		"updated_by":   operator,
	}
	if visibleTo != nil {
		updates["visible_to"] = *visibleTo
	} else {
		updates["visible_to"] = gorm.Expr("NULL")
	}
	if operator != "" && updates["updated_by"] == "" {
		updates["updated_by"] = operator
	}
	if err := r.db.WithContext(ctx).Model(&Announcement{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return Announcement{}, translateError(err)
	}

	return r.Get(ctx, id)
}

func (r *announcementRepository) Get(ctx context.Context, id uint64) (Announcement, error) {
	if err := ctx.Err(); err != nil {
		return Announcement{}, err
	}

	var announcement Announcement
	if err := r.db.WithContext(ctx).First(&announcement, id).Error; err != nil {
		return Announcement{}, translateError(err)
	}
	return announcement, nil
}

func normalizeListAnnouncementsOptions(opts ListAnnouncementsOptions) ListAnnouncementsOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 || opts.PerPage > 100 {
		opts.PerPage = 20
	}
	return opts
}

func buildAnnouncementOrderClause(sort, direction string) string {
	column := "published_at"
	dir := "DESC"

	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "created":
		column = "created_at"
	case "title":
		column = "title"
	case "priority":
		column = "priority"
	}

	if strings.EqualFold(direction, "asc") {
		dir = "ASC"
	}

	return fmt.Sprintf("is_pinned DESC, %s %s, id DESC", column, dir)
}
