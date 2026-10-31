package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKanjiStatsReadOneConsistentSnapshot(t *testing.T) {
	db := testDatabase(t)
	if _, err := db.Exec(`INSERT INTO anki_cards
		(card_id, note_id, deck_name, model_name, vocabulary, reading, answer_text, fields_json, imported_at)
		VALUES (1, 1, 'JP Vocabulary', 'Basic', '書く', 'かく', '', '{}', 'test');
		INSERT INTO review_events(vocabulary, rating, reviewed_at_ms, source, raw_json)
		VALUES ('書く', 'hard', 1, 'anki', '{}'), ('書く', 'hard', 2, 'anki', '{}'), ('書く', 'hard', 3, 'anki', '{}')`); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	writerErrors := make(chan error, 1)
	go func() {
		defer close(done)
		for version := 0; ctx.Err() == nil; version++ {
			rating, reading := "hard", "かく"
			if version%2 == 1 {
				rating, reading = "good", "しょ"
			}
			tx, err := db.BeginTx(ctx, nil)
			if err == nil {
				_, err = tx.ExecContext(ctx, "UPDATE review_events SET rating = ?", rating)
				if err == nil {
					_, err = tx.ExecContext(ctx, "UPDATE anki_cards SET reading = ?", reading)
				}
				if err == nil {
					err = tx.Commit()
				} else {
					tx.Rollback()
				}
			}
			if err != nil {
				if ctx.Err() == nil {
					writerErrors <- err
				}
				return
			}
		}
	}()
	defer func() {
		cancel()
		<-done
		select {
		case err := <-writerErrors:
			t.Errorf("snapshot writer failed: %v", err)
		default:
		}
	}()

	store := reviewStore{db: db}
	for attempt := 0; attempt < 100; attempt++ {
		response := httptest.NewRecorder()
		store.listKanjiStats(response, httptest.NewRequest(http.MethodGet, "/kanji-stats", nil))
		var result []kanjiStats
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result) != 1 || result[0].Difficulty == nil || len(result[0].Evidence) != 1 {
			t.Fatalf("unexpected kanji response: %+v", result)
		}
		score, reading := *result[0].Difficulty, result[0].Evidence[0].Reading
		if !((score == 50 && reading == "かく") || (score == 0 && reading == "しょ")) {
			t.Fatalf("mixed database snapshots: difficulty=%d reading=%q", score, reading)
		}
	}
}
