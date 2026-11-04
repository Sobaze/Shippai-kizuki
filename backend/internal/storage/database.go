// Package storage opens the application's SQLite database and manages its schema.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open creates the schema if needed and applies migrations to existing data.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	// database/sql is safe for concurrent use. One connection serializes SQLite
	// operations for this small application; network lookups never hold it open.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS kanji_entries (
			kanji TEXT PRIMARY KEY NOT NULL,
			meanings TEXT NOT NULL,
			on_readings TEXT NOT NULL,
			kun_readings TEXT NOT NULL,
			name_readings TEXT NOT NULL,
			stroke_count INTEGER NOT NULL CHECK (stroke_count > 0),
			source TEXT NOT NULL,
			fetched_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS anki_cards (
			card_id INTEGER PRIMARY KEY,
			note_id INTEGER NOT NULL,
			deck_name TEXT NOT NULL,
			model_name TEXT NOT NULL,
			vocabulary TEXT NOT NULL,
			reading TEXT NOT NULL,
			answer_text TEXT NOT NULL,
			fields_json TEXT NOT NULL,
			imported_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS review_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			anki_review_id INTEGER UNIQUE,
			anki_card_id INTEGER,
			vocabulary TEXT NOT NULL,
			rating TEXT,
			reviewed_at_ms INTEGER NOT NULL,
			source TEXT NOT NULL,
			raw_json TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS review_events_time
			ON review_events(reviewed_at_ms DESC, id DESC);
		CREATE TABLE IF NOT EXISTS quiz_attempts (
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
		CREATE INDEX IF NOT EXISTS quiz_attempts_mode_time
			ON quiz_attempts(mode, completed_at_ms DESC, id DESC);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	if err := migrateSentenceAttempts(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("upgrade quiz attempts: %w", err)
	}
	return db, nil
}
