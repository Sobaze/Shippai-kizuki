package quiz

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestN2CatalogMergesWithExistingPools(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData, n2QuizData)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Pools) != 6 {
		t.Fatalf("got %d pools", len(catalog.Pools))
	}
	for _, test := range []struct {
		mode  string
		count int
	}{{"meaning", 374}, {"reading", 1640}} {
		response := httptest.NewRecorder()
		catalog.listQuestions(response, httptest.NewRequest("GET", "/quiz/questions?mode="+test.mode+"&level=N2", nil))
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		var pool quizPool
		if err := json.Unmarshal(response.Body.Bytes(), &pool); err != nil {
			t.Fatal(err)
		}
		if len(pool.Questions) != test.count || pool.Level != "N2" {
			t.Fatalf("N2 pool metadata: %d %+v", len(pool.Questions), pool.Level)
		}
		for _, q := range pool.Questions {
			if q.Mode != test.mode || q.EstimatedLevel != "N2" {
				t.Fatalf("missing N2 snapshot level: %+v", q)
			}
		}
	}
	// The original snapshots keep their IDs and content; merging adds pools.
	original, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	for i, pool := range original.Pools {
		before, _ := json.Marshal(pool)
		after, _ := json.Marshal(catalog.Pools[i])
		if string(before) != string(after) {
			t.Fatal("N2 merge modified a previous pool")
		}
	}
}

func TestN2AttemptsCoverageAndQueue(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData, n2QuizData)
	if err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: testDatabase(t)}
	for index, pool := range catalog.Pools[4:] {
		input := quizFixture(pool.Mode, "full", 10)
		input.ID = "n2-" + pool.Mode
		input.Questions = pool.Questions[:10]
		input.Answers = nil
		for _, q := range input.Questions {
			input.Answers = append(input.Answers, quizAnswer{QuestionID: q.ID, SelectedAnswer: q.Choices[1]})
		}
		for _, status := range []int{201, 200} {
			response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
			if response.Code != status {
				t.Fatal(response.Body.String())
			}
		}
		coverage, err := store.poolCoverage(context.Background(), pool)
		if err != nil || coverage.Attempted != 10 || coverage.Total != len(pool.Questions) || coverage.Unseen != len(pool.Questions)-10 {
			t.Fatalf("N2 coverage: %+v %v", coverage, err)
		}
		candidates, err := store.practiceCandidates(context.Background(), pool.Mode)
		if err != nil || len(candidates) != 10 {
			t.Fatalf("N2 queue: %+v %v", candidates, err)
		}
		retry := input
		retry.Kind, retry.ID = "retry", input.ID+"-retry"
		retry.Questions = input.Questions[:1]
		retry.Answers = []quizAnswer{{QuestionID: retry.Questions[0].ID, SelectedAnswer: retry.Questions[0].Answer}}
		retry.CompletedAt = "2026-11-01T11:00:00Z"
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", retry); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
		retry.ID += "-again"
		retry.CompletedAt = "2026-11-01T12:00:00Z"
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", retry); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
		candidates, err = store.practiceCandidates(context.Background(), pool.Mode)
		if err != nil || len(candidates) != 9 {
			t.Fatal("two-correct rule broke for N2")
		}
		legacyCoverage, err := store.poolCoverage(context.Background(), catalog.Pools[index*2])
		if err != nil || legacyCoverage.Attempted != 0 {
			t.Fatal("N2 exposure leaked into older pools")
		}
	}
}
