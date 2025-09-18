package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TemplateVariable 描述模板可用变量。
type TemplateVariable struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

// SubscriptionTemplate 描述订阅模板的草稿状态。
type SubscriptionTemplate struct {
	ID              uint64                      `gorm:"primaryKey"`
	Name            string                      `gorm:"size:255;uniqueIndex:idx_templates_client_name,priority:2"`
	Description     string                      `gorm:"type:text"`
	ClientType      string                      `gorm:"size:64;uniqueIndex:idx_templates_client_name,priority:1"`
	Format          string                      `gorm:"size:32"`
	Content         string                      `gorm:"type:text"`
	Variables       map[string]TemplateVariable `gorm:"serializer:json"`
	IsDefault       bool
	Version         uint32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
	LastPublishedBy string `gorm:"size:128"`
}

// TableName 自定义订阅模板表名。
func (SubscriptionTemplate) TableName() string { return "subscription_templates" }

// SubscriptionTemplateHistory 记录每次发布的快照。
type SubscriptionTemplateHistory struct {
	TemplateID  uint64                      `gorm:"primaryKey;autoIncrement:false"`
	Version     uint32                      `gorm:"primaryKey;autoIncrement:false"`
	Content     string                      `gorm:"type:text"`
	Variables   map[string]TemplateVariable `gorm:"serializer:json"`
	Format      string                      `gorm:"size:32"`
	Changelog   string                      `gorm:"type:text"`
	PublishedAt time.Time
	PublishedBy string `gorm:"size:128"`
}

// TableName 自定义历史表名。
func (SubscriptionTemplateHistory) TableName() string { return "subscription_template_histories" }

// ListTemplatesOptions 控制列表查询行为。
type ListTemplatesOptions struct {
	Page          int
	PerPage       int
	Sort          string
	Direction     string
	Query         string
	ClientType    string
	Format        string
	IncludeDrafts bool
}

// CreateSubscriptionTemplateInput 用于创建模板。
type CreateSubscriptionTemplateInput struct {
	Name        string
	Description string
	ClientType  string
	Format      string
	Content     string
	Variables   map[string]TemplateVariable
	IsDefault   bool
}

// UpdateSubscriptionTemplateInput 用于更新模板。
type UpdateSubscriptionTemplateInput struct {
	Name        *string
	Description *string
	Format      *string
	Content     *string
	Variables   map[string]TemplateVariable
	IsDefault   *bool
}

// PublishSubscriptionTemplateInput 用于发布模板。
type PublishSubscriptionTemplateInput struct {
	Changelog string
	Operator  string
}

// SubscriptionTemplateRepository 定义模板操作接口。
type SubscriptionTemplateRepository interface {
	List(ctx context.Context, opts ListTemplatesOptions) ([]SubscriptionTemplate, int64, error)
	Create(ctx context.Context, input CreateSubscriptionTemplateInput) (SubscriptionTemplate, error)
	Update(ctx context.Context, id uint64, input UpdateSubscriptionTemplateInput) (SubscriptionTemplate, error)
	Publish(ctx context.Context, id uint64, input PublishSubscriptionTemplateInput) (SubscriptionTemplate, SubscriptionTemplateHistory, error)
	History(ctx context.Context, id uint64) ([]SubscriptionTemplateHistory, error)
	Get(ctx context.Context, id uint64) (SubscriptionTemplate, error)
}

type subscriptionTemplateRepository struct {
	db *gorm.DB
}

// NewSubscriptionTemplateRepository 创建模板仓储。
func NewSubscriptionTemplateRepository(db *gorm.DB) (SubscriptionTemplateRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &subscriptionTemplateRepository{db: db}, nil
}

func (r *subscriptionTemplateRepository) List(ctx context.Context, opts ListTemplatesOptions) ([]SubscriptionTemplate, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListTemplatesOptions(opts)

	base := r.db.WithContext(ctx).Model(&SubscriptionTemplate{})

	if query := strings.TrimSpace(strings.ToLower(opts.Query)); query != "" {
		like := fmt.Sprintf("%%%s%%", query)
		base = base.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", like, like)
	}
	if client := strings.TrimSpace(strings.ToLower(opts.ClientType)); client != "" {
		base = base.Where("LOWER(client_type) = ?", client)
	}
	if format := strings.TrimSpace(strings.ToLower(opts.Format)); format != "" {
		base = base.Where("LOWER(format) = ?", format)
	}
	if !opts.IncludeDrafts {
		base = base.Where("version > 0")
	}

	countQuery := base.Session(&gorm.Session{}).Distinct("subscription_templates.id")
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []SubscriptionTemplate{}, 0, nil
	}

	orderClause := buildTemplateOrderClause(opts.Sort, opts.Direction)
	offset := (opts.Page - 1) * opts.PerPage
	listQuery := base.Session(&gorm.Session{}).Distinct().Order(orderClause).Limit(opts.PerPage).Offset(offset)

	var templates []SubscriptionTemplate
	if err := listQuery.Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *subscriptionTemplateRepository) Create(ctx context.Context, input CreateSubscriptionTemplateInput) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	name := strings.TrimSpace(input.Name)
	clientType := strings.ToLower(strings.TrimSpace(input.ClientType))
	content := strings.TrimSpace(input.Content)
	if name == "" || clientType == "" || content == "" {
		return SubscriptionTemplate{}, ErrInvalidArgument
	}

	format := strings.ToLower(strings.TrimSpace(input.Format))
	if format == "" {
		format = "go_template"
	}

	tpl := SubscriptionTemplate{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		ClientType:  clientType,
		Format:      format,
		Content:     input.Content,
		Variables:   cloneTemplateVariables(input.Variables),
		IsDefault:   input.IsDefault,
		Version:     0,
	}

	now := time.Now().UTC()
	tpl.CreatedAt = now
	tpl.UpdatedAt = now

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&SubscriptionTemplate{}).
			Where("LOWER(name) = ? AND LOWER(client_type) = ?", strings.ToLower(name), clientType).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrConflict
		}

		if err := tx.Create(&tpl).Error; err != nil {
			return err
		}

		if tpl.IsDefault {
			if err := tx.Model(&SubscriptionTemplate{}).
				Where("client_type = ? AND id <> ?", tpl.ClientType, tpl.ID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return SubscriptionTemplate{}, translateError(err)
	}

	return tpl, nil
}

func (r *subscriptionTemplateRepository) Update(ctx context.Context, id uint64, input UpdateSubscriptionTemplateInput) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	var tpl SubscriptionTemplate
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tpl, id).Error; err != nil {
			return err
		}

		if input.Name != nil {
			newName := strings.TrimSpace(*input.Name)
			if newName == "" {
				return ErrInvalidArgument
			}
			var count int64
			if err := tx.Model(&SubscriptionTemplate{}).
				Where("LOWER(name) = ? AND LOWER(client_type) = ? AND id <> ?", strings.ToLower(newName), tpl.ClientType, tpl.ID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrConflict
			}
			tpl.Name = newName
		}

		if input.Description != nil {
			tpl.Description = strings.TrimSpace(*input.Description)
		}
		if input.Format != nil {
			newFormat := strings.ToLower(strings.TrimSpace(*input.Format))
			if newFormat == "" {
				return ErrInvalidArgument
			}
			tpl.Format = newFormat
		}
		if input.Content != nil {
			newContent := strings.TrimSpace(*input.Content)
			if newContent == "" {
				return ErrInvalidArgument
			}
			tpl.Content = *input.Content
		}
		if input.Variables != nil {
			tpl.Variables = cloneTemplateVariables(input.Variables)
		}
		if input.IsDefault != nil {
			if *input.IsDefault && !tpl.IsDefault {
				if err := tx.Model(&SubscriptionTemplate{}).
					Where("client_type = ? AND id <> ?", tpl.ClientType, tpl.ID).
					Update("is_default", false).Error; err != nil {
					return err
				}
			}
			tpl.IsDefault = *input.IsDefault
		}

		tpl.UpdatedAt = time.Now().UTC()
		return tx.Save(&tpl).Error
	})

	if err != nil {
		return SubscriptionTemplate{}, translateError(err)
	}

	return tpl, nil
}

func (r *subscriptionTemplateRepository) Publish(ctx context.Context, id uint64, input PublishSubscriptionTemplateInput) (SubscriptionTemplate, SubscriptionTemplateHistory, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, SubscriptionTemplateHistory{}, err
	}

	var tpl SubscriptionTemplate
	var history SubscriptionTemplateHistory

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tpl, id).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		tpl.Version++
		tpl.UpdatedAt = now
		tpl.LastPublishedBy = strings.TrimSpace(input.Operator)
		tpl.PublishedAt = &now

		if err := tx.Save(&tpl).Error; err != nil {
			return err
		}

		history = SubscriptionTemplateHistory{
			TemplateID:  tpl.ID,
			Version:     tpl.Version,
			Content:     tpl.Content,
			Variables:   cloneTemplateVariables(tpl.Variables),
			Format:      tpl.Format,
			Changelog:   strings.TrimSpace(input.Changelog),
			PublishedAt: now,
			PublishedBy: tpl.LastPublishedBy,
		}

		return tx.Create(&history).Error
	})

	if err != nil {
		return SubscriptionTemplate{}, SubscriptionTemplateHistory{}, translateError(err)
	}

	return tpl, history, nil
}

func (r *subscriptionTemplateRepository) History(ctx context.Context, id uint64) ([]SubscriptionTemplateHistory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var history []SubscriptionTemplateHistory
	if err := r.db.WithContext(ctx).
		Where("template_id = ?", id).
		Order("version DESC").
		Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}

func (r *subscriptionTemplateRepository) Get(ctx context.Context, id uint64) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	var tpl SubscriptionTemplate
	if err := r.db.WithContext(ctx).First(&tpl, id).Error; err != nil {
		return SubscriptionTemplate{}, translateError(err)
	}

	return tpl, nil
}

func buildTemplateOrderClause(field, direction string) string {
	column := "updated_at"
	switch strings.ToLower(field) {
	case "name":
		column = "name"
	case "client_type":
		column = "client_type"
	case "version":
		column = "version"
	case "created_at":
		column = "created_at"
	}

	dir := "ASC"
	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}

	return fmt.Sprintf("%s %s", column, dir)
}

func cloneTemplateVariables(vars map[string]TemplateVariable) map[string]TemplateVariable {
	if vars == nil {
		return nil
	}
	cloned := make(map[string]TemplateVariable, len(vars))
	for k, v := range vars {
		cloned[k] = v
	}
	return cloned
}

func normalizeListTemplatesOptions(opts ListTemplatesOptions) ListTemplatesOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 20
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100
	}
	if opts.Sort == "" {
		opts.Sort = "updated_at"
	}
	opts.Sort = strings.ToLower(opts.Sort)
	if opts.Direction == "" {
		opts.Direction = "desc"
	}
	return opts
}
