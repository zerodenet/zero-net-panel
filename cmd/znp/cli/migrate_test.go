package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/migrations"
	"github.com/zero-net-panel/zero-net-panel/internal/testutil"
)

func writeTestConfig(t *testing.T, dir, dsn string) string {
	t.Helper()

	content := fmt.Sprintf(`Name: znp.test
Host: 127.0.0.1
Port: 0
Timeout: 1000

Project:
  Name: Test
  Description: CLI migrate test
  Version: 0.0.0

Database:
  Driver: sqlite
  DSN: %q
  MaxOpenConns: 1
  MaxIdleConns: 1
  ConnMaxLifetime: 0s
  LogLevel: silent

Cache:
  Provider: memory
  Memory:
    Size: 32
  Redis:
    Host: ""
    Type: ""
    Password: ""
    TLS: false
    NonBlock: false
    PingTimeout: 0s

Kernel:
  DefaultProtocol: http
  HTTP:
    BaseURL: http://localhost
    Token: ""
    Timeout: 1s
  GRPC:
    Endpoint: 127.0.0.1:9000
    TLSCert: ""
    Timeout: 1s

Auth:
  AccessSecret: test
  AccessExpire: 1h
  RefreshSecret: test-refresh
  RefreshExpire: 1h

Metrics:
  Enable: false
  Path: /metrics
  ListenOn: ""

Admin:
  RoutePrefix: admin

GRPCServer:
  Enable: true
  ListenOn: 127.0.0.1:0
  Reflection: true
`, dsn)

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func openSQLiteTestDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	testutil.RequireSQLite(t)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestMigrateCommandApplyAndRollbackOutput(t *testing.T) {
	testutil.RequireSQLite(t)

	if len(migrationsList(t)) < 1 {
		t.Skip("no migrations registered")
	}

	dir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1", filepath.Join(dir, "cli-test.db"))
	configPath := writeTestConfig(t, dir, dsn)
	opts := &GlobalOptions{ConfigFile: configPath}

	applyCmd := NewMigrateCommand(opts)
	applyCmd.SetContext(context.Background())
	applyCmd.SilenceErrors = true
	applyCmd.SilenceUsage = true
	applyBuf := new(bytes.Buffer)
	applyCmd.SetOut(applyBuf)
	applyCmd.SetErr(applyBuf)

	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("execute apply: %v", err)
	}

	applyOutput := applyBuf.String()
	if !strings.Contains(applyOutput, "Migrations applied") {
		t.Fatalf("expected apply output to mention migrations applied, got %q", applyOutput)
	}
	if !strings.Contains(applyOutput, "versions=[") {
		t.Fatalf("expected apply output to include versions, got %q", applyOutput)
	}

	db := openSQLiteTestDB(t, dsn)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	var versions []uint64
	if err := db.WithContext(context.Background()).Model(&migrations.SchemaMigration{}).Order("version asc").Pluck("version", &versions).Error; err != nil {
		t.Fatalf("load versions: %v", err)
	}
	if len(versions) == 0 {
		t.Fatalf("expected at least one migration to be applied")
	}

	originalFirst := versions[0]
	originalLast := versions[len(versions)-1]

	rollbackTarget := versions[0]
	if len(versions) > 1 {
		rollbackTarget = versions[len(versions)-2]
	} else if rollbackTarget > 0 {
		rollbackTarget--
	} else {
		t.Skip("cannot determine rollback target")
	}

	rollbackCmd := NewMigrateCommand(opts)
	rollbackCmd.SetContext(context.Background())
	rollbackCmd.SilenceErrors = true
	rollbackCmd.SilenceUsage = true
	rollbackBuf := new(bytes.Buffer)
	rollbackCmd.SetOut(rollbackBuf)
	rollbackCmd.SetErr(rollbackBuf)
	rollbackCmd.SetArgs([]string{"--apply", "--rollback", "--to", fmt.Sprintf("%d", rollbackTarget)})

	if err := rollbackCmd.Execute(); err != nil {
		t.Fatalf("execute rollback: %v", err)
	}

	rollbackOutput := rollbackBuf.String()
	if !strings.Contains(rollbackOutput, "Rollback complete") {
		t.Fatalf("expected rollback output to mention completion, got %q", rollbackOutput)
	}
	if !strings.Contains(rollbackOutput, "versions=[") {
		t.Fatalf("expected rollback output to list versions, got %q", rollbackOutput)
	}

	versions = versions[:0]
	if err := db.WithContext(context.Background()).Model(&migrations.SchemaMigration{}).Order("version asc").Pluck("version", &versions).Error; err != nil {
		t.Fatalf("load versions after rollback: %v", err)
	}
	if len(versions) == 0 {
		if rollbackTarget >= originalFirst {
			t.Fatalf("expected at least one migration to remain after rollback, target=%d, first=%d", rollbackTarget, originalFirst)
		}
		return
	}

	last := versions[len(versions)-1]
	if rollbackTarget < originalFirst {
		if last != 0 {
			t.Fatalf("expected schema to reset, got last version %d", last)
		}
	} else {
		if last > rollbackTarget {
			t.Fatalf("expected highest version <= %d, got %d", rollbackTarget, last)
		}
		if originalLast > rollbackTarget && last != rollbackTarget {
			t.Fatalf("expected last version %d, got %d", rollbackTarget, last)
		}
	}
}

func TestMigrateCommandTargetTooNew(t *testing.T) {
	testutil.RequireSQLite(t)

	if len(migrationsList(t)) == 0 {
		t.Skip("no migrations registered")
	}

	dir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1", filepath.Join(dir, "cli-too-new.db"))
	configPath := writeTestConfig(t, dir, dsn)
	opts := &GlobalOptions{ConfigFile: configPath}

	versions := migrationsList(t)
	latest := versions[len(versions)-1]
	if latest == ^uint64(0) {
		t.Skip("cannot exceed max uint64")
	}

	cmd := NewMigrateCommand(opts)
	cmd.SetContext(context.Background())
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--apply", "--to", fmt.Sprintf("%d", latest+1)})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when targeting version beyond latest")
	}
}

func migrationsList(t *testing.T) []uint64 {
	t.Helper()

	db := openSQLiteTestDB(t, fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=1", t.Name()))
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	result, err := migrations.Apply(context.Background(), db, 0, false)
	if err != nil {
		t.Fatalf("apply migrations for helper: %v", err)
	}
	return result.AppliedVersions
}
