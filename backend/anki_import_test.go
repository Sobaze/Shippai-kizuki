package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAnkiVocabularyImportAndRepeat(t *testing.T) {
	var fail, removeSecond atomic.Bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Action string `json:"action"`
			Params struct {
				Query string  `json:"query"`
				Cards []int64 `json:"cards"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		var result any
		switch request.Action {
		case "deckNames":
			result = []string{vocabularyDeck, "JP Grammar"}
		case "findCards":
			if request.Params.Query != `deck:"JP Vocabulary"` {
				t.Errorf("wrong deck query: %q", request.Params.Query)
			}
			result = []int64{1, 2}
			if removeSecond.Load() {
				result = []int64{1}
			}
		case "cardsInfo":
			if fail.Load() {
				w.WriteHeader(503)
				return
			}
			cards := []ankiCard{
				{CardID: 1, NoteID: 101, DeckName: vocabularyDeck, ModelName: "Basic", Fields: map[string]ankiField{
					"Framsida": {Value: "<b>書く</b>"}, "Baksida": {Value: "かく<br>write"},
				}},
				{CardID: 2, NoteID: 102, DeckName: vocabularyDeck, ModelName: "Basic", Fields: map[string]ankiField{
					"Framsida": {Value: "読書"}, "Baksida": {Value: "【読み】<br>どくしょ<br>【英語】<br>reading"},
				}},
			}
			if removeSecond.Load() {
				cards = cards[:1]
			}
			result = cards
		case "getReviewsOfCards":
			result = map[string][]ankiReview{
				"1": {{ID: 1001, Ease: 1, Type: 0}, {ID: 1002, Ease: 2, Type: 2}, {ID: 1003, Ease: 3, Type: 1}, {ID: 1004, Ease: 0, Type: 4}},
				"2": {{ID: 2001, Ease: 3, Type: 1}, {ID: 2002, Ease: 2, Type: 1}, {ID: 2003, Ease: 3, Type: 1}},
			}
		default:
			t.Errorf("unexpected or write action: %s", request.Action)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": result, "error": nil})
	}))
	defer upstream.Close()
	db := testDatabase(t)
	client := ankiClient{client: upstream.Client(), endpoint: upstream.URL}
	importer := ankiImporter{anki: &client, db: db}
	if _, err := db.Exec(`INSERT INTO review_events(vocabulary, rating, reviewed_at_ms, source, raw_json)
		VALUES ('manual', 'good', 1, 'manual', '{}')`); err != nil {
		t.Fatal(err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		response := httptest.NewRecorder()
		importer.importVocabulary(response, httptest.NewRequest(http.MethodPost, "/anki/import-vocabulary", nil))
		if response.Code != 200 {
			t.Fatalf("import failed: %d %s", response.Code, response.Body.String())
		}
		var summary importSummary
		if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
			t.Fatal(err)
		}
		if summary.CardCount != 2 || summary.ReviewCount != 6 || summary.UnscoredReviewCount != 1 || summary.MissingReadingCount != 0 || summary.ImportedAt == nil {
			t.Fatalf("unexpected summary: %+v", summary)
		}
	}
	var total, manual int
	if err := db.QueryRow("SELECT COUNT(*), SUM(CASE WHEN source = 'manual' THEN 1 ELSE 0 END) FROM review_events").Scan(&total, &manual); err != nil {
		t.Fatal(err)
	}
	if total != 8 || manual != 1 {
		t.Fatalf("repeat import duplicated history or removed manual data: total=%d manual=%d", total, manual)
	}
	store := reviewStore{db: db}
	response := httptest.NewRecorder()
	store.listKanjiStats(response, httptest.NewRequest(http.MethodGet, "/kanji-stats", nil))
	var kanji []kanjiStats
	if err := json.Unmarshal(response.Body.Bytes(), &kanji); err != nil {
		t.Fatal(err)
	}
	for _, entry := range kanji {
		if entry.Kanji == "書" {
			if entry.Difficulty == nil || *entry.Difficulty != 34 || entry.QualifiedWordCount != 2 {
				t.Fatalf("real history was not used: %+v", entry)
			}
			for _, word := range entry.Evidence {
				if word.Reading == "" {
					t.Error("imported reading missing from evidence")
				}
			}
		}
	}

	fail.Store(true)
	response = httptest.NewRecorder()
	importer.importVocabulary(response, httptest.NewRequest(http.MethodPost, "/anki/import-vocabulary", nil))
	if response.Code != 502 {
		t.Fatalf("expected Anki failure, got %d", response.Code)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM anki_cards").Scan(&total); err != nil || total != 2 {
		t.Fatalf("failed fetch changed saved data: count=%d err=%v", total, err)
	}
	fail.Store(false)
	removeSecond.Store(true)
	response = httptest.NewRecorder()
	importer.importVocabulary(response, httptest.NewRequest(http.MethodPost, "/anki/import-vocabulary", nil))
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM anki_cards").Scan(&total); err != nil || total != 1 {
		t.Fatalf("deleted Anki card was retained: count=%d err=%v", total, err)
	}
}

func TestVocabularySnapshotRollback(t *testing.T) {
	db := testDatabase(t)
	ctx := context.Background()
	cards := []ankiCard{{CardID: 1, NoteID: 10, DeckName: vocabularyDeck, Vocabulary: "書く", Reading: "かく"}}
	history := map[int64][]ankiReview{1: {{ID: 100, Ease: 3, Type: 1}}}
	if err := saveVocabularySnapshot(ctx, db, cards, history); err != nil {
		t.Fatal(err)
	}
	history[1] = append(history[1], history[1][0]) // Duplicate Anki review ID must fail.
	if err := saveVocabularySnapshot(ctx, db, cards, history); err == nil {
		t.Fatal("expected a duplicate-review error")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM review_events").Scan(&count); err != nil || count != 1 {
		t.Fatalf("transaction failed to preserve the previous snapshot: count=%d err=%v", count, err)
	}
}

func TestAnkiTextAndReadings(t *testing.T) {
	cases := []struct{ front, back, word, reading string }{
		{"<b>表現</b>", "ひょうげん<br>Expression", "表現", "ひょうげん"},
		{"気休め", "【読み】<br>きやすめ<br>【英語】<br>reassurance", "気休め", "きやすめ"},
		{"浮気(する)", "うわき(する)<br>cheating", "浮気(する)", "うわき(する)"},
		{"さらに", "furthermore<br>例文", "さらに", "さらに"},
		{"<ruby>書<rt>しょ</rt></ruby>[sound:a.mp3]", "write", "書", ""},
		{"<script>bad()</script>書く", "write<br>これは例文です", "書く", ""},
		{"書く&nbsp;", "<div>【読み】かく</div><div>write</div>", "書く", "かく"},
	}
	for _, tc := range cases {
		word := strings.Join(strings.Fields(plainAnkiText(tc.front)), " ")
		reading := extractAnkiReading(word, plainAnkiText(tc.back))
		if word != tc.word || reading != tc.reading {
			t.Errorf("got word=%q reading=%q, want %q %q", word, reading, tc.word, tc.reading)
		}
	}
	mapping := kanjiMappingForReviews([]review{{Vocabulary: "時々"}, {Vocabulary: "食べる"}, {Vocabulary: "さらに"}})
	if len(mapping["時々"]) != 1 || mapping["時々"][0] != "時" || len(mapping["さらに"]) != 0 {
		t.Fatalf("incorrect kanji extraction: %v", mapping)
	}
}

func TestImportStatusStartsEmpty(t *testing.T) {
	importer := ankiImporter{db: testDatabase(t)}
	response := httptest.NewRecorder()
	importer.importStatus(response, httptest.NewRequest(http.MethodGet, "/anki/import-status", nil))
	var summary importSummary
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.CardCount != 0 || summary.ReviewCount != 0 || summary.ImportedAt != nil {
		t.Fatalf("unexpected initial status: %+v", summary)
	}
}

func TestImportedReviewsSurviveReopening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saved.sqlite")
	db, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cards := []ankiCard{{CardID: 5, NoteID: 50, DeckName: vocabularyDeck, Vocabulary: "書く", Reading: "かく"}}
	history := map[int64][]ankiReview{5: {{ID: 1234567890123, Ease: 4, Type: 1}}}
	if err := saveVocabularySnapshot(context.Background(), db, cards, history); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	store := reviewStore{db: reopened}
	reviews, err := store.loadReviews(context.Background(), 0)
	if err != nil || len(reviews) != 1 {
		t.Fatalf("persistent review missing: %v, %v", reviews, err)
	}
	if reviews[0].Vocabulary != "書く" || reviews[0].Rating != "easy" || reviews[0].Source != "anki" || reviews[0].ReviewedAt != 1234567890123 {
		t.Fatalf("persistent review changed: %+v", reviews[0])
	}
}

func TestAnkiRatingMapping(t *testing.T) {
	for ease, want := range map[int]string{0: "", 1: "again", 2: "hard", 3: "good", 4: "easy", 5: ""} {
		if got := ankiRating(ankiReview{Ease: ease, Type: 1}); got != want {
			t.Errorf("ease %d: got %q, want %q", ease, got, want)
		}
	}
	if got := ankiRating(ankiReview{Ease: 3, Type: 4}); got != "" {
		t.Error("manual/rescheduling records must not be treated as recalled answers")
	}
}
