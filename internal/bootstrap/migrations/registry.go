package migrations

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

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
func Apply(ctx context.Context, db *gorm.DB, targetVersion uint64, _ bool) error {
	if db == nil {
		return result, fmt.Errorf("migrations: database connection is required")
	}

	if err := db.WithContext(ctx).AutoMigrate(&SchemaMigration{}); err != nil {
		return result, fmt.Errorf("migrations: prepare metadata table: %w", err)
	}

	var applied []SchemaMigration
	if err := db.WithContext(ctx).Order("version ASC").Find(&applied).Error; err != nil {
		return result, fmt.Errorf("migrations: load applied versions: %w", err)
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

	for _, m := range migrationRegistry {
		registryMap[m.Version] = m
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

	if effectiveTarget < currentVersion && !allowRollback {
		result.AfterVersion = currentVersion
		return result, fmt.Errorf("migrations: target version %d is older than current version %d; enable rollback (e.g. --rollback) to continue", effectiveTarget, currentVersion)
	}

	if effectiveTarget > currentVersion {
		for _, m := range migrationRegistry {
			if m.Version > effectiveTarget {
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
				result.AfterVersion = currentVersion
				return result, fmt.Errorf("migrations: apply %d (%s): %w", m.Version, m.Name, err)
			}

			result.AppliedVersions = append(result.AppliedVersions, m.Version)
			appliedSet[m.Version] = SchemaMigration{Version: m.Version}
		}
	} else if effectiveTarget < currentVersion {
		for i := len(applied) - 1; i >= 0; i-- {
			record := applied[i]
			if record.Version <= effectiveTarget {
				break
			}

			migration := registryMap[record.Version]
			if migration.Down == nil {
				result.AfterVersion = currentVersion
				return result, fmt.Errorf("migrations: migration %d (%s) does not support rollback", record.Version, record.Name)
			}

			if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := migration.Down(ctx, tx); err != nil {
					return err
				}
				if err := tx.Where("version = ?", migration.Version).Delete(&SchemaMigration{}).Error; err != nil {
					return err
				}
				return nil
			}); err != nil {
				result.AfterVersion = currentVersion
				return result, fmt.Errorf("migrations: rollback %d (%s): %w", migration.Version, migration.Name, err)
			}

			result.RolledBackVersions = append(result.RolledBackVersions, migration.Version)
		}
	}

	var latest SchemaMigration
	err := db.WithContext(ctx).Order("version DESC").Take(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		result.AfterVersion = 0
	case err != nil:
		return result, fmt.Errorf("migrations: determine current version: %w", err)
	default:
		result.AfterVersion = latest.Version
	}

	return result, nil
}
