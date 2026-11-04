package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestPracticeQueueReplay(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	step := 0
	save := func(mode, question string, correct bool) quizSubmission {
		t.Helper()
		step++
		input := quizFixture(mode, "retry", 1)
		input.ID = fmt.Sprintf("attempt-%03d", step)
		input.CompletedAt = time.Date(2026, 11, 1, 12, 0, step, 0, time.UTC).Format(time.RFC3339Nano)
		input.Questions[0].ID = question
		input.Answers[0].QuestionID = question
		if !correct {
			input.Answers[0].SelectedAnswer = input.Questions[0].Choices[1]
		}
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatalf("save: %s", response.Body.String())
		}
		return input
	}
	check := func(mode string, want map[string]int) {
		t.Helper()
		response := quizRequest(t, &store, "GET", "/quiz/practice?mode="+mode, nil)
		var body struct {
			Candidates []quizCandidate `json:"candidates"`
		}
		if response.Code != 200 {
			t.Fatalf("queue: %s", response.Body.String())
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Candidates == nil || len(body.Candidates) != len(want) {
			t.Fatalf("queue = %+v; want %v", body.Candidates, want)
		}
		for _, candidate := range body.Candidates {
			streak, ok := want[candidate.Question.ID]
			if !ok || streak != candidate.Streak {
				t.Fatalf("unexpected candidate: %+v", candidate)
			}
		}
	}
	check("meaning", map[string]int{})
	save("meaning", "a", true) // Correct-only questions never enter the queue.
	check("meaning", map[string]int{})
	save("meaning", "a", false)
	check("meaning", map[string]int{"a": 0})
	firstCorrect := save("meaning", "a", true)
	check("meaning", map[string]int{"a": 1})
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", firstCorrect); response.Code != 200 {
		t.Fatal("idempotent save failed")
	}
	check("meaning", map[string]int{"a": 1})
	save("reading", "a", false)
	for range 25 {
		save("meaning", "unrelated", true)
	}
	check("meaning", map[string]int{"a": 1}) // Replay includes more than 20 attempts; unasked items are unchanged.
	check("reading", map[string]int{"a": 0})
	save("meaning", "a", false)
	check("meaning", map[string]int{"a": 0})
	save("meaning", "a", true)
	save("meaning", "a", true)
	check("meaning", map[string]int{})
	save("meaning", "a", false)
	check("meaning", map[string]int{"a": 0})
	save("meaning", "b", false)
	candidates, err := store.practiceCandidates(context.Background(), "meaning")
	if err != nil || len(candidates) != 2 || candidates[0].Question.ID != "a" {
		t.Fatalf("oldest first: %+v, %v", candidates, err)
	}
	if response := quizRequest(t, &store, "GET", "/quiz/practice?mode=invalid", nil); response.Code != 400 {
		t.Fatal("invalid mode accepted")
	}
}

func TestPracticeQueueChronologyAndFullQuiz(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	// Save out of completion order. Sub-millisecond precision must be preserved.
	for _, entry := range []struct {
		id, timestamp string
		correct       bool
	}{
		{"c", "2026-11-01T12:00:00.000000003Z", true},
		{"a", "2026-11-01T12:00:00.000000001Z", false},
		{"b", "2026-11-01T12:00:00.000000002Z", true},
	} {
		input := quizFixture("meaning", "full", 10)
		input.ID, input.CompletedAt = entry.id, entry.timestamp
		for i := range input.Answers {
			input.Answers[i].SelectedAnswer = input.Questions[i].Answer
			if !entry.correct {
				input.Answers[i].SelectedAnswer = input.Questions[i].Choices[1]
			}
		}
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatalf("save: %s", response.Body.String())
		}
	}
	candidates, err := store.practiceCandidates(context.Background(), "meaning")
	if err != nil || len(candidates) != 0 {
		t.Fatalf("chronological replay: %+v %v", candidates, err)
	}
	// Exact-time ties use ID, not insertion order.
	if _, err := store.db.Exec(`UPDATE quiz_attempts SET completed_at = '2026-11-01T12:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	candidates, err = store.practiceCandidates(context.Background(), "meaning")
	if err != nil || len(candidates) != 0 {
		t.Fatalf("tie replay: %+v %v", candidates, err)
	}
}
