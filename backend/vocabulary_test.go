package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestVocabularyDifficulty(t *testing.T) {
	cases := []struct {
		name           string
		ratings        []string
		difficulty     int
		recommendation string
	}{
		{"mixed and rounded", []string{"again", "hard", "good", "again"}, 63, "suggested_practice"},
		{"at threshold", []string{"hard", "hard", "hard"}, 50, "suggested_practice"},
		{"below threshold", []string{"hard", "good", "easy"}, 17, "no_extra_practice"},
		{"all successful", []string{"good", "easy", "good"}, 0, "no_extra_practice"},
		{"all failed", []string{"again", "again", "again"}, 100, "suggested_practice"},
		{"two failures are limited evidence", []string{"again", "again"}, 100, "limited_evidence"},
		{"one success is limited evidence", []string{"easy"}, 0, "limited_evidence"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reviews []review
			for _, rating := range tc.ratings {
				reviews = append(reviews, review{Vocabulary: "図書館", Rating: rating})
			}
			got := calculateVocabularyStats(reviews)
			if len(got) != 1 {
				t.Fatalf("got %d words, want 1", len(got))
			}
			if got[0].Difficulty != tc.difficulty || got[0].Recommendation != tc.recommendation || got[0].ReviewCount != len(tc.ratings) {
				t.Fatalf("unexpected stats: %+v", got[0])
			}
		})
	}
}

func TestVocabularyGroupingAndOrdering(t *testing.T) {
	reviews := []review{
		{Vocabulary: "B", Rating: "hard"},
		{Vocabulary: "A", Rating: "hard"},
		{Vocabulary: "C", Rating: "again"},
		{Vocabulary: "B", Rating: "good"},
		{Vocabulary: "A", Rating: "easy"},
	}
	want := []vocabularyStats{
		{"C", 1, ratingCounts{Again: 1}, 100, "limited_evidence"},
		{"A", 2, ratingCounts{Hard: 1, Easy: 1}, 25, "limited_evidence"},
		{"B", 2, ratingCounts{Hard: 1, Good: 1}, 25, "limited_evidence"},
	}
	if got := calculateVocabularyStats(reviews); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestVocabularyStatsReflectNewReviews(t *testing.T) {
	store := reviewStore{}
	response := httptest.NewRecorder()
	store.listVocabularyStats(response, httptest.NewRequest(http.MethodGet, "/vocabulary-stats", nil))
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("empty stats: status %d, body %s", response.Code, response.Body.String())
	}

	for _, rating := range []string{"again", "hard", "good"} {
		body := `{"vocabulary":"図書館","rating":"` + rating + `"}`
		created := httptest.NewRecorder()
		store.createReview(created, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(body)))
		if created.Code != http.StatusCreated {
			t.Fatalf("create review: %d", created.Code)
		}
	}
	response = httptest.NewRecorder()
	store.listVocabularyStats(response, httptest.NewRequest(http.MethodGet, "/vocabulary-stats", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: %d, %v", response.Code, response.Header())
	}
	var got []vocabularyStats
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Difficulty != 50 || got[0].ReviewCount != 3 || got[0].Recommendation != "suggested_practice" {
		t.Fatalf("unexpected stats after POST: %+v", got)
	}
}
