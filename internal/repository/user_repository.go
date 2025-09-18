package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// User 描述系统用户信息。
type User struct {
	ID           uint64   `gorm:"primaryKey"`
	Email        string   `gorm:"size:255;uniqueIndex"`
	DisplayName  string   `gorm:"size:255"`
	PasswordHash string   `gorm:"size:255"`
	Roles        []string `gorm:"serializer:json"`
	Status       string   `gorm:"size:32"`
	LastLoginAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName 自定义用户表名。
func (User) TableName() string { return "users" }

// UserRepository 定义用户仓储接口。
type UserRepository interface {
	Get(ctx context.Context, id uint64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateLastLogin(ctx context.Context, id uint64, ts time.Time) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储，当前以内存实现模拟。
func NewUserRepository(db *gorm.DB) (UserRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &userRepository{db: db}, nil
}

func (r *userRepository) Get(ctx context.Context, id uint64) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	var user User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return User{}, translateError(err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	var user User
	if err := r.db.WithContext(ctx).
		Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error; err != nil {
		return User{}, translateError(err)
	}

	return user, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint64, ts time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	updates := map[string]any{
		"last_login_at": ts,
		"updated_at":    ts,
	}

	if err := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return translateError(err)
	}

	return nil
}
