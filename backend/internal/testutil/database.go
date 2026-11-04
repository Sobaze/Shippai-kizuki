// Package testutil provides shared fixtures for backend tests.
package testutil

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Sobaze/Shippai-kizuki/backend/internal/storage"
)

// Database opens an isolated SQLite database and closes it when the test finishes.
func Database(t testing.TB) *sql.DB {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "data", "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
