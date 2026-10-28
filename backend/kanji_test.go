package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKanjiEvidenceRules(t *testing.T) {
	cases := []struct {
		name           string
		words          []vocabularyStats
		wantScore      int // -1 means no score.
		qualified      int
		recommendation string
	}{
		{
			name: "early failures do not distort qualified words",
			words: []vocabularyStats{
				{Vocabulary: "A", Difficulty: 80, ReviewCount: 5},
				{Vocabulary: "B", Difficulty: 20, ReviewCount: 3},
				{Vocabulary: "C", Difficulty: 100, ReviewCount: 1},
			},
			wantScore: 50, qualified: 2, recommendation: "suggested_practice",
		},
		{
			name: "early successes do not lower qualified scores",
			words: []vocabularyStats{
				{Vocabulary: "A", Difficulty: 100, ReviewCount: 3},
				{Vocabulary: "B", Difficulty: 0, ReviewCount: 2},
			},
			wantScore: 100, qualified: 1, recommendation: "limited_cross_word_evidence",
		},
		{
			name: "words have equal influence despite different review counts",
			words: []vocabularyStats{
				{Vocabulary: "A", Difficulty: 80, ReviewCount: 100},
				{Vocabulary: "B", Difficulty: 20, ReviewCount: 3},
			},
			wantScore: 50, qualified: 2, recommendation: "suggested_practice",
		},
		{
			name: "only early evidence has no score",
			words: []vocabularyStats{
				{Vocabulary: "A", Difficulty: 100, ReviewCount: 1},
				{Vocabulary: "B", Difficulty: 100, ReviewCount: 2},
			},
			wantScore: -1, qualified: 0, recommendation: "limited_review_evidence",
		},
		{
			name: "below practice threshold",
			words: []vocabularyStats{
				{Vocabulary: "A", Difficulty: 48, ReviewCount: 3},
				{Vocabulary: "B", Difficulty: 50, ReviewCount: 3},
			},
			wantScore: 49, qualified: 2, recommendation: "no_extra_practice",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mapping := map[string][]string{"A": {"書", "書"}, "B": {"書"}, "C": {"書"}}
			got := calculateKanjiStats(tc.words, mapping)
			if len(got) != 1 {
				t.Fatalf("got %d kanji, want 1", len(got))
			}
			stats := got[0]
			if stats.WordCount != len(tc.words) || len(stats.Evidence) != len(tc.words) || stats.QualifiedWordCount != tc.qualified || stats.Recommendation != tc.recommendation {
				t.Fatalf("unexpected stats: %+v", stats)
			}
			if tc.wantScore < 0 {
				if stats.Difficulty != nil {
					t.Errorf("got score %d, want no score", *stats.Difficulty)
				}
			} else if stats.Difficulty == nil || *stats.Difficulty != tc.wantScore {
				t.Errorf("unexpected difficulty, want %d: %+v", tc.wantScore, stats)
			}
			for _, evidence := range stats.Evidence {
				if evidence.Qualifies != (evidence.ReviewCount >= 3) {
					t.Errorf("incorrect qualification: %+v", evidence)
				}
			}
		})
	}
}

func TestKanjiDemoAndOrdering(t *testing.T) {
	stats := calculateKanjiStats(calculateVocabularyStats(demoReviews()), demoKanjiMapping())
	if len(stats) != 8 || stats[0].Kanji != "書" || stats[0].Difficulty == nil || *stats[0].Difficulty != 64 || stats[0].QualifiedWordCount != 3 {
		t.Fatalf("unexpected demo results: %+v", stats)
	}
	for i, entry := range stats {
		if i > 0 && stats[i-1].Difficulty == nil && entry.Difficulty != nil {
			t.Error("scored kanji must precede unscored kanji")
		}
		if entry.Kanji == "館" && entry.Recommendation != "limited_cross_word_evidence" {
			t.Errorf("unexpected hall evidence: %+v", entry)
		}
	}
	words := []vocabularyStats{{Vocabulary: "unknown", Difficulty: 100, ReviewCount: 5}}
	if got := calculateKanjiStats(words, demoKanjiMapping()); len(got) != 0 {
		t.Fatalf("unmapped vocabulary should not invent kanji: %+v", got)
	}
}

func TestKanjiStatsAPIAndNewReviews(t *testing.T) {
	store := reviewStore{db: testDatabase(t)}
	response := httptest.NewRecorder()
	store.listKanjiStats(response, httptest.NewRequest(http.MethodGet, "/kanji-stats", nil))
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("empty response: %d %s", response.Code, response.Body.String())
	}

	created := httptest.NewRecorder()
	store.createReview(created, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(`{"vocabulary":"日本語","rating":"easy"}`)))
	if created.Code != http.StatusCreated {
		t.Fatal("could not save the initial review")
	}
	for count := 1; count <= 3; count++ {
		response = httptest.NewRecorder()
		store.listKanjiStats(response, httptest.NewRequest(http.MethodGet, "/kanji-stats", nil))
		if response.Header().Get("Content-Type") != "application/json" {
			t.Fatal("expected JSON")
		}
		var stats []kanjiStats
		if err := json.Unmarshal(response.Body.Bytes(), &stats); err != nil {
			t.Fatal(err)
		}
		if len(stats) != 3 {
			t.Fatalf("got %d kanji, want 3", len(stats))
		}
		for _, entry := range stats {
			if entry.Evidence[0].Reading != "" {
				t.Errorf("a manual review must not invent an imported reading: %+v", entry.Evidence[0])
			}
			if count < 3 {
				if entry.Difficulty != nil || entry.Evidence[0].Qualifies {
					t.Errorf("early evidence should not contribute: %+v", entry)
				}
			} else if entry.Difficulty == nil || *entry.Difficulty != 0 || !entry.Evidence[0].Qualifies || entry.Recommendation != "limited_cross_word_evidence" {
				t.Errorf("third review should qualify the word, not establish a cross-word pattern: %+v", entry)
			}
		}
		if count < 3 {
			created := httptest.NewRecorder()
			store.createReview(created, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(`{"vocabulary":"日本語","rating":"good"}`)))
			if created.Code != http.StatusCreated {
				t.Fatalf("create: %d", created.Code)
			}
		}
	}
}
