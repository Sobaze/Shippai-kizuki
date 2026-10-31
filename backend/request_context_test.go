package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCancelledDatabaseRequestsDoNotWriteServerErrors(t *testing.T) {
	db := testDatabase(t)
	store := reviewStore{db: db}
	dictionary := kanjiDictionary{db: db}
	importer := ankiImporter{db: db}
	cases := []struct {
		name    string
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{"reviews", "GET", "/reviews", store.listReviews},
		{"vocabulary", "GET", "/vocabulary-stats", store.listVocabularyStats},
		{"kanji", "GET", "/kanji-stats", store.listKanjiStats},
		{"import status", "GET", "/anki/import-status", importer.importStatus},
		{"dictionary", "GET", "/kanji/書", dictionary.lookup},
		{"manual review", "POST", "/reviews", store.createReview},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"vocabulary":"書く","rating":"good"}`)).WithContext(ctx)
			request.SetPathValue("character", "書")
			response := httptest.NewRecorder()
			tc.handler(response, request)
			if response.Body.Len() != 0 {
				t.Errorf("cancelled request produced an error response: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
