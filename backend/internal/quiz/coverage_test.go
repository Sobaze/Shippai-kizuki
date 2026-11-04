package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestQuizCoverageCountsSavedExposureInCurrentPools(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: testDatabase(t)}
	read := func(mode, level string) quizCoverage {
		t.Helper()
		response := httptest.NewRecorder()
		store.listCoverage(&catalog)(response, httptest.NewRequest("GET", "/quiz/coverage?mode="+mode+"&level="+level, nil))
		if response.Code != 200 {
			t.Fatalf("coverage: %s", response.Body.String())
		}
		if response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("coverage must not be cached")
		}
		var coverage quizCoverage
		if err := json.Unmarshal(response.Body.Bytes(), &coverage); err != nil {
			t.Fatal(err)
		}
		return coverage
	}
	for _, pool := range catalog.Pools {
		coverage := read(pool.Mode, pool.Level)
		if coverage.Total != 24 || coverage.Attempted != 0 || coverage.Unseen != 24 || len(coverage.Questions) != 24 {
			t.Fatalf("empty coverage: %+v", coverage)
		}
		for i, entry := range coverage.Questions {
			if entry.QuestionID != pool.Questions[i].ID || entry.Attempts != 0 || entry.LastPracticedAt != nil {
				t.Fatalf("unseen entry: %+v", entry)
			}
		}
	}
	save := func(id, at string, questions []quizQuestion, kind string) quizSubmission {
		t.Helper()
		input := quizSubmission{ID: id, Mode: questions[0].Mode, Kind: kind, StartedAt: at, CompletedAt: at, Questions: questions}
		for _, q := range questions {
			input.Answers = append(input.Answers, quizAnswer{QuestionID: q.ID, SelectedAnswer: q.Choices[1]})
		}
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
		return input
	}
	pool := catalog.Pools[2] // Reading N5.
	legacy := pool.Questions[0]
	legacy.EstimatedLevel = ""
	old := save("legacy", "2026-11-01T10:00:00Z", []quizQuestion{legacy}, "retry")
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", old); response.Code != 200 {
		t.Fatal("idempotent save failed")
	}
	coverage := read("reading", "N5")
	if coverage.Attempted != 1 || coverage.Unseen != 23 || coverage.Questions[0].Attempts != 1 {
		t.Fatalf("legacy/duplicate accounting: %+v", coverage)
	}
	save("full", "2026-11-01T12:00:00.000000002Z", pool.Questions[:10], "full")
	// Insert an older attempt afterwards; last practice is by completion time.
	save("older-retry", "2026-11-01T12:00:00.000000001Z", pool.Questions[:1], "retry")
	// More than 20 saved attempts must contribute, even outside the history list.
	for i := range 25 {
		save(fmt.Sprintf("extra-%d", i), "2026-11-01T11:00:00Z", pool.Questions[:1], "retry")
	}
	coverage = read("reading", "N5")
	if coverage.Attempted != 10 || coverage.Unseen != 14 || coverage.Questions[0].Attempts != 28 || *coverage.Questions[0].LastPracticedAt != "2026-11-01T12:00:00.000000002Z" {
		t.Fatalf("full history accounting: %+v", coverage)
	}
	for i := 1; i < 10; i++ {
		if coverage.Questions[i].Attempts != 1 {
			t.Fatal("full quiz items should count once regardless of correctness")
		}
	}
	// Other levels, retired IDs, and identical IDs in another mode do not count.
	save("other-level", "2026-11-01T13:00:00Z", catalog.Pools[3].Questions[:1], "retry")
	retired := legacy
	retired.ID = "reading-retired"
	save("retired", "2026-11-01T14:00:00Z", []quizQuestion{retired}, "retry")
	otherMode := catalog.Pools[0].Questions[0]
	otherMode.ID = legacy.ID
	save("other-mode", "2026-11-01T15:00:00Z", []quizQuestion{otherMode}, "retry")
	coverage = read("reading", "N5")
	if coverage.Attempted != 10 || coverage.Questions[0].Attempts != 28 {
		t.Fatal("unrelated items affected this pool")
	}
	if other := read("reading", "N4"); other.Attempted != 1 || other.Unseen != 23 {
		t.Fatal("levels not separated")
	}
	if other := read("meaning", "N5"); other.Attempted != 0 {
		t.Fatal("IDs outside the current pool must be ignored")
	}
}

func TestQuizCoveragePersistsAndHandlesErrors(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "coverage.sqlite")
	db, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: db}
	input := quizFixture("meaning", "retry", 1)
	input.Questions = catalog.Pools[0].Questions[:1]
	input.Answers = []quizAnswer{{QuestionID: input.Questions[0].ID, SelectedAnswer: input.Questions[0].Answer}}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store.db = db
	coverage, err := store.poolCoverage(context.Background(), catalog.Pools[0])
	if err != nil || coverage.Attempted != 1 {
		t.Fatalf("reopened coverage: %+v %v", coverage, err)
	}
	for _, query := range []string{"", "mode=reading&level=starter", "mode=meaning&level=N1", "mode=other&level=N5"} {
		response := httptest.NewRecorder()
		store.listCoverage(&catalog)(response, httptest.NewRequest("GET", "/quiz/coverage?"+query, nil))
		if response.Code != 400 {
			t.Fatalf("accepted invalid query %q", query)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.poolCoverage(ctx, catalog.Pools[0]); err == nil {
		t.Fatal("cancelled read succeeded")
	}
	response := httptest.NewRecorder()
	store.listCoverage(&catalog)(response, httptest.NewRequest("GET", "/quiz/coverage?mode=meaning&level=N5", nil).WithContext(ctx))
	if response.Body.Len() != 0 {
		t.Fatal("cancelled request should not write a server error")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	store.listCoverage(&catalog)(response, httptest.NewRequest("GET", "/quiz/coverage?mode=meaning&level=N5", nil))
	if response.Code != 500 {
		t.Fatal("database failure not reported")
	}
}
