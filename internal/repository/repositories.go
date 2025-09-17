package repository

import "gorm.io/gorm"

// Repositories 聚合各领域仓储，方便在 ServiceContext 中注入。
type Repositories struct {
	Node                 NodeRepository
	SubscriptionTemplate SubscriptionTemplateRepository
	Subscription         SubscriptionRepository
	User                 UserRepository
}

// NewRepositories 根据数据库实例创建仓储集合。
func NewRepositories(db *gorm.DB) *Repositories {
	templateRepo := NewSubscriptionTemplateRepository(db)
	userRepo := NewUserRepository(db)
	return &Repositories{
		Node:                 NewNodeRepository(db),
		SubscriptionTemplate: templateRepo,
		Subscription:         NewSubscriptionRepository(db, templateRepo),
		User:                 userRepo,
	}
}
