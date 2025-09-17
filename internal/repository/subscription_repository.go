package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Subscription 表示用户订阅信息。
type Subscription struct {
	ID                   uint64
	UserID               uint64
	Name                 string
	PlanName             string
	Status               string
	TemplateID           uint64
	AvailableTemplateIDs []uint64
	Token                string
	ExpiresAt            time.Time
	TrafficTotalBytes    int64
	TrafficUsedBytes     int64
	DevicesLimit         int
	LastRefreshedAt      time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ListSubscriptionsOptions 控制订阅列表的分页与过滤。
type ListSubscriptionsOptions struct {
	Page      int
	PerPage   int
	Sort      string
	Direction string
	Query     string
	Status    string
}

// SubscriptionRepository 提供订阅相关操作。
type SubscriptionRepository interface {
	ListByUser(ctx context.Context, userID uint64, opts ListSubscriptionsOptions) ([]Subscription, int64, error)
	Get(ctx context.Context, id uint64) (Subscription, error)
	UpdateTemplate(ctx context.Context, subscriptionID uint64, templateID uint64, userID uint64) (Subscription, error)
}

type subscriptionRepository struct {
	db           *gorm.DB
	templateRepo SubscriptionTemplateRepository

	mu            sync.RWMutex
	subscriptions map[uint64]*Subscription
	nextID        uint64
}

// NewSubscriptionRepository 创建订阅仓储。
func NewSubscriptionRepository(db *gorm.DB, templateRepo SubscriptionTemplateRepository) SubscriptionRepository {
	repo := &subscriptionRepository{
		db:            db,
		templateRepo:  templateRepo,
		subscriptions: make(map[uint64]*Subscription),
		nextID:        1,
	}
	repo.seed()
	return repo
}

func (r *subscriptionRepository) seed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	templates, _, err := r.templateRepo.List(context.Background(), ListTemplatesOptions{PerPage: 50, IncludeDrafts: true})
	if err != nil {
		return
	}

	allowed := make([]uint64, 0, len(templates))
	defaultTemplate := uint64(0)
	for _, tpl := range templates {
		allowed = append(allowed, tpl.ID)
		if tpl.IsDefault && defaultTemplate == 0 {
			defaultTemplate = tpl.ID
		}
	}
	if defaultTemplate == 0 && len(allowed) > 0 {
		defaultTemplate = allowed[0]
	}

	now := time.Now().UTC()
	sub := &Subscription{
		Name:                 "VIP 全球高速",
		PlanName:             "VIP-Plus",
		Status:               "active",
		UserID:               2,
		TemplateID:           defaultTemplate,
		AvailableTemplateIDs: allowed,
		Token:                "demo-token-123",
		ExpiresAt:            now.Add(30 * 24 * time.Hour),
		TrafficTotalBytes:    1 << 40, // 1 TiB
		TrafficUsedBytes:     256 << 30,
		DevicesLimit:         5,
		LastRefreshedAt:      now.Add(-1 * time.Hour),
		CreatedAt:            now.Add(-48 * time.Hour),
		UpdatedAt:            now.Add(-2 * time.Hour),
	}
	r.addSubscriptionLocked(sub)
}

func (r *subscriptionRepository) ListByUser(ctx context.Context, userID uint64, opts ListSubscriptionsOptions) ([]Subscription, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListSubscriptionsOptions(opts)

	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.TrimSpace(strings.ToLower(opts.Query))
	status := strings.TrimSpace(strings.ToLower(opts.Status))

	items := make([]Subscription, 0)
	for _, sub := range r.subscriptions {
		if sub.UserID != userID {
			continue
		}
		if query != "" {
			if !strings.Contains(strings.ToLower(sub.Name), query) &&
				!strings.Contains(strings.ToLower(sub.PlanName), query) {
				continue
			}
		}
		if status != "" && !strings.EqualFold(sub.Status, status) {
			continue
		}
		items = append(items, cloneSubscription(sub))
	}

	sortField := opts.Sort
	desc := strings.EqualFold(opts.Direction, "desc")
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return subscriptionLess(items[j], items[i], sortField)
		}
		return subscriptionLess(items[i], items[j], sortField)
	})

	total := int64(len(items))
	start := (opts.Page - 1) * opts.PerPage
	if start >= len(items) {
		return []Subscription{}, total, nil
	}

	end := start + opts.PerPage
	if end > len(items) {
		end = len(items)
	}

	result := make([]Subscription, end-start)
	copy(result, items[start:end])
	return result, total, nil
}

func (r *subscriptionRepository) Get(ctx context.Context, id uint64) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	sub, ok := r.subscriptions[id]
	if !ok {
		return Subscription{}, ErrNotFound
	}

	return cloneSubscription(sub), nil
}

func (r *subscriptionRepository) UpdateTemplate(ctx context.Context, subscriptionID uint64, templateID uint64, userID uint64) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	sub, ok := r.subscriptions[subscriptionID]
	if !ok {
		return Subscription{}, ErrNotFound
	}
	if sub.UserID != userID {
		return Subscription{}, ErrForbidden
	}

	targetTemplate := templateID
	if targetTemplate == 0 {
		targetTemplate = sub.TemplateID
	}

	allowed := targetTemplate == sub.TemplateID
	if !allowed {
		for _, id := range sub.AvailableTemplateIDs {
			if id == targetTemplate {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		return Subscription{}, ErrForbidden
	}

	if _, err := r.templateRepo.Get(ctx, targetTemplate); err != nil {
		return Subscription{}, err
	}

	if sub.TemplateID != targetTemplate {
		sub.TemplateID = targetTemplate
		sub.UpdatedAt = time.Now().UTC()
	}

	return cloneSubscription(sub), nil
}

func (r *subscriptionRepository) addSubscriptionLocked(sub *Subscription) uint64 {
	id := r.nextID
	r.nextID++

	copied := cloneSubscription(sub)
	copied.ID = id
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = time.Now().UTC()
	}
	if copied.UpdatedAt.IsZero() {
		copied.UpdatedAt = copied.CreatedAt
	}

	r.subscriptions[id] = &copied
	return id
}

func normalizeListSubscriptionsOptions(opts ListSubscriptionsOptions) ListSubscriptionsOptions {
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
		opts.Sort = "expires_at"
	}
	opts.Sort = strings.ToLower(opts.Sort)
	if opts.Direction == "" {
		opts.Direction = "desc"
	}
	return opts
}

func subscriptionLess(a, b Subscription, field string) bool {
	switch field {
	case "name":
		if strings.EqualFold(a.Name, b.Name) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "plan_name":
		if strings.EqualFold(a.PlanName, b.PlanName) {
			return a.ID < b.ID
		}
		return strings.ToLower(a.PlanName) < strings.ToLower(b.PlanName)
	case "created_at":
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.ID < b.ID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	case "updated_at":
		if a.UpdatedAt.Equal(b.UpdatedAt) {
			return a.ID < b.ID
		}
		return a.UpdatedAt.Before(b.UpdatedAt)
	default: // expires_at
		if a.ExpiresAt.Equal(b.ExpiresAt) {
			return a.ID < b.ID
		}
		return a.ExpiresAt.Before(b.ExpiresAt)
	}
}

func cloneSubscription(sub *Subscription) Subscription {
	copied := *sub
	if sub.AvailableTemplateIDs != nil {
		copied.AvailableTemplateIDs = append([]uint64(nil), sub.AvailableTemplateIDs...)
	}
	return copied
}
