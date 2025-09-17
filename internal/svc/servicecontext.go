package svc

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/pkg/cache"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  cache.Cache

	cleanup func()
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, dbClose, err := database.NewGorm(c.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	cacheProvider, err := cache.New(c.Cache)
	if err != nil {
		dbClose()
		return nil, fmt.Errorf("init cache: %w", err)
	}

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
		Cache:  cacheProvider,
	}

	ctx.cleanup = func() {
		if cacheProvider != nil {
			_ = cacheProvider.Close()
		}
		dbClose()
	}

	return ctx, nil
}

func (s *ServiceContext) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}
