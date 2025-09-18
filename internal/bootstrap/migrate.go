package bootstrap

import (
	"context"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/migrations"
)

// AutoMigrate is kept for backwards compatibility and applies the latest migrations.
func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	return ApplyMigrations(ctx, db, 0)
}

// ApplyMigrations executes schema migrations up to the target version (0 = latest).
func ApplyMigrations(ctx context.Context, db *gorm.DB, targetVersion uint64) error {
	return migrations.Apply(ctx, db, targetVersion)
}
