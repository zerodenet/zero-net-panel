package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Plan represents purchasable subscription bundles similar to xboard 套餐。
type Plan struct {
	ID                uint64   `gorm:"primaryKey"`
	Name              string   `gorm:"size:255"`
	Slug              string   `gorm:"size:128;uniqueIndex"`
	Description       string   `gorm:"type:text"`
	Tags              []string `gorm:"serializer:json"`
	Features          []string `gorm:"serializer:json"`
	PriceCents        int64    `gorm:"column:price_cents"`
	Currency          string   `gorm:"size:16"`
	DurationDays      int      `gorm:"column:duration_days"`
	TrafficLimitBytes int64    `gorm:"column:traffic_limit_bytes"`
	DevicesLimit      int      `gorm:"column:devices_limit"`
	SortOrder         int      `gorm:"column:sort_order"`
	Status            string   `gorm:"size:32"`
	Visible           bool     `gorm:"column:is_visible"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName provides explicit table binding.
func (Plan) TableName() string { return "plans" }

// ListPlansOptions controls filtering and pagination for plan listing.
type ListPlansOptions struct {
	Page      int
	PerPage   int
	Sort      string
	Direction string
	Query     string
	Status    string
	Visible   *bool
}

// PlanRepository exposes persistence helpers for plan management.
type PlanRepository interface {
	List(ctx context.Context, opts ListPlansOptions) ([]Plan, int64, error)
	Create(ctx context.Context, plan Plan) (Plan, error)
	Update(ctx context.Context, id uint64, updates Plan) (Plan, error)
	Get(ctx context.Context, id uint64) (Plan, error)
}

type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository constructs the repository using a gorm DB.
func NewPlanRepository(db *gorm.DB) (PlanRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &planRepository{db: db}, nil
}

func (r *planRepository) List(ctx context.Context, opts ListPlansOptions) ([]Plan, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListPlansOptions(opts)

	base := r.db.WithContext(ctx).Model(&Plan{})

	if query := strings.TrimSpace(strings.ToLower(opts.Query)); query != "" {
		like := fmt.Sprintf("%%%s%%", query)
		base = base.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", like, like)
	}
	if status := strings.TrimSpace(strings.ToLower(opts.Status)); status != "" {
		base = base.Where("LOWER(status) = ?", status)
	}
	if opts.Visible != nil {
		base = base.Where("is_visible = ?", *opts.Visible)
	}

	countQuery := base.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Plan{}, 0, nil
	}

	orderClause := buildPlanOrderClause(opts.Sort, opts.Direction)
	offset := (opts.Page - 1) * opts.PerPage
	listQuery := base.Session(&gorm.Session{}).Order(orderClause).Limit(opts.PerPage).Offset(offset)

	var plans []Plan
	if err := listQuery.Find(&plans).Error; err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

func (r *planRepository) Create(ctx context.Context, plan Plan) (Plan, error) {
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	now := time.Now().UTC()
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = now
	}
	plan.UpdatedAt = now
	plan.Slug = normalizeSlug(plan.Slug, plan.Name)
	if plan.Status == "" {
		plan.Status = "draft"
	}

	if err := r.db.WithContext(ctx).Create(&plan).Error; err != nil {
		return Plan{}, translateError(err)
	}

	return plan, nil
}

func (r *planRepository) Update(ctx context.Context, id uint64, updates Plan) (Plan, error) {
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	updates.Slug = normalizeSlug(updates.Slug, updates.Name)
	updates.UpdatedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Model(&Plan{}).Where("id = ?", id).Updates(map[string]any{
		"name":                updates.Name,
		"slug":                updates.Slug,
		"description":         updates.Description,
		"tags":                updates.Tags,
		"features":            updates.Features,
		"price_cents":         updates.PriceCents,
		"currency":            updates.Currency,
		"duration_days":       updates.DurationDays,
		"traffic_limit_bytes": updates.TrafficLimitBytes,
		"devices_limit":       updates.DevicesLimit,
		"sort_order":          updates.SortOrder,
		"status":              updates.Status,
		"is_visible":          updates.Visible,
		"updated_at":          updates.UpdatedAt,
	}).Error; err != nil {
		return Plan{}, translateError(err)
	}

	return r.Get(ctx, id)
}

func (r *planRepository) Get(ctx context.Context, id uint64) (Plan, error) {
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	var plan Plan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		return Plan{}, translateError(err)
	}

	return plan, nil
}

func normalizeListPlansOptions(opts ListPlansOptions) ListPlansOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 || opts.PerPage > 100 {
		opts.PerPage = 20
	}
	return opts
}

func buildPlanOrderClause(sort, direction string) string {
	column := "sort_order"
	dir := "ASC"

	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "price":
		column = "price_cents"
	case "name":
		column = "name"
	case "updated":
		column = "updated_at"
	}

	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}

	return fmt.Sprintf("%s %s, id ASC", column, dir)
}

func normalizeSlug(slug, name string) string {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug != "" {
		return strings.ReplaceAll(slug, " ", "-")
	}
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, "--", "-")
	if normalized == "" {
		return fmt.Sprintf("plan-%d", time.Now().UnixNano())
	}
	return normalized
}
