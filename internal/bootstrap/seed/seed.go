package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	adminroutes "github.com/zero-net-panel/zero-net-panel/internal/admin/routes"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
)

// Run applies demonstration data to the database if the target tables are empty.
func Run(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedAdminModules(tx); err != nil {
			return err
		}
		if err := seedUsers(tx); err != nil {
			return err
		}
		if err := seedNodes(tx); err != nil {
			return err
		}
		if err := seedTemplates(tx); err != nil {
			return err
		}
		if err := seedSubscriptions(tx); err != nil {
			return err
		}
		if err := seedPlans(tx); err != nil {
			return err
		}
		if err := seedAnnouncements(tx); err != nil {
			return err
		}
		if err := seedUserBalances(tx); err != nil {
			return err
		}
		if err := seedOrders(tx); err != nil {
			return err
		}
		if err := seedSecuritySettings(tx); err != nil {
			return err
		}
		return nil
	})
}

func seedUsers(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	users := []repository.User{
		{
			Email:        "admin@example.com",
			DisplayName:  "运营管理员",
			PasswordHash: "$2a$10$OmiVLT.Awz75.D1g1Rvm7.TPPaB399VUCpQCJiBCnWGEN2L4IyJTe",
			Roles:        []string{"admin", "user"},
			Status:       "active",
			LastLoginAt:  now.Add(-48 * time.Hour),
			CreatedAt:    now.Add(-72 * time.Hour),
			UpdatedAt:    now.Add(-24 * time.Hour),
		},
		{
			Email:        "user@example.com",
			DisplayName:  "高级会员",
			PasswordHash: "$2a$10$OmiVLT.Awz75.D1g1Rvm7.TPPaB399VUCpQCJiBCnWGEN2L4IyJTe",
			Roles:        []string{"user"},
			Status:       "active",
			LastLoginAt:  now.Add(-6 * time.Hour),
			CreatedAt:    now.Add(-48 * time.Hour),
			UpdatedAt:    now.Add(-12 * time.Hour),
		},
	}

	return tx.Create(&users).Error
}

func seedAdminModules(tx *gorm.DB) error {
	var existing []repository.AdminModule
	if err := tx.Find(&existing).Error; err != nil {
		return err
	}

	if len(existing) > 0 {
		for _, module := range existing {
			normalized := adminroutes.Normalize(module.Route, "admin")
			if normalized == module.Route {
				continue
			}
			if err := tx.Model(&repository.AdminModule{}).
				Where("id = ?", module.ID).
				Update("route", normalized).Error; err != nil {
				return err
			}
		}
		return nil
	}

	now := time.Now().UTC()
	modules := []repository.AdminModule{
		{
			Key:         "dashboard",
			Name:        "运营总览",
			Description: "可视化展示系统运行情况、节点健康度与订阅概况",
			Icon:        "dashboard",
			Route:       adminroutes.Normalize("/dashboard", ""),
			Permissions: []string{"admin"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "nodes",
			Name:        "节点管理",
			Description: "维护边缘节点与内核运行状态",
			Icon:        "deployment-unit",
			Route:       adminroutes.Normalize("/nodes", ""),
			Permissions: []string{"admin", "ops"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "subscriptions",
			Name:        "订阅模板",
			Description: "设计多种客户端的模板与变量",
			Icon:        "layout",
			Route:       adminroutes.Normalize("/subscription-templates", ""),
			Permissions: []string{"admin", "product"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Key:         "security",
			Name:        "安全配置",
			Description: "管理第三方调用的加密与签名开关",
			Icon:        "safety",
			Route:       adminroutes.Normalize("/security-settings", ""),
			Permissions: []string{"admin"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	return tx.Create(&modules).Error
}

func seedNodes(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.Node{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()

	hkNode := repository.Node{
		Name:         "edge-hk-1",
		Region:       "Hong Kong",
		Country:      "HK",
		ISP:          "HKIX",
		Status:       "online",
		Tags:         []string{"premium", "asia"},
		Protocols:    []string{"http", "grpc"},
		CapacityMbps: 1000,
		Description:  "香港高带宽边缘节点示例",
		CreatedAt:    now.Add(-36 * time.Hour),
		UpdatedAt:    now.Add(-30 * time.Minute),
		LastSyncedAt: now.Add(-30 * time.Minute),
	}
	if err := tx.Create(&hkNode).Error; err != nil {
		return err
	}

	hkKernels := []repository.NodeKernel{
		{
			NodeID:       hkNode.ID,
			Protocol:     "http",
			Endpoint:     "https://kernel-hk.example.com/api",
			Revision:     "rev-hk-http",
			Status:       "synced",
			Config:       map[string]any{"transport": "ws", "heartbeat": 30},
			LastSyncedAt: hkNode.LastSyncedAt,
			CreatedAt:    hkNode.CreatedAt,
			UpdatedAt:    hkNode.UpdatedAt,
		},
		{
			NodeID:       hkNode.ID,
			Protocol:     "grpc",
			Endpoint:     "kernel-hk.example.com:9000",
			Revision:     "rev-hk-grpc",
			Status:       "synced",
			Config:       map[string]any{"transport": "grpc", "heartbeat": 15},
			LastSyncedAt: hkNode.LastSyncedAt.Add(-5 * time.Minute),
			CreatedAt:    hkNode.CreatedAt,
			UpdatedAt:    hkNode.UpdatedAt,
		},
	}
	if err := tx.Create(&hkKernels).Error; err != nil {
		return err
	}

	laNode := repository.Node{
		Name:         "edge-la-1",
		Region:       "Los Angeles",
		Country:      "US",
		ISP:          "NTT",
		Status:       "maintenance",
		Tags:         []string{"standard", "america"},
		Protocols:    []string{"http"},
		CapacityMbps: 600,
		Description:  "北美标准线路示例节点",
		CreatedAt:    now.Add(-72 * time.Hour),
		UpdatedAt:    now.Add(-6 * time.Hour),
		LastSyncedAt: now.Add(-12 * time.Hour),
	}
	if err := tx.Create(&laNode).Error; err != nil {
		return err
	}

	laKernel := repository.NodeKernel{
		NodeID:       laNode.ID,
		Protocol:     "http",
		Endpoint:     "https://kernel-la.example.com/api",
		Revision:     "rev-la-http",
		Status:       "synced",
		Config:       map[string]any{"transport": "http", "heartbeat": 45},
		LastSyncedAt: laNode.LastSyncedAt,
		CreatedAt:    laNode.CreatedAt,
		UpdatedAt:    laNode.UpdatedAt,
	}
	return tx.Create(&laKernel).Error
}

func seedTemplates(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.SubscriptionTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()

	clashPublishedAt := now.Add(-2 * time.Hour)
	clashTemplate := repository.SubscriptionTemplate{
		Name:        "Clash Premium 默认模板",
		Description: "提供 Clash Premium YAML 订阅示例",
		ClientType:  "clash",
		Format:      "go_template",
		Content:     `# Clash Premium subscription\nproxies:\n  - name: {{ .subscription.name }}\n    type: trojan\n    server: {{ index .nodes 0 "hostname" }}\n    port: {{ index .nodes 0 "port" }}\n    password: {{ .subscription.token }}\n`,
                Variables: map[string]repository.TemplateVariable{
                        "subscription.name":  {ValueType: "string", Description: "订阅展示名称"},
                        "subscription.token": {ValueType: "string", Description: "鉴权密钥", Required: true},
		},
		IsDefault:       true,
		Version:         1,
		CreatedAt:       now.Add(-48 * time.Hour),
		UpdatedAt:       clashPublishedAt,
		PublishedAt:     &clashPublishedAt,
		LastPublishedBy: "system",
	}
	if err := tx.Create(&clashTemplate).Error; err != nil {
		return err
	}

	clashHistory := repository.SubscriptionTemplateHistory{
		TemplateID:  clashTemplate.ID,
		Version:     clashTemplate.Version,
		Content:     clashTemplate.Content,
		Variables:   clashTemplate.Variables,
		Format:      clashTemplate.Format,
		Changelog:   "初始化模板",
		PublishedAt: clashPublishedAt,
		PublishedBy: "system",
	}
	if err := tx.Create(&clashHistory).Error; err != nil {
		return err
	}

	singPublishedAt := now.Add(-3 * time.Hour)
	singTemplate := repository.SubscriptionTemplate{
		Name:        "Sing-box JSON 模板",
		Description: "返回标准 JSON 结构的订阅",
		ClientType:  "sing-box",
		Format:      "go_template",
		Content:     `{{ toJSON .subscription }}`,
                Variables: map[string]repository.TemplateVariable{
                        "subscription": {ValueType: "object", Description: "订阅完整上下文", Required: true},
                },
		IsDefault:       true,
		Version:         1,
		CreatedAt:       now.Add(-24 * time.Hour),
		UpdatedAt:       singPublishedAt,
		PublishedAt:     &singPublishedAt,
		LastPublishedBy: "system",
	}
	if err := tx.Create(&singTemplate).Error; err != nil {
		return err
	}

	singHistory := repository.SubscriptionTemplateHistory{
		TemplateID:  singTemplate.ID,
		Version:     singTemplate.Version,
		Content:     singTemplate.Content,
		Variables:   singTemplate.Variables,
		Format:      singTemplate.Format,
		Changelog:   "初始化模板",
		PublishedAt: singPublishedAt,
		PublishedBy: "system",
	}
	if err := tx.Create(&singHistory).Error; err != nil {
		return err
	}

	return nil
}

func seedSubscriptions(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.Subscription{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var member repository.User
	if err := tx.Where("LOWER(email) = ?", strings.ToLower("user@example.com")).First(&member).Error; err != nil {
		return err
	}

	var templates []repository.SubscriptionTemplate
	if err := tx.Find(&templates).Error; err != nil {
		return err
	}
	if len(templates) == 0 {
		return nil
	}

	allowed := make([]uint64, 0, len(templates))
	var defaultTemplateID uint64
	for _, tpl := range templates {
		allowed = append(allowed, tpl.ID)
		if tpl.IsDefault && defaultTemplateID == 0 {
			defaultTemplateID = tpl.ID
		}
	}
	if defaultTemplateID == 0 {
		defaultTemplateID = templates[0].ID
	}

	now := time.Now().UTC()
	subscription := repository.Subscription{
		UserID:               member.ID,
		Name:                 "VIP 全球高速",
		PlanName:             "VIP-Plus",
		Status:               "active",
		TemplateID:           defaultTemplateID,
		AvailableTemplateIDs: allowed,
		Token:                "demo-token-123",
		ExpiresAt:            now.Add(30 * 24 * time.Hour),
		TrafficTotalBytes:    1 << 40,
		TrafficUsedBytes:     256 << 30,
		DevicesLimit:         5,
		LastRefreshedAt:      now.Add(-1 * time.Hour),
		CreatedAt:            now.Add(-48 * time.Hour),
		UpdatedAt:            now.Add(-2 * time.Hour),
	}

	return tx.Create(&subscription).Error
}

func seedPlans(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.Plan{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	plans := []repository.Plan{
		{
			Name:              "标准套餐",
			Slug:              "standard",
			Description:       "基础体验套餐，适合日常浏览与轻量访问",
			Tags:              []string{"入门", "轻量"},
			Features:          []string{"全球节点访问", "每月 200GB 流量", "3 台设备同步"},
			PriceCents:        2999,
			Currency:          "CNY",
			DurationDays:      30,
			TrafficLimitBytes: 200 << 30,
			DevicesLimit:      3,
			SortOrder:         10,
			Status:            "active",
			Visible:           true,
			CreatedAt:         now.Add(-72 * time.Hour),
			UpdatedAt:         now.Add(-24 * time.Hour),
		},
		{
			Name:              "旗舰套餐",
			Slug:              "premium",
			Description:       "旗舰级套餐，满足企业远程办公与大流量需求",
			Tags:              []string{"旗舰", "高可用"},
			Features:          []string{"专属高速通道", "每月 1TB 流量", "10 台设备", "高级监控报告"},
			PriceCents:        9999,
			Currency:          "CNY",
			DurationDays:      30,
			TrafficLimitBytes: 1 << 40,
			DevicesLimit:      10,
			SortOrder:         20,
			Status:            "active",
			Visible:           true,
			CreatedAt:         now.Add(-96 * time.Hour),
			UpdatedAt:         now.Add(-12 * time.Hour),
		},
	}

	return tx.Create(&plans).Error
}

func seedAnnouncements(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.Announcement{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	nextWeek := now.Add(7 * 24 * time.Hour)
	announcements := []repository.Announcement{
		{
			Title:       "Zero Net Panel Beta 发布",
			Content:     "我们上线了全新的 Zero Net Panel 后台，提供节点监控、套餐管理与公告推送功能。欢迎社区体验反馈。",
			Category:    "product",
			Status:      "published",
			Audience:    "all",
			IsPinned:    true,
			Priority:    90,
			VisibleFrom: now.Add(-24 * time.Hour),
			VisibleTo:   &nextWeek,
			PublishedAt: &now,
			PublishedBy: "system",
			CreatedBy:   "system",
			UpdatedBy:   "system",
			CreatedAt:   now.Add(-24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Title:       "节点维护通知",
			Content:     "香港节点将在本周日 02:00-04:00 进行升级维护，期间服务可能短暂中断，请合理安排使用时间。",
			Category:    "maintenance",
			Status:      "published",
			Audience:    "user",
			IsPinned:    false,
			Priority:    60,
			VisibleFrom: now,
			VisibleTo:   nil,
			PublishedAt: &now,
			PublishedBy: "ops",
			CreatedBy:   "ops",
			UpdatedBy:   "ops",
			CreatedAt:   now.Add(-6 * time.Hour),
			UpdatedAt:   now,
		},
	}

	return tx.Create(&announcements).Error
}

func seedUserBalances(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.UserBalance{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var users []repository.User
	if err := tx.Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	now := time.Now().UTC()
	var transactions []repository.BalanceTransaction
	var balances []repository.UserBalance
	for _, user := range users {
		balanceCents := int64(0)
		if strings.Contains(strings.ToLower(user.Email), "admin") {
			balanceCents = 0
		} else {
			balanceCents = 12500
		}
		balance := repository.UserBalance{
			UserID:       user.ID,
			BalanceCents: balanceCents,
			Currency:     "CNY",
			CreatedAt:    now.Add(-48 * time.Hour),
			UpdatedAt:    now,
		}

		if balanceCents > 0 {
			txRecord := repository.BalanceTransaction{
				UserID:            user.ID,
				Type:              "recharge",
				AmountCents:       balanceCents,
				Currency:          "CNY",
				BalanceAfterCents: balanceCents,
				Reference:         "seed-order-001",
				Description:       "演示充值到账",
				Metadata: map[string]any{
					"channel": "demo",
				},
				CreatedAt: now.Add(-24 * time.Hour),
			}
			transactions = append(transactions, txRecord)
		}

		balances = append(balances, balance)
	}

	if len(balances) > 0 {
		if err := tx.Create(&balances).Error; err != nil {
			return err
		}
	}
	if len(transactions) > 0 {
		if err := tx.Create(&transactions).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedOrders(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.Order{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customer repository.User
	if err := tx.Where("email = ?", "user@example.com").First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	var plan repository.Plan
	if err := tx.Where("slug = ?", "standard").First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	now := time.Now().UTC()
	paidAt := now.Add(-2 * time.Hour)
	number := repository.GenerateOrderNumber()

	order := repository.Order{
		Number:        number,
		UserID:        customer.ID,
		PlanID:        &plan.ID,
		Status:        repository.OrderStatusPaid,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    plan.PriceCents,
		Currency:      plan.Currency,
		RefundedCents: 0,
		Metadata: map[string]any{
			"quantity": 1,
			"seed":     true,
		},
		PlanSnapshot: map[string]any{
			"id":                  plan.ID,
			"name":                plan.Name,
			"slug":                plan.Slug,
			"description":         plan.Description,
			"price_cents":         plan.PriceCents,
			"currency":            plan.Currency,
			"duration_days":       plan.DurationDays,
			"traffic_limit_bytes": plan.TrafficLimitBytes,
			"devices_limit":       plan.DevicesLimit,
			"features":            plan.Features,
			"tags":                plan.Tags,
		},
		PaidAt:    &paidAt,
		CreatedAt: now.Add(-3 * time.Hour),
		UpdatedAt: now.Add(-1 * time.Hour),
	}
	if err := tx.Create(&order).Error; err != nil {
		return err
	}

	item := repository.OrderItem{
		OrderID:        order.ID,
		ItemType:       "plan",
		ItemID:         plan.ID,
		Name:           plan.Name,
		Quantity:       1,
		UnitPriceCents: plan.PriceCents,
		Currency:       plan.Currency,
		SubtotalCents:  plan.PriceCents,
		Metadata: map[string]any{
			"duration_days":       plan.DurationDays,
			"traffic_limit_bytes": plan.TrafficLimitBytes,
			"devices_limit":       plan.DevicesLimit,
		},
		CreatedAt: now.Add(-3 * time.Hour),
	}
	if err := tx.Create(&item).Error; err != nil {
		return err
	}

	var balance repository.UserBalance
	if err := tx.Where("user_id = ?", customer.ID).First(&balance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		balance = repository.UserBalance{
			UserID:       customer.ID,
			BalanceCents: 0,
			Currency:     plan.Currency,
			CreatedAt:    now.Add(-48 * time.Hour),
			UpdatedAt:    now.Add(-3 * time.Hour),
		}
		if err := tx.Create(&balance).Error; err != nil {
			return err
		}
	}

	newBalance := balance.BalanceCents - plan.PriceCents
	purchaseTx := repository.BalanceTransaction{
		UserID:            customer.ID,
		Type:              "purchase",
		AmountCents:       -plan.PriceCents,
		Currency:          plan.Currency,
		BalanceAfterCents: newBalance,
		Reference:         fmt.Sprintf("order:%s", order.Number),
		Description:       fmt.Sprintf("演示购买套餐 %s", plan.Name),
		Metadata: map[string]any{
			"plan_id":      plan.ID,
			"order_number": order.Number,
		},
		CreatedAt: paidAt,
	}
	if err := tx.Create(&purchaseTx).Error; err != nil {
		return err
	}

	if err := tx.Model(&repository.UserBalance{}).
		Where("user_id = ?", customer.ID).
		Updates(map[string]any{
			"balance_cents":       newBalance,
			"currency":            plan.Currency,
			"last_transaction_id": purchaseTx.ID,
			"updated_at":          paidAt,
		}).Error; err != nil {
		return err
	}

	return nil
}

func seedSecuritySettings(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&repository.SecuritySetting{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	setting := repository.SecuritySetting{
		ThirdPartyAPIEnabled: false,
		EncryptionAlgorithm:  "aes-gcm",
		NonceTTLSeconds:      300,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	return tx.Create(&setting).Error
}
