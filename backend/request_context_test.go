package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sobaze/Shippai-kizuki/backend/internal/testutil"
)

func TestCancelledDatabaseRequestsDoNotWriteServerErrors(t *testing.T) {
	db := testutil.Database(t)
	mux, err := newMux(db)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"reviews", "GET", "/reviews"},
		{"vocabulary", "GET", "/vocabulary-stats"},
		{"kanji", "GET", "/kanji-stats"},
		{"import status", "GET", "/anki/import-status"},
		{"dictionary", "GET", "/kanji/書"},
		{"manual review", "POST", "/reviews"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"vocabulary":"書く","rating":"good"}`)).WithContext(ctx)
			request.SetPathValue("character", "書")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Body.Len() != 0 {
				t.Errorf("cancelled request produced an error response: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
