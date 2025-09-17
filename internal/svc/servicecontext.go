package svc

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/pkg/auth"
	"github.com/zero-net-panel/zero-net-panel/pkg/cache"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
	"github.com/zero-net-panel/zero-net-panel/pkg/kernel"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	Cache        cache.Cache
	Repositories *repository.Repositories
	Kernel       *kernel.Registry
	Auth         *auth.Generator

	Ctx    context.Context
	cancel context.CancelFunc

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

	opts := kernel.Options{
		DefaultProtocol: c.Kernel.DefaultProtocol,
		HTTP: kernel.HTTPOptions{
			BaseURL: c.Kernel.HTTP.BaseURL,
			Token:   c.Kernel.HTTP.Token,
			Timeout: c.Kernel.HTTP.Timeout,
		},
		GRPC: kernel.GRPCOptions{
			Endpoint: c.Kernel.GRPC.Endpoint,
			TLSCert:  c.Kernel.GRPC.TLSCert,
			Timeout:  c.Kernel.GRPC.Timeout,
		},
	}

	kernelRegistry, err := kernel.NewRegistry(opts)
	if err != nil {
		_ = cacheProvider.Close()
		dbClose()
		return nil, fmt.Errorf("init kernel registry: %w", err)
	}

	repos := repository.NewRepositories(db)

	ctx, cancel := context.WithCancel(context.Background())

	authGenerator := auth.NewGenerator(
		c.Auth.AccessSecret,
		c.Auth.RefreshSecret,
		c.Auth.AccessExpire,
		c.Auth.RefreshExpire,
	)

	svcCtx := &ServiceContext{
		Config:       c,
		DB:           db,
		Cache:        cacheProvider,
		Repositories: repos,
		Kernel:       kernelRegistry,
		Auth:         authGenerator,
		Ctx:          ctx,
		cancel:       cancel,
	}

	svcCtx.cleanup = func() {
		if svcCtx.cancel != nil {
			svcCtx.cancel()
		}
		if kernelRegistry != nil {
			_ = kernelRegistry.Close()
		}
		if cacheProvider != nil {
			_ = cacheProvider.Close()
		}
		dbClose()
	}

	return svcCtx, nil
}

func (s *ServiceContext) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ServiceContext) Context() context.Context {
	return s.Ctx
}

func (s *ServiceContext) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}
