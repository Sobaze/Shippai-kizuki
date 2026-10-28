package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAnkiDeckRequestAndResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected request: %s %v", r.Method, r.Header)
		}
		var payload struct {
			Action  string         `json:"action"`
			Version int            `json:"version"`
			Params  map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.Action != "deckNames" || payload.Version != 6 || payload.Params == nil || len(payload.Params) != 0 {
			t.Errorf("unexpected Anki action: %+v", payload)
		}
		_, _ = w.Write([]byte(`{"result":["JP Grammar","JP Sentences","JP Vocabulary","Standard"],"error":null}`))
	}))
	defer upstream.Close()
	anki := ankiClient{client: upstream.Client(), endpoint: upstream.URL}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /anki/decks", anki.listDecks)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/anki/decks", nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "application/json" || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected response: %d %v %s", response.Code, response.Header(), response.Body.String())
	}
	var got ankiDeckList
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := []string{"JP Grammar", "JP Sentences", "JP Vocabulary", "Standard"}
	if !reflect.DeepEqual(got.Decks, want) {
		t.Fatalf("got %v, want %v", got.Decks, want)
	}
}

func TestAnkiDeckErrorsAndEmptyList(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		wantCode int
	}{
		{"empty collection", 200, `{"result":[],"error":null}`, 200},
		{"action error despite HTTP success", 200, `{"result":null,"error":"permission denied"}`, 502},
		{"missing result", 200, `{"error":null}`, 502},
		{"invalid result type", 200, `{"result":42,"error":null}`, 502},
		{"malformed JSON", 200, `invalid`, 502},
		{"HTTP failure", 500, `unavailable`, 502},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer upstream.Close()
			anki := ankiClient{client: upstream.Client(), endpoint: upstream.URL}
			response := httptest.NewRecorder()
			anki.listDecks(response, httptest.NewRequest(http.MethodGet, "/anki/decks", nil))
			if response.Code != tc.wantCode {
				t.Fatalf("got %d, want %d: %s", response.Code, tc.wantCode, response.Body.String())
			}
			if tc.wantCode == 200 && response.Body.String() != "{\"decks\":[]}\n" {
				t.Fatalf("empty decks must be an array: %s", response.Body.String())
			}
		})
	}
}

func TestAnkiConnectionFailure(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	upstream.Close()
	anki := ankiClient{client: upstream.Client(), endpoint: upstream.URL}
	response := httptest.NewRecorder()
	anki.listDecks(response, httptest.NewRequest(http.MethodGet, "/anki/decks", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", response.Code)
	}
}

func TestAnkiTimeout(t *testing.T) {
	anki := ankiClient{
		client:   &http.Client{Transport: ankiTimeoutTransport{}},
		endpoint: "http://example.invalid/",
	}
	response := httptest.NewRecorder()
	anki.listDecks(response, httptest.NewRequest(http.MethodGet, "/anki/decks", nil))
	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("got %d, want 504", response.Code)
	}
}

type ankiTimeoutTransport struct{}

func (ankiTimeoutTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}
