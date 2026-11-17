package quiz

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestOriginalN2CatalogAndCategoryLists(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData, n2QuizData, exerciseQuizData, additionalExerciseQuizData, originalQuizData)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		category         string
		count, originals int
	}{{"kanji", 13, 2}, {"comprehension", 3, 3}} {
		response := httptest.NewRecorder()
		catalog.listExercises(response, httptest.NewRequest("GET", "/quiz/exercises?level=N2&category="+tc.category, nil))
		var result struct {
			Exercises []exerciseSummary `json:"exercises"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if response.Code != 200 || len(result.Exercises) != tc.count {
			t.Fatalf("category list: %s", response.Body.String())
		}
		for i, summary := range result.Exercises {
			if summary.Category != tc.category || (i < tc.originals && !summary.Original) {
				t.Fatalf("category/order mismatch: %+v", summary)
			}
			if summary.ReadingCount+summary.SpellingCount+summary.ComprehensionCount != summary.QuestionCount {
				t.Fatal("task counts do not match set size")
			}
		}
	}
	response := httptest.NewRecorder()
	catalog.listExercises(response, httptest.NewRequest("GET", "/quiz/exercises?level=N2&category=unknown", nil))
	if response.Code != 400 {
		t.Fatal("unknown category accepted")
	}
}

func TestOriginalSetsSaveGradeAndPreserveExplanations(t *testing.T) {
	catalog, err := loadQuizCatalog(originalQuizData)
	if err != nil {
		t.Fatal(err)
	}
	store := quizStore{db: testDatabase(t)}
	for _, pool := range catalog.Pools {
		count := 6
		if pool.Category == "comprehension" {
			count = 2
		}
		if len(pool.Questions) != count || !pool.Original {
			t.Fatal("unexpected original pilot size")
		}
		input := quizSubmission{ID: pool.ExerciseID, Mode: "sentence", Kind: "full", StartedAt: "2026-11-10T10:00:00Z", CompletedAt: "2026-11-10T10:01:00Z", Questions: pool.Questions}
		for _, question := range input.Questions {
			input.Answers = append(input.Answers, quizAnswer{QuestionID: question.ID, SelectedAnswer: question.Answer})
		}
		for _, status := range []int{201, 200} {
			if response := quizRequest(t, &store, "POST", "/quiz/attempts", input); response.Code != status {
				t.Fatal(response.Body.String())
			}
		}
		saved, err := store.loadAttempt(context.Background(), input.ID)
		if err != nil || saved.Total != count || saved.CorrectCount != count {
			t.Fatalf("original result: %+v %v", saved, err)
		}
		for i, item := range saved.Result.Items {
			if item.Question.Explanation == "" || item.Question.Explanation != input.Questions[i].Explanation {
				t.Fatal("explanation snapshot lost")
			}
			if pool.Category == "comprehension" && (item.Question.Passage == nil || item.Question.Passage.Text != input.Questions[i].Passage.Text || item.Question.Evidence != input.Questions[i].Evidence) {
				t.Fatal("passage/evidence snapshot lost")
			}
		}
		coverage, err := store.poolCoverage(context.Background(), pool)
		if err != nil || coverage.Attempted != count {
			t.Fatal("original-set coverage broken")
		}
		incomplete := input
		incomplete.Questions, incomplete.Answers = input.Questions[:count-1], input.Answers[:count-1]
		if _, err := validateQuizSubmission(&incomplete); err == nil {
			t.Fatal("incomplete original set accepted")
		}
	}
}

func TestComprehensionRequiresGroundedPassageFeedback(t *testing.T) {
	for name, mutate := range map[string]func(*quizQuestion){
		"missing passage":       func(q *quizQuestion) { q.Passage = nil },
		"blank text":            func(q *quizQuestion) { q.Passage.Text = " " },
		"missing explanation":   func(q *quizQuestion) { q.Explanation = "" },
		"ungrounded evidence":   func(q *quizQuestion) { q.Evidence = "not in the passage" },
		"word context mixed in": func(q *quizQuestion) { q.Context = &quizSentenceContext{Before: "before", After: "after"} },
		"word example mixed in": func(q *quizQuestion) { q.Example = quizExample{Written: "海", Reading: "うみ", Meaning: "sea"} },
	} {
		t.Run(name, func(t *testing.T) {
			catalog, err := loadQuizCatalog(originalQuizData)
			if err != nil {
				t.Fatal(err)
			}
			question := catalog.Pools[2].Questions[0]
			mutate(&question)
			input := quizSubmission{ID: "bad-comprehension", Mode: "sentence", Kind: "retry", StartedAt: "2026-11-10T10:00:00Z", CompletedAt: "2026-11-10T10:01:00Z", Questions: []quizQuestion{question}, Answers: []quizAnswer{{QuestionID: question.ID, SelectedAnswer: question.Answer}}}
			if _, err := validateQuizSubmission(&input); err == nil {
				t.Fatal("malformed comprehension accepted")
			}
		})
	}
}
