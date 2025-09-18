package repository

import (
	"errors"

	"gorm.io/gorm"
)

// Repositories 聚合各领域仓储，方便在 ServiceContext 中注入。
type Repositories struct {
	AdminModule          AdminModuleRepository
	Node                 NodeRepository
	SubscriptionTemplate SubscriptionTemplateRepository
	Subscription         SubscriptionRepository
	User                 UserRepository
	Plan                 PlanRepository
	Announcement         AnnouncementRepository
	Balance              BalanceRepository
	Security             SecurityRepository
	Order                OrderRepository
}

// NewRepositories 根据数据库实例创建仓储集合。
func NewRepositories(db *gorm.DB) (*Repositories, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}

	adminModuleRepo, err := NewAdminModuleRepository(db)
	if err != nil {
		return nil, err
	}

	templateRepo, err := NewSubscriptionTemplateRepository(db)
	if err != nil {
		return nil, err
	}

	userRepo, err := NewUserRepository(db)
	if err != nil {
		return nil, err
	}

	nodeRepo, err := NewNodeRepository(db)
	if err != nil {
		return nil, err
	}

	subscriptionRepo, err := NewSubscriptionRepository(db, templateRepo)
	if err != nil {
		return nil, err
	}

	planRepo, err := NewPlanRepository(db)
	if err != nil {
		return nil, err
	}

	announcementRepo, err := NewAnnouncementRepository(db)
	if err != nil {
		return nil, err
	}

	balanceRepo, err := NewBalanceRepository(db)
	if err != nil {
		return nil, err
	}

	securityRepo, err := NewSecurityRepository(db)
	if err != nil {
		return nil, err
	}

	orderRepo, err := NewOrderRepository(db)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		AdminModule:          adminModuleRepo,
		Node:                 nodeRepo,
		SubscriptionTemplate: templateRepo,
		Subscription:         subscriptionRepo,
		User:                 userRepo,
		Plan:                 planRepo,
		Announcement:         announcementRepo,
		Balance:              balanceRepo,
		Security:             securityRepo,
		Order:                orderRepo,
	}, nil
}
