package migrations

import (
	"context"
	"errors"
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
	Down    func(ctx context.Context, db *gorm.DB) error
}

// Result captures the state transition produced by Apply.
type Result struct {
	BeforeVersion      uint64
	AfterVersion       uint64
	TargetVersion      uint64
	AppliedVersions    []uint64
	RolledBackVersions []uint64
}

// ApplyResult captures details about a migration attempt.
type ApplyResult struct {
	// PreviousVersion reflects the version recorded before applying
	// any pending migrations.
	PreviousVersion uint64
	// CurrentVersion reflects the schema version after migrations have
	// been executed.
	CurrentVersion uint64
	// Applied enumerates the migrations that were newly executed within
	// the current invocation.
	Applied []SchemaMigration
	// Seeded indicates whether demo seed data was populated.
	Seeded bool
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
		Down: func(ctx context.Context, db *gorm.DB) error {
			migrator := db.WithContext(ctx).Migrator()
			tables := []any{
				&repository.SecuritySetting{},
				&repository.BalanceTransaction{},
				&repository.UserBalance{},
				&repository.Announcement{},
				&repository.Plan{},
				&repository.Subscription{},
				&repository.SubscriptionTemplateHistory{},
				&repository.SubscriptionTemplate{},
				&repository.NodeKernel{},
				&repository.Node{},
				&repository.User{},
				&repository.AdminModule{},
			}

			for _, table := range tables {
				if err := migrator.DropTable(table); err != nil {
					return err
				}
			}

			return nil
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
		Down: func(ctx context.Context, db *gorm.DB) error {
			migrator := db.WithContext(ctx).Migrator()
			if err := migrator.DropTable(&repository.OrderItem{}); err != nil {
				return err
			}
			if err := migrator.DropTable(&repository.Order{}); err != nil {
				return err
			}
			return nil
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
	{
		Version: 2024071501,
		Name:    "order-refund-tracking",
		Up: func(ctx context.Context, db *gorm.DB) error {
			return db.WithContext(ctx).AutoMigrate(
				&repository.Order{},
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
func Apply(ctx context.Context, db *gorm.DB, targetVersion uint64, allowRollback bool) error {
	if db == nil {
		return nil, fmt.Errorf("migrations: database connection is required")
	}

	if err := db.WithContext(ctx).AutoMigrate(&SchemaMigration{}); err != nil {
		return nil, fmt.Errorf("migrations: prepare metadata table: %w", err)
	}

	var applied []SchemaMigration
	if result := db.WithContext(ctx).Order("version ASC").Find(&applied); result.Error != nil {
		return fmt.Errorf("migrations: load applied versions: %w", result.Error)
	}

	appliedSet := make(map[uint64]SchemaMigration, len(applied))
	registryMap := make(map[uint64]Migration, len(migrationRegistry))
	var currentVersion uint64
	for _, record := range applied {
		appliedSet[record.Version] = record
		if record.Version > currentVersion {
			currentVersion = record.Version
		}
	}

	if targetVersion != 0 && targetVersion < currentVersion {
		if !allowRollback {
			return fmt.Errorf("migrations: target version %d is older than current version %d", targetVersion, currentVersion)
		}
		return fmt.Errorf("migrations: rollback to version %d from %d is not implemented", targetVersion, currentVersion)
	}

	for version := range appliedSet {
		if _, ok := registryMap[version]; !ok {
			return result, fmt.Errorf("migrations: applied version %d is not registered", version)
		}
	}

	result.BeforeVersion = currentVersion

	effectiveTarget := targetVersion
	if targetVersion == 0 {
		if len(migrationRegistry) > 0 {
			effectiveTarget = migrationRegistry[len(migrationRegistry)-1].Version
		}
	}
	result.TargetVersion = effectiveTarget

  var entry SchemaMigration
  if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := m.Up(ctx, tx); err != nil {
      return err
    }
    appliedAt := time.Now().UTC()
    entry = SchemaMigration{
      Version:   m.Version,
      Name:      m.Name,
      AppliedAt: appliedAt,
    }
    if result := tx.Create(&entry); result.Error != nil {
      return result.Error
    } else if result.RowsAffected != 1 {
      return fmt.Errorf("migrations: record version %d affected %d rows", m.Version, result.RowsAffected)
    }
			return nil
		}); err != nil {
			return nil, fmt.Errorf("migrations: apply %d (%s): %w", m.Version, m.Name, err)
		}
	}

	if withSeed {
		if targetVersion != 0 {
			return nil, fmt.Errorf("migrations: seed demo data requires latest schema (target version 0), got %d", targetVersion)
		}

		if err := seed.Run(ctx, db); err != nil {
			return nil, fmt.Errorf("migrations: seed demo data: %w", err)
		}
		appliedSet[m.Version] = entry
		if entry.Version > currentVersion {
			currentVersion = entry.Version
		}
	}

	return result, nil
}
