package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"
)

func testDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDatabase(filepath.Join(t.TempDir(), "data", "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestStoredKanjiRoundTripAndUpsert(t *testing.T) {
	db := testDatabase(t)
	ctx := context.Background()
	if _, found, err := getStoredKanji(ctx, db, "書"); err != nil || found {
		t.Fatalf("expected a cache miss, got found=%v err=%v", found, err)
	}
	want := kanjiDetails{
		Kanji: "書", Meanings: []string{"write", `a "quoted" meaning`},
		OnReadings: []string{"ショ"}, KunReadings: []string{"か.く"}, NameReadings: []string{}, StrokeCount: 10,
	}
	for i := 0; i < 2; i++ {
		if i == 1 {
			want.Meanings = []string{"updated meaning"}
		}
		if err := saveKanji(ctx, db, want, "https://kanjiapi.dev/v1/kanji/書"); err != nil {
			t.Fatal(err)
		}
		got, found, err := getStoredKanji(ctx, db, "書")
		if err != nil || !found || !reflect.DeepEqual(got, want) {
			t.Fatalf("round trip: got %+v found=%v err=%v, want %+v", got, found, err, want)
		}
	}
	var rows int
	if err := db.QueryRow("SELECT COUNT(*) FROM kanji_entries").Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("got %d rows after upsert, want 1", rows)
	}
}

func TestStoredKanjiRejectsCorruptJSON(t *testing.T) {
	db := testDatabase(t)
	ctx := context.Background()
	if err := saveKanji(ctx, db, kanjiDetails{Kanji: "書", StrokeCount: 10}, "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE kanji_entries SET meanings = ? WHERE kanji = ?", "invalid JSON", "書"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := getStoredKanji(ctx, db, "書"); err == nil {
		t.Fatal("corrupt stored JSON must return an error rather than valid dictionary data")
	}
}
