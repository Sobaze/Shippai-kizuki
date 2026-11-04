package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sobaze/Shippai-kizuki/backend/internal/testutil"
)

func TestApplicationRoutes(t *testing.T) {
	mux, err := newMux(testutil.Database(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"/health", "/reviews", "/vocabulary-stats", "/kanji-stats", "/anki/import-status",
		"/quiz/questions?mode=meaning&level=N2", "/quiz/questions?mode=reading&level=N2",
		"/quiz/questions?mode=sentence&level=N2&exercise=exercise-4",
		"/quiz/coverage?mode=sentence&level=N2&exercise=exercise-4",
		"/quiz/attempts?mode=sentence", "/quiz/practice?mode=sentence",
	} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != 200 || !json.Valid(response.Body.Bytes()) {
				t.Fatalf("%d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestApplicationExerciseRoundTrip(t *testing.T) {
	mux, err := newMux(testutil.Database(t))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(mux)
	defer server.Close()
	response, err := server.Client().Get(server.URL + "/quiz/questions?mode=sentence&level=N2&exercise=exercise-4")
	if err != nil {
		t.Fatal(err)
	}
	var pool struct {
		Questions []json.RawMessage `json:"questions"`
	}
	err = json.NewDecoder(response.Body).Decode(&pool)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 || len(pool.Questions) != 12 {
		t.Fatalf("load exercise: %v", err)
	}
	answers := []map[string]string{}
	for _, encoded := range pool.Questions {
		var question struct{ ID, Answer string }
		if err := json.Unmarshal(encoded, &question); err != nil {
			t.Fatal(err)
		}
		answers = append(answers, map[string]string{"questionId": question.ID, "selectedAnswer": question.Answer})
	}
	input := map[string]any{"id": "router-exercise", "mode": "sentence", "kind": "full",
		"startedAt": "2026-11-01T10:00:00Z", "completedAt": "2026-11-01T10:01:00Z",
		"questions": pool.Questions, "answers": answers}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	response, err = server.Client().Post(server.URL+"/quiz/attempts", "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 201 {
		t.Fatalf("save exercise: %d", response.StatusCode)
	}
	response, err = server.Client().Get(server.URL + "/quiz/attempts/router-exercise")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var saved struct {
		ID, ExerciseID      string
		CorrectCount, Total int
	}
	if err := json.NewDecoder(response.Body).Decode(&saved); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || saved.ID != "router-exercise" || saved.ExerciseID != "exercise-4" || saved.CorrectCount != 12 || saved.Total != 12 {
		t.Fatalf("saved exercise changed: %+v", saved)
	}
}
