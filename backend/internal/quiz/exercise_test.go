package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func exerciseFixture(t *testing.T) quizSubmission {
	t.Helper()
	catalog, err := loadQuizCatalog(exerciseQuizData)
	if err != nil {
		t.Fatal(err)
	}
	input := quizSubmission{ID: "exercise-attempt", Mode: "sentence", Kind: "full", StartedAt: "2026-11-01T10:00:00Z", CompletedAt: "2026-11-01T10:01:00Z", Questions: catalog.Pools[0].Questions}
	for _, q := range input.Questions {
		input.Answers = append(input.Answers, quizAnswer{QuestionID: q.ID, SelectedAnswer: q.Answer})
	}
	return input
}

func TestExerciseFourSourceKeyAndAPI(t *testing.T) {
	input := exerciseFixture(t)
	key := []int{1, 3, 4, 3, 1, 3, 1, 3, 4, 3, 4, 2}
	if len(input.Questions) != len(key) {
		t.Fatal("Exercise 4 must have twelve questions")
	}
	for i, q := range input.Questions {
		if q.Choices[key[i]-1] != q.Answer || q.ExerciseID != "exercise-4" || q.Context == nil {
			t.Fatalf("source key/context mismatch at %d", i+1)
		}
		if (i%2 == 0 && q.Task != "reading") || (i%2 == 1 && q.Task != "spelling") {
			t.Fatal("reading/spelling task identity changed")
		}
	}
	result, err := validateQuizSubmission(&input)
	if err != nil || result.CorrectCount != 12 || result.Total != 12 {
		t.Fatalf("full exercise: %+v %v", result, err)
	}
	catalog, err := loadQuizCatalog(quizData, n2QuizData, exerciseQuizData)
	if err != nil || len(catalog.Pools) != 7 {
		t.Fatalf("combined catalog: %v", err)
	}
	response := httptest.NewRecorder()
	catalog.listQuestions(response, httptest.NewRequest("GET", "/quiz/questions?mode=sentence&level=N2&exercise=exercise-4", nil))
	var pool quizPool
	if err := json.Unmarshal(response.Body.Bytes(), &pool); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || pool.Title != "Exercise 4" || pool.ExerciseID != "exercise-4" || len(pool.Questions) != 12 {
		t.Fatalf("exercise API: %+v", pool)
	}
	response = httptest.NewRecorder()
	catalog.listQuestions(response, httptest.NewRequest("GET", "/quiz/questions?mode=sentence&level=N2&exercise=exercise-99", nil))
	if response.Code != 400 {
		t.Fatal("unknown exercise accepted")
	}
}

func TestExerciseValidationAndSeparateTaskProgress(t *testing.T) {
	for name, change := range map[string]func(*quizSubmission){
		"incomplete": func(input *quizSubmission) {
			input.Questions = input.Questions[:11]
			input.Answers = input.Answers[:11]
		},
		"missing context":       func(input *quizSubmission) { input.Questions[0].Context = nil },
		"unknown task":          func(input *quizSubmission) { input.Questions[0].Task = "meaning" },
		"unknown exercise":      func(input *quizSubmission) { input.Questions[0].ExerciseID = "exercise-99" },
		"reading key mismatch":  func(input *quizSubmission) { input.Questions[0].Example.Reading = "different" },
		"spelling key mismatch": func(input *quizSubmission) { input.Questions[1].Example.Written = "different" },
	} {
		t.Run(name, func(t *testing.T) {
			input := exerciseFixture(t)
			change(&input)
			if _, err := validateQuizSubmission(&input); err == nil {
				t.Fatal("malformed exercise accepted")
			}
		})
	}
	store := quizStore{db: testDatabase(t)}
	input := exerciseFixture(t)
	for i, q := range input.Questions {
		input.Answers[i].SelectedAnswer = q.Choices[(indexOfChoice(q.Choices, q.Answer)+1)%4]
	}
	for _, status := range []int{201, 200} {
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != status {
			t.Fatal(response.Body.String())
		}
	}
	saved, err := store.loadAttempt(context.Background(), input.ID)
	if err != nil || saved.ExerciseID != "exercise-4" || saved.Total != 12 || saved.Result.Items[0].Question.Context.Before != input.Questions[0].Context.Before {
		t.Fatal("exercise snapshot lost")
	}
	retry := exerciseFixture(t)
	retry.ID, retry.Kind, retry.CompletedAt = "retry-all", "retry", "2026-11-01T11:00:00Z"
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", retry); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	candidates, err := store.practiceCandidates(context.Background(), "sentence")
	if err != nil || len(candidates) != 12 {
		t.Fatal("all twelve misses should remain after one correct retry")
	}
	retry.ID, retry.CompletedAt = "retry-readings", "2026-11-01T12:00:00Z"
	retry.Questions, retry.Answers = nil, nil
	for _, q := range input.Questions {
		if q.Task == "reading" {
			retry.Questions = append(retry.Questions, q)
			retry.Answers = append(retry.Answers, quizAnswer{QuestionID: q.ID, SelectedAnswer: q.Answer})
		}
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", retry); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	candidates, err = store.practiceCandidates(context.Background(), "sentence")
	if err != nil || len(candidates) != 6 {
		t.Fatal("reading progress must not clear spelling targets")
	}
	for _, c := range candidates {
		if c.Question.Task != "spelling" || c.Streak != 1 {
			t.Fatal("spelling streak changed with reading retry")
		}
	}
	catalog, _ := loadQuizCatalog(exerciseQuizData)
	coverage, err := store.poolCoverage(context.Background(), catalog.Pools[0])
	if err != nil || coverage.Attempted != 12 || coverage.ExerciseID != "exercise-4" {
		t.Fatal("exercise coverage broken")
	}
	history, err := store.recentAttempts(context.Background(), "sentence")
	if err != nil || history[0].ExerciseID != "exercise-4" {
		t.Fatal("exercise identity missing in history")
	}
}

func indexOfChoice(choices []string, answer string) int {
	for i, choice := range choices {
		if choice == answer {
			return i
		}
	}
	return -1
}

const legacyQuizSchema = `CREATE TABLE quiz_attempts (
	id TEXT PRIMARY KEY NOT NULL, mode TEXT NOT NULL CHECK (mode IN ('meaning', 'reading')),
	kind TEXT NOT NULL CHECK (kind IN ('full', 'retry')), parent_attempt_id TEXT,
	started_at TEXT NOT NULL, completed_at TEXT NOT NULL, completed_at_ms INTEGER NOT NULL,
	correct_count INTEGER NOT NULL, total INTEGER NOT NULL CHECK (total BETWEEN 1 AND 10),
	result_json TEXT NOT NULL, CHECK (correct_count BETWEEN 0 AND total));`

func TestSentenceMigrationPreservesExistingAttempts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(legacyQuizSchema + `CREATE TABLE preserve_me (value TEXT); INSERT INTO preserve_me VALUES ('keep');`); err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: db}
	legacy := quizFixture("reading", "full", 10)
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacy); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	legacyRetry := quizFixture("reading", "retry", 2)
	legacyRetry.ID, legacyRetry.ParentAttemptID = "legacy-retry", &legacy.ID
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacyRetry); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	var before string
	if err := db.QueryRow(`SELECT result_json FROM quiz_attempts WHERE id = 'attempt-1'`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store.db = db
	var after, marker string
	if err := db.QueryRow(`SELECT result_json FROM quiz_attempts WHERE id = 'attempt-1'`).Scan(&after); err != nil || before != after {
		t.Fatal("migration rewrote a legacy result")
	}
	if err := db.QueryRow(`SELECT value FROM preserve_me`).Scan(&marker); err != nil || marker != "keep" {
		t.Fatal("unrelated table changed")
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacy); response.Code != 200 {
		t.Fatal("legacy idempotent resave broke")
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacyRetry); response.Code != 200 {
		t.Fatal("legacy retry or parent reference changed")
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", exerciseFixture(t)); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = openDatabase(path)
	if err != nil {
		t.Fatalf("migration is not repeatable: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM quiz_attempts`).Scan(&count); err != nil || count != 3 {
		t.Fatal("migration lost saved attempts")
	}
}
