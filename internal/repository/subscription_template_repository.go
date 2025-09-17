package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TemplateVariable 描述模板可用变量。
type TemplateVariable struct {
	Type        string
	Required    bool
	Description string
	Default     any
}

// SubscriptionTemplate 描述订阅模板的草稿状态。
type SubscriptionTemplate struct {
	ID              uint64
	Name            string
	Description     string
	ClientType      string
	Format          string
	Content         string
	Variables       map[string]TemplateVariable
	IsDefault       bool
	Version         uint32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
	LastPublishedBy string
}

// SubscriptionTemplateHistory 记录每次发布的快照。
type SubscriptionTemplateHistory struct {
	TemplateID  uint64
	Version     uint32
	Content     string
	Variables   map[string]TemplateVariable
	Format      string
	Changelog   string
	PublishedAt time.Time
	PublishedBy string
}

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

	mu        sync.RWMutex
	templates map[uint64]*SubscriptionTemplate
	history   map[uint64][]*SubscriptionTemplateHistory
	nextID    uint64
}

// NewSubscriptionTemplateRepository 创建模板仓储。
func NewSubscriptionTemplateRepository(db *gorm.DB) SubscriptionTemplateRepository {
	repo := &subscriptionTemplateRepository{
		db:        db,
		templates: make(map[uint64]*SubscriptionTemplate),
		history:   make(map[uint64][]*SubscriptionTemplateHistory),
		nextID:    1,
	}
	repo.seed()
	return repo
}

func (r *subscriptionTemplateRepository) seed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()

	clash := &SubscriptionTemplate{
		Name:        "Clash Premium 默认模板",
		Description: "提供 Clash Premium YAML 订阅示例",
		ClientType:  "clash",
		Format:      "go_template",
		Content:     `# Clash Premium subscription\nproxies:\n  - name: {{ .subscription.name }}\n    type: trojan\n    server: {{ index .nodes 0 "hostname" }}\n    port: {{ index .nodes 0 "port" }}\n    password: {{ .subscription.token }}\n`,
		Variables: map[string]TemplateVariable{
			"subscription.name":  {Type: "string", Description: "订阅展示名称"},
			"subscription.token": {Type: "string", Description: "鉴权密钥", Required: true},
		},
		IsDefault: true,
		Version:   1,
		CreatedAt: now.Add(-48 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
	}
	clashID := r.addTemplateLocked(clash)
	r.appendHistoryLocked(r.templates[clashID], &SubscriptionTemplateHistory{
		TemplateID:  clashID,
		Version:     1,
		Content:     clash.Content,
		Variables:   cloneTemplateVariables(clash.Variables),
		Format:      clash.Format,
		Changelog:   "初始化模板",
		PublishedAt: clash.UpdatedAt,
		PublishedBy: "system",
	})

	singbox := &SubscriptionTemplate{
		Name:        "Sing-box JSON 模板",
		Description: "返回标准 JSON 结构的订阅",
		ClientType:  "sing-box",
		Format:      "go_template",
		Content:     `{{ toJSON .subscription }}`,
		Variables: map[string]TemplateVariable{
			"subscription": {Type: "object", Description: "订阅完整上下文", Required: true},
		},
		IsDefault: true,
		Version:   1,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now.Add(-3 * time.Hour),
	}
	singboxID := r.addTemplateLocked(singbox)
	r.appendHistoryLocked(r.templates[singboxID], &SubscriptionTemplateHistory{
		TemplateID:  singboxID,
		Version:     1,
		Content:     singbox.Content,
		Variables:   cloneTemplateVariables(singbox.Variables),
		Format:      singbox.Format,
		Changelog:   "初始化模板",
		PublishedAt: singbox.UpdatedAt,
		PublishedBy: "system",
	})
}

func (r *subscriptionTemplateRepository) List(ctx context.Context, opts ListTemplatesOptions) ([]SubscriptionTemplate, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListTemplatesOptions(opts)

	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.TrimSpace(strings.ToLower(opts.Query))
	clientType := strings.TrimSpace(strings.ToLower(opts.ClientType))
	format := strings.TrimSpace(strings.ToLower(opts.Format))

	items := make([]SubscriptionTemplate, 0, len(r.templates))
	for _, tpl := range r.templates {
		if query != "" {
			if !strings.Contains(strings.ToLower(tpl.Name), query) &&
				!strings.Contains(strings.ToLower(tpl.Description), query) {
				continue
			}
		}
		if clientType != "" && !strings.EqualFold(tpl.ClientType, clientType) {
			continue
		}
		if format != "" && !strings.EqualFold(tpl.Format, format) {
			continue
		}
		if !opts.IncludeDrafts && tpl.Version == 0 {
			continue
		}

		items = append(items, cloneTemplate(tpl))
	}

	sortField := opts.Sort
	desc := strings.EqualFold(opts.Direction, "desc")
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return templateLess(items[j], items[i], sortField)
		}
		return templateLess(items[i], items[j], sortField)
	})

	total := int64(len(items))
	start := (opts.Page - 1) * opts.PerPage
	if start >= len(items) {
		return []SubscriptionTemplate{}, total, nil
	}

	end := start + opts.PerPage
	if end > len(items) {
		end = len(items)
	}

	result := make([]SubscriptionTemplate, end-start)
	copy(result, items[start:end])
	return result, total, nil
}

func (r *subscriptionTemplateRepository) Create(ctx context.Context, input CreateSubscriptionTemplateInput) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	name := strings.TrimSpace(input.Name)
	clientType := strings.ToLower(strings.TrimSpace(input.ClientType))
	if name == "" || clientType == "" || strings.TrimSpace(input.Content) == "" {
		return SubscriptionTemplate{}, ErrInvalidArgument
	}

	format := strings.ToLower(strings.TrimSpace(input.Format))
	if format == "" {
		format = "go_template"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tpl := range r.templates {
		if strings.EqualFold(tpl.Name, name) && strings.EqualFold(tpl.ClientType, clientType) {
			return SubscriptionTemplate{}, ErrConflict
		}
	}

	now := time.Now().UTC()
	tpl := &SubscriptionTemplate{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		ClientType:  clientType,
		Format:      format,
		Content:     input.Content,
		Variables:   cloneTemplateVariables(input.Variables),
		IsDefault:   input.IsDefault,
		Version:     0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	id := r.addTemplateLocked(tpl)
	return cloneTemplate(r.templates[id]), nil
}

func (r *subscriptionTemplateRepository) Update(ctx context.Context, id uint64, input UpdateSubscriptionTemplateInput) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tpl, ok := r.templates[id]
	if !ok {
		return SubscriptionTemplate{}, ErrNotFound
	}

	if input.Name != nil {
		newName := strings.TrimSpace(*input.Name)
		if newName == "" {
			return SubscriptionTemplate{}, ErrInvalidArgument
		}
		for _, other := range r.templates {
			if other.ID == id {
				continue
			}
			if strings.EqualFold(other.Name, newName) && strings.EqualFold(other.ClientType, tpl.ClientType) {
				return SubscriptionTemplate{}, ErrConflict
			}
		}
		tpl.Name = newName
	}

	if input.Description != nil {
		tpl.Description = strings.TrimSpace(*input.Description)
	}
	if input.Format != nil {
		newFormat := strings.ToLower(strings.TrimSpace(*input.Format))
		if newFormat == "" {
			return SubscriptionTemplate{}, ErrInvalidArgument
		}
		tpl.Format = newFormat
	}
	if input.Content != nil {
		newContent := strings.TrimSpace(*input.Content)
		if newContent == "" {
			return SubscriptionTemplate{}, ErrInvalidArgument
		}
		tpl.Content = *input.Content
	}
	if input.Variables != nil {
		tpl.Variables = cloneTemplateVariables(input.Variables)
	}
	if input.IsDefault != nil {
		if *input.IsDefault && !tpl.IsDefault {
			r.resetDefaultLocked(tpl.ClientType)
		}
		tpl.IsDefault = *input.IsDefault
	}

	tpl.UpdatedAt = time.Now().UTC()
	r.templates[id] = tpl
	return cloneTemplate(tpl), nil
}

func (r *subscriptionTemplateRepository) Publish(ctx context.Context, id uint64, input PublishSubscriptionTemplateInput) (SubscriptionTemplate, SubscriptionTemplateHistory, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, SubscriptionTemplateHistory{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tpl, ok := r.templates[id]
	if !ok {
		return SubscriptionTemplate{}, SubscriptionTemplateHistory{}, ErrNotFound
	}

	now := time.Now().UTC()
	tpl.Version++
	tpl.UpdatedAt = now
	tpl.LastPublishedBy = strings.TrimSpace(input.Operator)
	publishedAt := now
	tpl.PublishedAt = &publishedAt

	history := &SubscriptionTemplateHistory{
		TemplateID:  tpl.ID,
		Version:     tpl.Version,
		Content:     tpl.Content,
		Variables:   cloneTemplateVariables(tpl.Variables),
		Format:      tpl.Format,
		Changelog:   strings.TrimSpace(input.Changelog),
		PublishedAt: now,
		PublishedBy: tpl.LastPublishedBy,
	}

	r.history[id] = append(r.history[id], history)

	return cloneTemplate(tpl), cloneTemplateHistory(history), nil
}

func (r *subscriptionTemplateRepository) History(ctx context.Context, id uint64) ([]SubscriptionTemplateHistory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, ok := r.history[id]
	if !ok {
		return []SubscriptionTemplateHistory{}, nil
	}

	result := make([]SubscriptionTemplateHistory, len(entries))
	for i, entry := range entries {
		result[i] = cloneTemplateHistory(entry)
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Version == result[j].Version {
			return result[i].PublishedAt.After(result[j].PublishedAt)
		}
		return result[i].Version > result[j].Version
	})

	return result, nil
}

func (r *subscriptionTemplateRepository) Get(ctx context.Context, id uint64) (SubscriptionTemplate, error) {
	if err := ctx.Err(); err != nil {
		return SubscriptionTemplate{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tpl, ok := r.templates[id]
	if !ok {
		return SubscriptionTemplate{}, ErrNotFound
	}

	return cloneTemplate(tpl), nil
}

func (r *subscriptionTemplateRepository) addTemplateLocked(tpl *SubscriptionTemplate) uint64 {
	id := r.nextID
	r.nextID++

	tplCopy := cloneTemplate(tpl)
	tplCopy.ID = id
	tplCopy.ClientType = strings.ToLower(strings.TrimSpace(tplCopy.ClientType))
	if tplCopy.ClientType == "" {
		tplCopy.ClientType = "generic"
	}
	tplCopy.Format = strings.ToLower(strings.TrimSpace(tplCopy.Format))
	if tplCopy.Format == "" {
		tplCopy.Format = "go_template"
	}
	if tplCopy.CreatedAt.IsZero() {
		tplCopy.CreatedAt = time.Now().UTC()
	}
	if tplCopy.UpdatedAt.IsZero() {
		tplCopy.UpdatedAt = tplCopy.CreatedAt
	}

	if tplCopy.IsDefault {
		r.resetDefaultLocked(tplCopy.ClientType)
	}

	r.templates[id] = &tplCopy
	return id
}

func (r *subscriptionTemplateRepository) appendHistoryLocked(tpl *SubscriptionTemplate, history *SubscriptionTemplateHistory) {
	if history == nil {
		return
	}

	entry := cloneTemplateHistory(history)
	if entry.TemplateID == 0 && tpl != nil {
		entry.TemplateID = tpl.ID
	}
	if entry.Version == 0 && tpl != nil {
		entry.Version = tpl.Version
	}
	if entry.PublishedAt.IsZero() {
		if tpl != nil {
			entry.PublishedAt = tpl.UpdatedAt
		} else {
			entry.PublishedAt = time.Now().UTC()
		}
	}

	r.history[entry.TemplateID] = append(r.history[entry.TemplateID], &entry)
}

func (r *subscriptionTemplateRepository) resetDefaultLocked(clientType string) {
	for _, tpl := range r.templates {
		if strings.EqualFold(tpl.ClientType, clientType) {
			tpl.IsDefault = false
		}
	}
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

func templateLess(a, b SubscriptionTemplate, field string) bool {
	switch field {
	case "name":
		if strings.EqualFold(a.Name, b.Name) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "client_type":
		if strings.EqualFold(a.ClientType, b.ClientType) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.ClientType) < strings.ToLower(b.ClientType)
	case "version":
		if a.Version == b.Version {
			return a.ID < b.ID
		}
		return a.Version < b.Version
	case "created_at":
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.ID < b.ID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	default: // updated_at
		if a.UpdatedAt.Equal(b.UpdatedAt) {
			return a.ID < b.ID
		}
		return a.UpdatedAt.Before(b.UpdatedAt)
	}
}

func cloneTemplate(tpl *SubscriptionTemplate) SubscriptionTemplate {
	copied := *tpl
	copied.Variables = cloneTemplateVariables(tpl.Variables)
	if tpl.PublishedAt != nil {
		ts := *tpl.PublishedAt
		copied.PublishedAt = &ts
	}
	return copied
}

func cloneTemplateHistory(history *SubscriptionTemplateHistory) SubscriptionTemplateHistory {
	copied := *history
	copied.Variables = cloneTemplateVariables(history.Variables)
	return copied
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
