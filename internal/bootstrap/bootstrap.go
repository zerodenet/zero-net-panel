package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/migrations"
	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/seed"
	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

// DatabaseOptions controls bootstrap behaviour for database tasks.
type DatabaseOptions struct {
	AutoMigrate   bool
	SeedDemo      bool
	TargetVersion uint64
	AllowRollback bool
}

// PrepareDatabase ensures that migrations and seed data are applied before the
// service starts. It opens a dedicated connection, applies the requested
// operations and closes the connection afterwards.
func PrepareDatabase(ctx context.Context, cfg config.Config, opts DatabaseOptions) (migrations.Result, error) {
	var result migrations.Result

	db, closeFn, err := database.NewGorm(cfg.Database)
	if err != nil {
		return result, fmt.Errorf("connect database: %w", err)
	}
	if db == nil {
		closeFn()
		return result, errors.New("database configuration is required")
	}
	defer closeFn()

	if opts.AutoMigrate {
		var applyErr error
		result, applyErr = ApplyMigrations(ctx, db, opts.TargetVersion, opts.AllowRollback)
		if applyErr != nil {
			return result, applyErr
		}
	}

	if opts.SeedDemo {
		if err := seed.Run(ctx, db); err != nil {
			return result, err
		}
	}

	return result, nil
}

// WithTransaction executes fn within a transaction respecting the context.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
