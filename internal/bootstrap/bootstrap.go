package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/seed"
	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

// DatabaseOptions controls bootstrap behaviour for database tasks.
type DatabaseOptions struct {
	AutoMigrate   bool
	SeedDemo      bool
	TargetVersion uint64
}

// PrepareDatabase ensures that migrations and seed data are applied before the
// service starts. It opens a dedicated connection, applies the requested
// operations and closes the connection afterwards.
func PrepareDatabase(ctx context.Context, cfg config.Config, opts DatabaseOptions) error {
	db, closeFn, err := database.NewGorm(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	if db == nil {
		closeFn()
		return errors.New("database configuration is required")
	}
	defer closeFn()

	if opts.AutoMigrate {
		if err := ApplyMigrations(ctx, db, opts.TargetVersion); err != nil {
			return err
		}
	}

	if opts.SeedDemo {
		if err := seed.Run(ctx, db); err != nil {
			return err
		}
	}

	return nil
}

// WithTransaction executes fn within a transaction respecting the context.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
