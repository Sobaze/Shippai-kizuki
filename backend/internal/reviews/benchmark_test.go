package reviews

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

// Match the current collection's scale without reading or changing personal data.
func BenchmarkVocabularyStats(b *testing.B) {
	db, err := openDatabase(filepath.Join(b.TempDir(), "benchmark.sqlite"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { db.Close() })
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Rollback()
	statement, err := tx.Prepare(`INSERT INTO review_events
		(vocabulary, rating, reviewed_at_ms, source, raw_json) VALUES (?, ?, ?, 'anki', '{}')`)
	if err != nil {
		b.Fatal(err)
	}
	ratings := []string{"again", "hard", "good", "easy"}
	for i := 0; i < 39701; i++ {
		if _, err := statement.Exec(fmt.Sprintf("語%04d", i%1741), ratings[i%4], i+1); err != nil {
			b.Fatal(err)
		}
	}
	statement.Close()
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
	store := reviewStore{db: db}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, err := store.loadVocabularyStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if len(stats) != 1741 {
			b.Fatalf("got %d vocabulary items", len(stats))
		}
	}
}
