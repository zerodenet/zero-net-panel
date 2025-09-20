package testutil

import (
	"database/sql"
	"go/build"
	"runtime"
	"testing"
)

const sqliteDriver = "sqlite3"

// RequireSQLite skips the current test when SQLite support is unavailable.
func RequireSQLite(t *testing.T) {
	t.Helper()

	if !build.Default.CgoEnabled {
		t.Skipf("sqlite-dependent test skipped: CGO_ENABLED=0 on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	available := false
	for _, name := range sql.Drivers() {
		if name == sqliteDriver {
			available = true
			break
		}
	}
	if !available {
		t.Skipf("sqlite-dependent test skipped: driver %q not registered", sqliteDriver)
	}

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		t.Skipf("sqlite-dependent test skipped: open %q driver failed: %v", sqliteDriver, err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.Ping(); err != nil {
		t.Skipf("sqlite-dependent test skipped: ping %q driver failed: %v", sqliteDriver, err)
	}
}
