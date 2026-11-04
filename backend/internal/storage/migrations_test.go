package storage

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestSentenceMigrationRollsBackOnLateFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollback.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE quiz_attempts (
		id TEXT PRIMARY KEY NOT NULL, mode TEXT NOT NULL CHECK (mode IN ('meaning', 'reading')),
		kind TEXT NOT NULL CHECK (kind IN ('full', 'retry')), parent_attempt_id TEXT,
		started_at TEXT NOT NULL, completed_at TEXT NOT NULL, completed_at_ms INTEGER NOT NULL,
		correct_count INTEGER NOT NULL, total INTEGER NOT NULL CHECK (total BETWEEN 1 AND 10),
		result_json TEXT NOT NULL, CHECK (correct_count BETWEEN 0 AND total));
		CREATE TABLE other_table (value TEXT);
		CREATE INDEX quiz_attempts_mode_time ON other_table(value);
		INSERT INTO quiz_attempts VALUES ('attempt-1', 'meaning', 'retry', NULL,
			'2026-11-01T10:00:00Z', '2026-11-01T10:01:00Z', 1, 1, 1, '{"items":[]}');`)
	if err != nil {
		t.Fatal(err)
	}
	// The occupied index name fails after the copy/drop/rename statements. The
	// original schema and exact snapshot must both survive transaction rollback.
	if opened, err := Open(path); err == nil {
		opened.Close()
		t.Fatal("expected occupied index name to reject the migration")
	}
	var schema, snapshot string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE name = 'quiz_attempts'`).Scan(&schema); err != nil || strings.Contains(schema, "'sentence'") {
		t.Fatal("old schema was not restored")
	}
	if err := db.QueryRow(`SELECT result_json FROM quiz_attempts WHERE id = 'attempt-1'`).Scan(&snapshot); err != nil || snapshot != `{"items":[]}` {
		t.Fatal("rollback lost or rewrote the existing attempt")
	}
}
