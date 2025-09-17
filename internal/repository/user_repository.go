package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// User 描述系统用户信息。
type User struct {
	ID           uint64
	Email        string
	DisplayName  string
	PasswordHash string
	Roles        []string
	Status       string
	LastLoginAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository 定义用户仓储接口。
type UserRepository interface {
	Get(ctx context.Context, id uint64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateLastLogin(ctx context.Context, id uint64, ts time.Time) error
}

type userRepository struct {
	db *gorm.DB

	mu         sync.RWMutex
	users      map[uint64]*User
	emailIndex map[string]uint64
	nextID     uint64
}

// NewUserRepository 创建用户仓储，当前以内存实现模拟。
func NewUserRepository(db *gorm.DB) UserRepository {
	repo := &userRepository{
		db:         db,
		users:      make(map[uint64]*User),
		emailIndex: make(map[string]uint64),
		nextID:     1,
	}
	repo.seed()
	return repo
}

func (r *userRepository) seed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()

	admin := &User{
		Email:        "admin@example.com",
		DisplayName:  "运营管理员",
		PasswordHash: "$2a$10$OmiVLT.Awz75.D1g1Rvm7.TPPaB399VUCpQCJiBCnWGEN2L4IyJTe",
		Roles:        []string{"admin", "user"},
		Status:       "active",
		CreatedAt:    now.Add(-72 * time.Hour),
		UpdatedAt:    now.Add(-24 * time.Hour),
		LastLoginAt:  now.Add(-48 * time.Hour),
	}
	adminID := r.addUserLocked(admin)
	r.emailIndex[strings.ToLower(admin.Email)] = adminID

	member := &User{
		Email:        "user@example.com",
		DisplayName:  "高级会员",
		PasswordHash: "$2a$10$OmiVLT.Awz75.D1g1Rvm7.TPPaB399VUCpQCJiBCnWGEN2L4IyJTe",
		Roles:        []string{"user"},
		Status:       "active",
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now.Add(-12 * time.Hour),
		LastLoginAt:  now.Add(-6 * time.Hour),
	}
	memberID := r.addUserLocked(member)
	r.emailIndex[strings.ToLower(member.Email)] = memberID
}

func (r *userRepository) Get(ctx context.Context, id uint64) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}

	return cloneUser(user), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.emailIndex[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, ErrNotFound
	}

	user, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}

	return cloneUser(user), nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint64, ts time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return ErrNotFound
	}

	user.LastLoginAt = ts
	user.UpdatedAt = ts
	return nil
}

func (r *userRepository) addUserLocked(user *User) uint64 {
	id := r.nextID
	r.nextID++

	copied := cloneUser(user)
	copied.ID = id
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = time.Now().UTC()
	}
	if copied.UpdatedAt.IsZero() {
		copied.UpdatedAt = copied.CreatedAt
	}

	r.users[id] = &copied
	return id
}

func cloneUser(user *User) User {
	copied := *user
	if user.Roles != nil {
		copied.Roles = append([]string(nil), user.Roles...)
	}
	return copied
}
