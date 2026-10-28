package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDictionaryLookup(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/書" {
			t.Errorf("unexpected upstream path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"kanji":"書","meanings":["write"],"on_readings":["ショ"],"kun_readings":["か.く","-が.き","-がき"],"name_readings":[],"stroke_count":10}`))
	}))
	defer upstream.Close()
	dictionary := kanjiDictionary{db: testDatabase(t), client: upstream.Client(), baseURL: upstream.URL + "/"}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /kanji/{character}", dictionary.lookup)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/kanji/書", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("response: %d %s", response.Code, response.Body.String())
	}
	var got kanjiDetails
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := kanjiDetails{
		Kanji: "書", Meanings: []string{"write"}, OnReadings: []string{"ショ"},
		KunReadings: []string{"か.く", "-が.き", "-がき"}, NameReadings: []string{}, StrokeCount: 10,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestDictionaryErrorsAndEmptyReadings(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		wantCode int
	}{
		{"not found", 404, `{}`, 404},
		{"service failure", 503, `unavailable`, 502},
		{"invalid JSON", 200, `<html>error</html>`, 502},
		{"wrong character", 200, `{"kanji":"日","stroke_count":4}`, 502},
		{"incomplete entry", 200, `{}`, 502},
		{"empty reading lists", 200, `{"kanji":"書","stroke_count":10}`, 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer upstream.Close()
			db := testDatabase(t)
			dictionary := kanjiDictionary{db: db, client: upstream.Client(), baseURL: upstream.URL + "/"}
			request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
			request.SetPathValue("character", "書")
			response := httptest.NewRecorder()
			dictionary.lookup(response, request)
			if response.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.wantCode, response.Body.String())
			}
			var rows int
			if err := db.QueryRow("SELECT COUNT(*) FROM kanji_entries").Scan(&rows); err != nil {
				t.Fatal(err)
			}
			if (tc.wantCode == 200 && rows != 1) || (tc.wantCode != 200 && rows != 0) {
				t.Fatalf("unexpected saved row count: %d", rows)
			}
			if tc.wantCode == 200 {
				var got kanjiDetails
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Meanings == nil || got.OnReadings == nil || got.KunReadings == nil || got.NameReadings == nil {
					t.Error("empty lists must be arrays rather than null")
				}
			}
		})
	}
}

func TestDictionaryRejectsNonKanji(t *testing.T) {
	// A nil client also ensures invalid input cannot make an outgoing request.
	dictionary := kanjiDictionary{}
	for _, character := range []string{"", "図書館", "a", "あ", "1"} {
		request := httptest.NewRequest(http.MethodGet, "/kanji/test", nil)
		request.SetPathValue("character", character)
		response := httptest.NewRecorder()
		dictionary.lookup(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%q: got %d, want 400", character, response.Code)
		}
	}
}

func TestDictionaryTimeout(t *testing.T) {
	client := &http.Client{
		Timeout:   time.Second,
		Transport: dictionaryTimeoutTransport{},
	}
	dictionary := kanjiDictionary{db: testDatabase(t), client: client, baseURL: "https://example.invalid/"}
	request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
	request.SetPathValue("character", "書")
	response := httptest.NewRecorder()
	dictionary.lookup(response, request)
	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", response.Code)
	}
}

type dictionaryTimeoutTransport struct{}

func (dictionaryTimeoutTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

func TestDictionaryCacheSurvivesReopen(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"kanji":"書","meanings":["write"],"on_readings":["ショ"],"kun_readings":["か.く"],"name_readings":["かき"],"stroke_count":10}`))
	}))
	defer upstream.Close()
	path := filepath.Join(t.TempDir(), "dictionary.sqlite")
	db, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dictionary := kanjiDictionary{db: db, client: upstream.Client(), baseURL: upstream.URL + "/"}

	var firstBody string
	for _, wantSource := range []string{"kanjiapi", "sqlite"} {
		request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
		request.SetPathValue("character", "書")
		response := httptest.NewRecorder()
		dictionary.lookup(response, request)
		if response.Code != 200 || response.Header().Get("X-Dictionary-Source") != wantSource {
			t.Fatalf("expected %s response: %d %v %s", wantSource, response.Code, response.Header(), response.Body.String())
		}
		if firstBody == "" {
			firstBody = response.Body.String()
		} else if response.Body.String() != firstBody {
			t.Fatal("cached response differs from the fetched response")
		}
	}
	var source, fetchedAt string
	if err := db.QueryRow("SELECT source, fetched_at FROM kanji_entries WHERE kanji = ?", "書").Scan(&source, &fetchedAt); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(source, upstream.URL) {
		t.Fatalf("unexpected dictionary source: %q", source)
	}
	if _, err := time.Parse(time.RFC3339, fetchedAt); err != nil {
		t.Fatalf("invalid fetch timestamp: %v", err)
	}

	// Reopening the file simulates a backend restart. The upstream is now offline.
	upstream.Close()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	dictionary.db = reopened
	request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
	request.SetPathValue("character", "書")
	response := httptest.NewRecorder()
	dictionary.lookup(response, request)
	if response.Code != 200 || response.Body.String() != firstBody || response.Header().Get("X-Dictionary-Source") != "sqlite" {
		t.Fatalf("persistent cache failed: %d %s", response.Code, response.Body.String())
	}
	if calls.Load() != 1 {
		t.Fatalf("made %d upstream calls, want exactly one", calls.Load())
	}
}

func TestDictionaryDatabaseErrors(t *testing.T) {
	t.Run("read failure does not trigger a network lookup", func(t *testing.T) {
		db := testDatabase(t)
		db.Close()
		dictionary := kanjiDictionary{db: db}
		request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
		request.SetPathValue("character", "書")
		response := httptest.NewRecorder()
		dictionary.lookup(response, request)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("read failure: status %d", response.Code)
		}
	})
	t.Run("write failure does not report a saved lookup", func(t *testing.T) {
		db := testDatabase(t)
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db.Close()
			_, _ = w.Write([]byte(`{"kanji":"書","meanings":["write"],"stroke_count":10}`))
		}))
		defer upstream.Close()
		dictionary := kanjiDictionary{db: db, client: upstream.Client(), baseURL: upstream.URL + "/"}
		request := httptest.NewRequest(http.MethodGet, "/kanji/書", nil)
		request.SetPathValue("character", "書")
		response := httptest.NewRecorder()
		dictionary.lookup(response, request)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("write failure: status %d", response.Code)
		}
	})
}
