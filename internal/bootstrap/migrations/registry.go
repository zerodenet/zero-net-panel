package migrations

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/seed"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
)

// SchemaMigration stores executed migration metadata.
type SchemaMigration struct {
	Version   uint64    `gorm:"primaryKey"`
	Name      string    `gorm:"size:128"`
	AppliedAt time.Time `gorm:"column:applied_at"`
}

// TableName ensures deterministic naming.
func (SchemaMigration) TableName() string { return "schema_migrations" }

// Migration describes an idempotent schema evolution step.
type Migration struct {
	Version uint64
	Name    string
	Up      func(ctx context.Context, db *gorm.DB) error
}

var migrationRegistry = []Migration{
	{
		Version: 2024060101,
		Name:    "base-schema",
		Up: func(ctx context.Context, db *gorm.DB) error {
			return db.WithContext(ctx).AutoMigrate(
				&repository.AdminModule{},
				&repository.User{},
				&repository.Node{},
				&repository.NodeKernel{},
				&repository.SubscriptionTemplate{},
				&repository.SubscriptionTemplateHistory{},
				&repository.Subscription{},
				&repository.Plan{},
				&repository.Announcement{},
				&repository.UserBalance{},
				&repository.BalanceTransaction{},
				&repository.SecuritySetting{},
			)
		},
	},
	{
		Version: 2024063001,
		Name:    "billing-orders",
		Up: func(ctx context.Context, db *gorm.DB) error {
			return db.WithContext(ctx).AutoMigrate(
				&repository.Order{},
				&repository.OrderItem{},
			)
		},
	},
	{
		Version: 2024100101,
		Name:    "billing-order-refunds",
		Up: func(ctx context.Context, db *gorm.DB) error {
			return db.WithContext(ctx).AutoMigrate(
				&repository.Order{},
				&repository.OrderRefund{},
			)
		},
	},
}

func init() {
	sort.Slice(migrationRegistry, func(i, j int) bool {
		return migrationRegistry[i].Version < migrationRegistry[j].Version
	})
}

// Apply executes migrations up to targetVersion (0 denotes latest).
// When withSeed is true, demo seed data will be populated after the
// migrations have been successfully applied. Seeding currently requires
// running against the latest schema (target version 0).
func Apply(ctx context.Context, db *gorm.DB, targetVersion uint64, withSeed bool) error {
	if db == nil {
		return fmt.Errorf("migrations: database connection is required")
	}

	if err := db.WithContext(ctx).AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("migrations: prepare metadata table: %w", err)
	}

	var applied []SchemaMigration
	if err := db.WithContext(ctx).Order("version ASC").Find(&applied).Error; err != nil {
		return fmt.Errorf("migrations: load applied versions: %w", err)
	}

	appliedSet := make(map[uint64]SchemaMigration, len(applied))
	var currentVersion uint64
	for _, record := range applied {
		appliedSet[record.Version] = record
		if record.Version > currentVersion {
			currentVersion = record.Version
		}
	}

	if targetVersion != 0 && targetVersion < currentVersion {
		return fmt.Errorf("migrations: target version %d is older than current version %d", targetVersion, currentVersion)
	}

	for _, m := range migrationRegistry {
		if targetVersion != 0 && m.Version > targetVersion {
			break
		}
		if _, ok := appliedSet[m.Version]; ok {
			continue
		}

		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := m.Up(ctx, tx); err != nil {
				return err
			}
			record := SchemaMigration{
				Version:   m.Version,
				Name:      m.Name,
				AppliedAt: time.Now().UTC(),
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("migrations: apply %d (%s): %w", m.Version, m.Name, err)
		}
	}

	if withSeed {
		if targetVersion != 0 {
			return fmt.Errorf("migrations: seed demo data requires latest schema (target version 0), got %d", targetVersion)
		}

		if err := seed.Run(ctx, db); err != nil {
			return fmt.Errorf("migrations: seed demo data: %w", err)
		}
	}

	return nil
}
