package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func quizFixture(mode, kind string, count int) quizSubmission {
	input := quizSubmission{
		ID: "attempt-1", Mode: mode, Kind: kind,
		StartedAt: "2026-11-01T10:00:00.000Z", CompletedAt: "2026-11-01T10:01:00.000Z",
	}
	for index := 0; index < count; index++ {
		question := quizQuestion{
			ID: fmt.Sprintf("question-%d", index), Mode: mode, Prompt: "書", Answer: "write",
			Choices: []string{"write", "read", "walk", "sleep"},
			Example: quizExample{Written: "書く", Reading: "かく", Meaning: "write"},
		}
		if mode == "reading" {
			question.Prompt = "書く"
			question.Answer = "かく"
			question.Choices = []string{"かく", "きく", "よむ", "ねる"}
		}
		selected := question.Answer
		if index%2 == 1 {
			selected = question.Choices[1]
		}
		input.Questions = append(input.Questions, question)
		input.Answers = append(input.Answers, quizAnswer{QuestionID: question.ID, SelectedAnswer: selected})
	}
	return input
}

func quizRequest(t *testing.T, store *quizStore, method, path string, input any) *httptest.ResponseRecorder {
	t.Helper()
	var body string
	if raw, ok := input.(string); ok {
		body = raw
	} else if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		body = string(encoded)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /quiz/attempts", store.createAttempt)
	mux.HandleFunc("GET /quiz/attempts", store.listAttempts)
	mux.HandleFunc("GET /quiz/practice", store.listPracticeCandidates)
	mux.HandleFunc("GET /quiz/attempts/{id}", store.getAttempt)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
	return response
}

func TestQuizAttemptSaveIsIdempotentAndServerGraded(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	input := quizFixture("meaning", "full", 10)
	for index, wantCode := range []int{201, 200} {
		response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
		if response.Code != wantCode {
			t.Fatalf("save %d: %d %s", index, response.Code, response.Body.String())
		}
		var saved savedQuizAttempt
		if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
			t.Fatal(err)
		}
		if saved.CorrectCount != 5 || saved.Total != 10 || saved.Result.CorrectCount != 5 || len(saved.Result.Items) != 10 {
			t.Fatalf("incorrect server grading: %+v", saved)
		}
		if saved.Result.Items[1].Correct || saved.Result.Items[1].SelectedAnswer != "read" || saved.Result.Items[1].Question.Answer != "write" {
			t.Fatal("saved snapshot lost the selected or correct answer")
		}
	}
	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM quiz_attempts").Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate attempt saved: %d %v", count, err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM review_events").Scan(&count); err != nil || count != 0 {
		t.Fatal("quiz results must not enter Anki/manual review scoring")
	}
	input.Answers[0].SelectedAnswer = "read"
	response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
	if response.Code != 409 {
		t.Fatalf("changed saved ID: %d", response.Code)
	}
	response = quizRequest(t, &store, "GET", "/quiz/attempts/attempt-1", nil)
	var saved savedQuizAttempt
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.CorrectCount != 5 {
		t.Fatal("conflicting submission overwrote the original")
	}
}

func TestQuizValidationRejectsIncompleteOrMalformedAttempts(t *testing.T) {
	cases := map[string]func(*quizSubmission){
		"invalid mode":                func(input *quizSubmission) { input.Mode = "unknown" },
		"invalid kind":                func(input *quizSubmission) { input.Kind = "cancelled" },
		"invalid ID":                  func(input *quizSubmission) { input.ID = "bad/id" },
		"incomplete full quiz":        func(input *quizSubmission) { input.Questions = input.Questions[:9]; input.Answers = input.Answers[:9] },
		"missing answer":              func(input *quizSubmission) { input.Answers = input.Answers[:9] },
		"duplicate question":          func(input *quizSubmission) { input.Questions[1].ID = input.Questions[0].ID },
		"duplicate answer":            func(input *quizSubmission) { input.Answers[1] = input.Answers[0] },
		"unknown answer":              func(input *quizSubmission) { input.Answers[0].QuestionID = "other" },
		"unoffered selection":         func(input *quizSubmission) { input.Answers[0].SelectedAnswer = "not a choice" },
		"unoffered answer key":        func(input *quizSubmission) { input.Questions[0].Answer = "not a choice" },
		"duplicate choice":            func(input *quizSubmission) { input.Questions[0].Choices[1] = " WRITE " },
		"mixed question modes":        func(input *quizSubmission) { input.Questions[0].Mode = "reading" },
		"empty example":               func(input *quizSubmission) { input.Questions[0].Example.Reading = "" },
		"bad start date":              func(input *quizSubmission) { input.StartedAt = "yesterday" },
		"completion before start":     func(input *quizSubmission) { input.CompletedAt = "2020-01-01T00:00:00Z" },
		"full quiz references parent": func(input *quizSubmission) { parent := "parent"; input.ParentAttemptID = &parent },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			store := quizStore{db: testDatabase(t)}
			input := quizFixture("meaning", "full", 10)
			mutate(&input)
			response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
			if response.Code != 400 {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			var count int
			if err := store.db.QueryRow("SELECT COUNT(*) FROM quiz_attempts").Scan(&count); err != nil || count != 0 {
				t.Fatal("invalid attempt saved")
			}
		})
	}
	store := quizStore{db: testDatabase(t)}
	for _, body := range []string{"", `{`, `null`, `{} {}`, `{"id":"a","correctCount":10}`} {
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", body); response.Code != 400 {
			t.Fatalf("invalid JSON accepted: %s", body)
		}
	}
}

func TestQuizRetrySnapshotsPersistAcrossReopening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saved.sqlite")
	db, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := quizStore{db: db}
	input := quizFixture("reading", "retry", 3)
	parent := "source-run"
	input.ParentAttemptID = &parent
	response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
	if response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	var first savedQuizAttempt
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if first.Kind != "retry" || first.Mode != "reading" || first.Total != 3 || first.CorrectCount != 2 {
		t.Fatalf("wrong retry: %+v", first)
	}
	input.Questions[0].Example.Meaning = "a later pool edit"
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	store.db = reopened
	response = quizRequest(t, &store, "GET", "/quiz/attempts/attempt-1", nil)
	var saved savedQuizAttempt
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, saved) {
		t.Fatal("saved question/answer/word snapshot changed across restart")
	}
	if saved.Result.Items[0].Question.Example.Meaning != "write" || saved.ParentAttemptID == nil || *saved.ParentAttemptID != parent {
		t.Fatal("lost historical metadata")
	}
	if response := quizRequest(t, &store, "GET", "/quiz/attempts/not-found", nil); response.Code != 404 {
		t.Fatal("missing attempt should be 404")
	}
}

func TestQuizHistoryIsBoundedAndSeparatedByMode(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	for index := 0; index < 25; index++ {
		input := quizFixture("meaning", "full", 10)
		input.ID = fmt.Sprintf("meaning-%02d", index)
		input.CompletedAt = fmt.Sprintf("2026-11-01T10:01:%02dZ", index)
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
	}
	input := quizFixture("reading", "retry", 1)
	input.ID = "reading-1"
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	response := quizRequest(t, &store, "GET", "/quiz/attempts?mode=meaning", nil)
	var history struct {
		Attempts []quizAttemptSummary `json:"attempts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history.Attempts) != 20 || history.Attempts[0].ID != "meaning-24" || history.Attempts[19].ID != "meaning-05" {
		t.Fatalf("wrong history window: %+v", history)
	}
	if strings.Contains(response.Body.String(), `"result"`) {
		t.Fatal("summary list should not transfer full question snapshots")
	}
	response = quizRequest(t, &store, "GET", "/quiz/attempts?mode=reading", nil)
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history.Attempts) != 1 || history.Attempts[0].ID != "reading-1" || history.Attempts[0].Kind != "retry" {
		t.Fatal("mixed quiz modes in history")
	}
	if response := quizRequest(t, &store, "GET", "/quiz/attempts?mode=unknown", nil); response.Code != 400 {
		t.Fatal("invalid mode should fail")
	}
}

func TestQuizHistoryOrdersInstantsRatherThanTimestampStrings(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	for id, timestamp := range map[string]string{
		"earlier":    "2026-11-01T10:00:01Z",
		"fractional": "2026-11-01T10:00:01.001Z",
		"offset":     "2026-11-01T09:30:00-01:00",
	} {
		input := quizFixture("reading", "retry", 1)
		input.ID = id
		input.StartedAt = "2026-11-01T09:00:00Z"
		input.CompletedAt = timestamp
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
	}
	response := quizRequest(t, &store, "GET", "/quiz/attempts?mode=reading", nil)
	var history struct {
		Attempts []quizAttemptSummary `json:"attempts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history.Attempts) != 3 || history.Attempts[0].ID != "offset" || history.Attempts[1].ID != "fractional" || history.Attempts[2].ID != "earlier" {
		t.Fatalf("history order depends on timestamp spelling: %+v", history.Attempts)
	}
	if history.Attempts[0].CompletedAt != "2026-11-01T10:30:00Z" {
		t.Fatal("timestamps should be normalized to UTC")
	}
}

func TestQuizDatabaseFailureAndCancellation(t *testing.T) {
	closed := quizStore{db: testDatabase(t)}
	closed.db.Close()
	for _, method := range []string{"GET", "POST"} {
		var input any
		if method == "POST" {
			input = quizFixture("meaning", "full", 10)
		}
		if response := quizRequest(t, &closed, method, "/quiz/attempts", input); response.Code != 500 {
			t.Fatalf("closed database should report an error: %d", response.Code)
		}
	}
	store := quizStore{db: testDatabase(t)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for name, handler := range map[string]http.HandlerFunc{"history": store.listAttempts, "detail": store.getAttempt} {
		request := httptest.NewRequest("GET", "/quiz/attempts", nil).WithContext(ctx)
		request.SetPathValue("id", "attempt-1")
		response := httptest.NewRecorder()
		handler(response, request)
		if response.Body.Len() != 0 {
			t.Errorf("cancelled %s request wrote an error: %s", name, response.Body.String())
		}
	}
}

func TestQuizReadingSnapshotMustMatchItsPromptAndAnswer(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	input := quizFixture("reading", "retry", 1)
	input.Questions[0].Example.Reading = "よむ"
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 400 {
		t.Fatal("a mismatched word-reading snapshot was accepted")
	}
}
