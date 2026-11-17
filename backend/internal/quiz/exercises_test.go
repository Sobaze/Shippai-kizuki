package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExerciseListAndAllSetLengths(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData, n2QuizData, exerciseQuizData, additionalExerciseQuizData)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Pools) != 17 {
		t.Fatalf("got %d pools", len(catalog.Pools))
	}
	response := httptest.NewRecorder()
	catalog.listExercises(response, httptest.NewRequest("GET", "/quiz/exercises?level=N2", nil))
	var body struct {
		Exercises []exerciseSummary `json:"exercises"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(body.Exercises) != 11 {
		t.Fatalf("exercise list: %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), `"choices"`) || strings.Contains(response.Body.String(), `"answer"`) {
		t.Fatal("list endpoint should contain metadata only")
	}
	total, readings, spellings := 0, 0, 0
	store := quizStore{db: testDatabase(t)}
	for index, summary := range body.Exercises {
		id := fmt.Sprintf("exercise-%d", index+1)
		wantLength := 12
		if index >= 6 {
			wantLength = 10
		}
		if summary.ID != id || summary.QuestionCount != wantLength || summary.ReadingCount+summary.SpellingCount != wantLength || exerciseLength(id) != wantLength {
			t.Fatalf("metadata/ordering mismatch: %+v", summary)
		}
		total += summary.QuestionCount
		readings += summary.ReadingCount
		spellings += summary.SpellingCount
		var pool quizPool
		for _, candidate := range catalog.Pools {
			if candidate.ExerciseID == id {
				pool = candidate
			}
		}
		input := quizSubmission{ID: "full-" + id, Mode: "sentence", Kind: "full", StartedAt: "2026-11-01T10:00:00Z", CompletedAt: "2026-11-01T10:01:00Z", Questions: pool.Questions}
		for _, question := range pool.Questions {
			input.Answers = append(input.Answers, quizAnswer{QuestionID: question.ID, SelectedAnswer: question.Answer})
		}
		for _, status := range []int{201, 200} {
			if saved := quizRequest(t, &store, "POST", "/quiz/attempts", input); saved.Code != status {
				t.Fatalf("save %s: %s", id, saved.Body.String())
			}
		}
		coverage, err := store.poolCoverage(context.Background(), pool)
		if err != nil || coverage.ExerciseID != id || coverage.Attempted != wantLength || coverage.Total != wantLength {
			t.Fatalf("coverage %s: %+v %v", id, coverage, err)
		}
		for _, entry := range coverage.Questions {
			if entry.Attempts != 1 {
				t.Fatal("duplicate save inflated exercise progress")
			}
		}
		input.Questions, input.Answers = input.Questions[:wantLength-1], input.Answers[:wantLength-1]
		if _, err := validateQuizSubmission(&input); err == nil {
			t.Fatalf("incomplete %s accepted", id)
		}
	}
	if total != 122 || readings != 68 || spellings != 54 {
		t.Fatalf("incorrect bank totals: %d %d %d", total, readings, spellings)
	}
	response = httptest.NewRecorder()
	catalog.listExercises(response, httptest.NewRequest("GET", "/quiz/exercises?level=unknown", nil))
	if response.Code != 400 {
		t.Fatal("invalid level accepted")
	}
}

func TestExerciseStreaksAreIndependentAcrossSets(t *testing.T) {
	catalog, err := loadQuizCatalog(exerciseQuizData, additionalExerciseQuizData)
	if err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: testDatabase(t)}
	save := func(exerciseID, id, completed string, correct bool) {
		t.Helper()
		var questions []quizQuestion
		for _, pool := range catalog.Pools {
			if pool.ExerciseID == exerciseID {
				questions = pool.Questions
			}
		}
		input := quizSubmission{ID: id, Mode: "sentence", Kind: "full", StartedAt: "2026-11-01T10:00:00Z", CompletedAt: completed, Questions: questions}
		for _, question := range questions {
			answer := question.Answer
			if !correct {
				for _, choice := range question.Choices {
					if choice != question.Answer {
						answer = choice
						break
					}
				}
			}
			input.Answers = append(input.Answers, quizAnswer{QuestionID: question.ID, SelectedAnswer: answer})
		}
		if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != 201 {
			t.Fatal(response.Body.String())
		}
	}
	save("exercise-1", "wrong-one", "2026-11-01T10:01:00Z", false)
	save("exercise-10", "wrong-ten", "2026-11-01T10:02:00Z", false)
	save("exercise-1", "correct-one", "2026-11-01T11:00:00Z", true)
	save("exercise-1", "correct-one-again", "2026-11-01T12:00:00Z", true)
	candidates, err := store.practiceCandidates(context.Background(), "sentence")
	if err != nil || len(candidates) != 10 {
		t.Fatalf("cross-exercise streaks: %+v %v", candidates, err)
	}
	for _, candidate := range candidates {
		if candidate.Question.ExerciseID != "exercise-10" || candidate.Streak != 0 {
			t.Fatal("practising one exercise changed another exercise's queue")
		}
	}
}
