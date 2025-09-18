package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AdminModule 描述管理后台的功能模块。
type AdminModule struct {
	ID          uint64   `gorm:"primaryKey"`
	Key         string   `gorm:"size:64;uniqueIndex"`
	Name        string   `gorm:"size:128"`
	Description string   `gorm:"size:255"`
	Icon        string   `gorm:"size:64"`
	Route       string   `gorm:"size:255"`
	Permissions []string `gorm:"serializer:json"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName 覆盖默认的表名。
func (AdminModule) TableName() string { return "admin_modules" }

// AdminModuleRepository 定义管理后台模块仓储接口。
type AdminModuleRepository interface {
	ListModules(ctx context.Context) ([]AdminModule, error)
}

type adminModuleRepository struct {
	db *gorm.DB
}

// NewAdminModuleRepository 创建管理后台模块仓储实例。
func NewAdminModuleRepository(db *gorm.DB) (AdminModuleRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &adminModuleRepository{db: db}, nil
}

// ListModules 返回所有模块定义。
func (r *adminModuleRepository) ListModules(ctx context.Context) ([]AdminModule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var modules []AdminModule
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&modules).Error; err != nil {
		return nil, translateError(err)
	}

	return modules, nil
}
