package storage

import (
	"database/sql"
	"strings"
)

// SQLite cannot alter CHECK constraints. Copy existing rows unchanged in one
// transaction before switching table names; failures roll back the entire copy.
func migrateSentenceAttempts(db *sql.DB) error {
	var schema string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'quiz_attempts'`).Scan(&schema); err != nil {
		return err
	}
	if strings.Contains(schema, "'sentence'") {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		CREATE TABLE quiz_attempts_sentence_migration (
			id TEXT PRIMARY KEY NOT NULL,
			mode TEXT NOT NULL CHECK (mode IN ('meaning', 'reading', 'sentence')),
			kind TEXT NOT NULL CHECK (kind IN ('full', 'retry')),
			parent_attempt_id TEXT,
			started_at TEXT NOT NULL,
			completed_at TEXT NOT NULL,
			completed_at_ms INTEGER NOT NULL,
			correct_count INTEGER NOT NULL,
			total INTEGER NOT NULL CHECK (total BETWEEN 1 AND 12),
			result_json TEXT NOT NULL,
			CHECK (correct_count BETWEEN 0 AND total)
		);
		INSERT INTO quiz_attempts_sentence_migration
			(id, mode, kind, parent_attempt_id, started_at, completed_at, completed_at_ms, correct_count, total, result_json)
			SELECT id, mode, kind, parent_attempt_id, started_at, completed_at, completed_at_ms, correct_count, total, result_json FROM quiz_attempts;
		DROP TABLE quiz_attempts;
		ALTER TABLE quiz_attempts_sentence_migration RENAME TO quiz_attempts;
		CREATE INDEX quiz_attempts_mode_time ON quiz_attempts(mode, completed_at_ms DESC, id DESC);
	`)
	if err != nil {
		return err
	}
	return tx.Commit()
}
