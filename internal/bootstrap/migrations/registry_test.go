package migrations

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testModelA struct {
	ID uint64 `gorm:"primaryKey"`
}

func (testModelA) TableName() string { return "test_model_a" }

type testModelB struct {
	ID uint64 `gorm:"primaryKey"`
}

func (testModelB) TableName() string { return "test_model_b" }

func openSQLite(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

func countMigrations(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&SchemaMigration{}).Count(&count).Error; err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	return count
}

func TestApplyMigrationsIdempotent(t *testing.T) {
	if len(migrationRegistry) == 0 {
		t.Skip("no migrations registered")
	}

	db := openSQLite(t)
	ctx := context.Background()

	first, err := Apply(ctx, db, 0, false)
	if err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if first.BeforeVersion != 0 {
		t.Fatalf("expected initial version 0, got %d", first.BeforeVersion)
	}
	if first.AfterVersion != first.TargetVersion {
		t.Fatalf("expected after version equals target, got after=%d target=%d", first.AfterVersion, first.TargetVersion)
	}

	second, err := Apply(ctx, db, 0, false)
	if err != nil {
		t.Fatalf("apply second time: %v", err)
	}
	if len(second.AppliedVersions) != 0 {
		t.Fatalf("expected no new migrations on second run, got %v", second.AppliedVersions)
	}
	if second.AfterVersion != first.AfterVersion {
		t.Fatalf("expected version unchanged, got %d", second.AfterVersion)
	}

	expectedCount := int64(0)
	target := first.TargetVersion
	for _, m := range migrationRegistry {
		if m.Version <= target {
			expectedCount++
		}
	}
	if count := countMigrations(t, db); count != expectedCount {
		t.Fatalf("expected %d migration records, got %d", expectedCount, count)
	}
}

func TestApplyMigrationsRollbackRequiresPermission(t *testing.T) {
	if len(migrationRegistry) < 2 {
		t.Skip("need at least two migrations to test rollback")
	}

	db := openSQLite(t)
	ctx := context.Background()

	if _, err := Apply(ctx, db, 0, false); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	target := migrationRegistry[len(migrationRegistry)-2].Version
	if _, err := Apply(ctx, db, target, false); err == nil {
		t.Fatalf("expected rollback rejection when allowRollback=false")
	}
	if count := countMigrations(t, db); count == 0 {
		t.Fatalf("expected migration metadata to remain after failed rollback")
	}
}

func TestApplyMigrationsRollbackOutOfRange(t *testing.T) {
	if len(migrationRegistry) == 0 {
		t.Skip("no migrations registered")
	}

	db := openSQLite(t)
	ctx := context.Background()

	if _, err := Apply(ctx, db, 0, false); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	firstVersion := migrationRegistry[0].Version
	if firstVersion == 0 {
		t.Skip("cannot derive out-of-range target when first version is 0")
	}
	target := firstVersion - 1

	result, err := Apply(ctx, db, target, true)
	if err != nil {
		t.Fatalf("rollback to %d: %v", target, err)
	}
	if result.AfterVersion != 0 {
		t.Fatalf("expected no migrations remaining, got %d", result.AfterVersion)
	}
	if count := countMigrations(t, db); count != 0 {
		t.Fatalf("expected metadata cleared after rollback, got %d", count)
	}
}

func TestApplyMigrationsUpFailureRecovery(t *testing.T) {
	db := openSQLite(t)
	ctx := context.Background()

	original := migrationRegistry
	t.Cleanup(func() { migrationRegistry = original })

	migrationRegistry = []Migration{
		{
			Version: 2100010101,
			Name:    "failing-up",
			Up: func(ctx context.Context, db *gorm.DB) error {
				if err := db.WithContext(ctx).AutoMigrate(&testModelA{}); err != nil {
					return err
				}
				return errors.New("boom")
			},
			Down: func(ctx context.Context, db *gorm.DB) error {
				return db.WithContext(ctx).Migrator().DropTable(&testModelA{})
			},
		},
	}

	if _, err := Apply(ctx, db, 0, false); err == nil {
		t.Fatalf("expected migration to fail")
	}
	if db.Migrator().HasTable(&testModelA{}) {
		t.Fatalf("table should not exist after failed migration")
	}
	if count := countMigrations(t, db); count != 0 {
		t.Fatalf("expected metadata rollback, got %d records", count)
	}
}

func TestApplyMigrationsRollbackFailureRecovery(t *testing.T) {
	db := openSQLite(t)
	ctx := context.Background()

	original := migrationRegistry
	t.Cleanup(func() { migrationRegistry = original })

	migrationRegistry = []Migration{
		{
			Version: 2100010101,
			Name:    "base",
			Up: func(ctx context.Context, db *gorm.DB) error {
				return db.WithContext(ctx).AutoMigrate(&testModelA{})
			},
			Down: func(ctx context.Context, db *gorm.DB) error {
				return db.WithContext(ctx).Migrator().DropTable(&testModelA{})
			},
		},
		{
			Version: 2100010102,
			Name:    "failing-down",
			Up: func(ctx context.Context, db *gorm.DB) error {
				return db.WithContext(ctx).AutoMigrate(&testModelB{})
			},
			Down: func(ctx context.Context, db *gorm.DB) error {
				if err := db.WithContext(ctx).Migrator().DropTable(&testModelB{}); err != nil {
					return err
				}
				return errors.New("down failure")
			},
		},
	}

	if _, err := Apply(ctx, db, 0, false); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	if _, err := Apply(ctx, db, migrationRegistry[0].Version, true); err == nil {
		t.Fatalf("expected rollback to fail")
	}

	if !db.Migrator().HasTable(&testModelB{}) {
		t.Fatalf("expected table to remain after failed rollback")
	}
	if count := countMigrations(t, db); count != 2 {
		t.Fatalf("expected both metadata records to remain, got %d", count)
	}
}
