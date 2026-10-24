package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReviewRoundTrip(t *testing.T) {
	store := reviewStore{}
	list := httptest.NewRecorder()
	store.listReviews(list, httptest.NewRequest(http.MethodGet, "/reviews", nil))
	if list.Code != http.StatusOK || strings.TrimSpace(list.Body.String()) != "[]" {
		t.Fatalf("empty list: status %d, body %s", list.Code, list.Body.String())
	}

	for _, rating := range []string{"again", "hard", "good", "easy"} {
		body := `{"vocabulary":"  図書館  ","rating":"` + rating + `"}`
		response := httptest.NewRecorder()
		store.createReview(response, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(body)))
		if response.Code != http.StatusCreated {
			t.Fatalf("create %s: status %d, body %s", rating, response.Code, response.Body.String())
		}
		if got := response.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		var created review
		if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
			t.Fatal(err)
		}
		if created.Vocabulary != "図書館" || created.Rating != rating {
			t.Fatalf("unexpected created review: %+v", created)
		}
	}

	list = httptest.NewRecorder()
	store.listReviews(list, httptest.NewRequest(http.MethodGet, "/reviews", nil))
	if list.Code != http.StatusOK || list.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected list response: %d, %v", list.Code, list.Header())
	}
	var entries []review
	if err := json.Unmarshal(list.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("got %d reviews, want 4", len(entries))
	}
	for i, rating := range []string{"again", "hard", "good", "easy"} {
		if entries[i].Vocabulary != "図書館" || entries[i].Rating != rating {
			t.Errorf("review %d = %+v", i, entries[i])
		}
	}
}

func TestCreateReviewRejectsInvalidInput(t *testing.T) {
	cases := map[string]string{
		"empty body":         ``,
		"malformed JSON":     `{"vocabulary":`,
		"array":              `[]`,
		"null":               `null`,
		"missing vocabulary": `{"rating":"hard"}`,
		"blank vocabulary":   `{"vocabulary":" \t ","rating":"hard"}`,
		"missing rating":     `{"vocabulary":"食べる"}`,
		"unknown rating":     `{"vocabulary":"食べる","rating":"failed"}`,
		"wrong field type":   `{"vocabulary":42,"rating":"hard"}`,
		"unknown field":      `{"vocabulary":"食べる","rating":"hard","mistakeType":"reading"}`,
		"multiple objects":   `{"vocabulary":"食べる","rating":"hard"} {}`,
		"trailing garbage":   `{"vocabulary":"食べる","rating":"hard"} junk`,
		"oversized body":     `{"vocabulary":"` + strings.Repeat("a", 1024*1024) + `","rating":"hard"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			store := reviewStore{}
			response := httptest.NewRecorder()
			store.createReview(response, httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(body)))
			if response.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body = %s", response.Code, response.Body.String())
			}
			if len(store.reviews) != 0 {
				t.Error("invalid input was stored")
			}
		})
	}
}
