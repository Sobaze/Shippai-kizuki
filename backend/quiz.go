package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type quizStore struct {
	db *sql.DB
}

type quizExample struct {
	Written string `json:"written"`
	Reading string `json:"reading"`
	Meaning string `json:"meaning"`
}

type quizQuestion struct {
	ID             string               `json:"id"`
	Mode           string               `json:"mode"`
	Prompt         string               `json:"prompt"`
	Answer         string               `json:"answer"`
	Choices        []string             `json:"choices"`
	Example        quizExample          `json:"example"`
	EstimatedLevel string               `json:"estimatedLevel,omitempty"`
	ExerciseID     string               `json:"exerciseId,omitempty"`
	Task           string               `json:"task,omitempty"`
	Context        *quizSentenceContext `json:"context,omitempty"`
}

type quizSentenceContext struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

func exerciseLength(id string) int {
	var catalog quizCatalog
	if json.Unmarshal(exerciseQuizData, &catalog) != nil {
		return 0
	}
	for _, pool := range catalog.Pools {
		if pool.ExerciseID == id {
			return len(pool.Questions)
		}
	}
	return 0
}

type quizAnswer struct {
	QuestionID     string `json:"questionId"`
	SelectedAnswer string `json:"selectedAnswer"`
}

type quizSubmission struct {
	ID              string         `json:"id"`
	Mode            string         `json:"mode"`
	Kind            string         `json:"kind"`
	ParentAttemptID *string        `json:"parentAttemptId,omitempty"`
	StartedAt       string         `json:"startedAt"`
	CompletedAt     string         `json:"completedAt"`
	Questions       []quizQuestion `json:"questions"`
	Answers         []quizAnswer   `json:"answers"`
}

type quizResultItem struct {
	Question       quizQuestion `json:"question"`
	SelectedAnswer string       `json:"selectedAnswer"`
	Correct        bool         `json:"correct"`
}

type quizResult struct {
	CorrectCount int              `json:"correctCount"`
	Total        int              `json:"total"`
	Items        []quizResultItem `json:"items"`
}

type quizAttemptSummary struct {
	ID           string `json:"id"`
	Mode         string `json:"mode"`
	Kind         string `json:"kind"`
	CompletedAt  string `json:"completedAt"`
	CorrectCount int    `json:"correctCount"`
	Total        int    `json:"total"`
	ExerciseID   string `json:"exerciseId,omitempty"`
}

func resultExerciseID(result quizResult) string {
	if len(result.Items) == 0 {
		return ""
	}
	id := result.Items[0].Question.ExerciseID
	for _, item := range result.Items {
		if item.Question.ExerciseID != id {
			return ""
		}
	}
	return id
}

type savedQuizAttempt struct {
	quizAttemptSummary
	StartedAt       string     `json:"startedAt"`
	ParentAttemptID *string    `json:"parentAttemptId,omitempty"`
	Result          quizResult `json:"result"`
}

var (
	quizAttemptID          = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)
	errQuizAttemptConflict = errors.New("quiz attempt contents conflict")
)

func validQuizMode(mode string) bool {
	return mode == "meaning" || mode == "reading" || mode == "sentence"
}

// The client supplies the exact question snapshots. Go validates completion and
// calculates correctness itself; scores and correctness flags are not accepted.
func validateQuizSubmission(input *quizSubmission) (quizResult, error) {
	if !quizAttemptID.MatchString(input.ID) || !validQuizMode(input.Mode) {
		return quizResult{}, errors.New("Provide a valid attempt ID and quiz mode.")
	}
	if input.Kind != "full" && input.Kind != "retry" {
		return quizResult{}, errors.New("Attempt kind must be full or retry.")
	}
	if input.ParentAttemptID != nil && (!quizAttemptID.MatchString(*input.ParentAttemptID) || *input.ParentAttemptID == input.ID || input.Kind != "retry") {
		return quizResult{}, errors.New("Only a retry may reference a different source attempt.")
	}
	started, err := time.Parse(time.RFC3339Nano, input.StartedAt)
	if err != nil {
		return quizResult{}, errors.New("Provide a valid start timestamp.")
	}
	completed, err := time.Parse(time.RFC3339Nano, input.CompletedAt)
	if err != nil || completed.Before(started) {
		return quizResult{}, errors.New("Completion must be a valid timestamp at or after the start.")
	}
	input.StartedAt = started.UTC().Format(time.RFC3339Nano)
	input.CompletedAt = completed.UTC().Format(time.RFC3339Nano)
	count := len(input.Questions)
	maximum, expected := 10, 10
	if input.Mode == "sentence" && count > 0 {
		maximum, expected = 12, exerciseLength(input.Questions[0].ExerciseID)
	}
	if count < 1 || count > maximum || (input.Kind == "full" && (expected == 0 || count != expected)) || len(input.Answers) != count {
		return quizResult{}, errors.New("Answer every question in the full quiz or exercise; retries must have 1–10 questions (1–12 for exercises).")
	}
	answers := make(map[string]string, count)
	for _, answer := range input.Answers {
		if _, exists := answers[answer.QuestionID]; exists {
			return quizResult{}, errors.New("Each question must have exactly one answer.")
		}
		answers[answer.QuestionID] = answer.SelectedAnswer
	}
	result := quizResult{Total: count, Items: make([]quizResultItem, 0, count)}
	seen := make(map[string]bool, count)
	for _, question := range input.Questions {
		if strings.TrimSpace(question.ID) == "" || seen[question.ID] || question.Mode != input.Mode || strings.TrimSpace(question.Prompt) == "" || len(question.Choices) != 4 {
			return quizResult{}, errors.New("Questions need unique IDs, the attempt's mode, a prompt, and four choices.")
		}
		seen[question.ID] = true
		if input.Mode == "sentence" {
			if exerciseLength(question.ExerciseID) == 0 || question.Context == nil || strings.TrimSpace(question.Context.Before+question.Context.After) == "" || (question.Task != "reading" && question.Task != "spelling") || (input.Kind == "full" && question.ExerciseID != input.Questions[0].ExerciseID) {
				return quizResult{}, errors.New("Sentence questions need a known exercise, reading/spelling task, and sentence context.")
			}
			if (question.Task == "reading" && (question.Prompt != question.Example.Written || question.Answer != question.Example.Reading)) || (question.Task == "spelling" && (question.Prompt != question.Example.Reading || question.Answer != question.Example.Written)) {
				return quizResult{}, errors.New("Sentence targets and answers must match their word/reading snapshots.")
			}
		} else if question.Task != "" || question.Context != nil || question.ExerciseID != "" {
			return quizResult{}, errors.New("Sentence metadata belongs only to sentence exercises.")
		}
		if question.EstimatedLevel != "" && !validQuizLevel(question.EstimatedLevel) {
			return quizResult{}, errors.New("Question levels must be estimated N2, N4, or N5 when provided.")
		}
		if strings.TrimSpace(question.Example.Written) == "" || strings.TrimSpace(question.Example.Reading) == "" || strings.TrimSpace(question.Example.Meaning) == "" {
			return quizResult{}, errors.New("Each question needs a written-word, reading, and meaning snapshot.")
		}
		if input.Mode == "reading" && (question.Prompt != question.Example.Written || question.Answer != question.Example.Reading) {
			return quizResult{}, errors.New("A reading question must match its word and reading snapshot.")
		}
		selected, answered := answers[question.ID]
		choices := make(map[string]bool, 4)
		hasAnswer, hasSelection := false, false
		for _, choice := range question.Choices {
			key := strings.ToLower(strings.TrimSpace(choice))
			if key == "" || choices[key] {
				return quizResult{}, errors.New("Choices must be non-empty and distinct.")
			}
			choices[key] = true
			hasAnswer = hasAnswer || choice == question.Answer
			hasSelection = hasSelection || choice == selected
		}
		if !answered || !hasAnswer || !hasSelection {
			return quizResult{}, errors.New("The answer key and each selected answer must be offered choices.")
		}
		correct := selected == question.Answer
		if correct {
			result.CorrectCount++
		}
		result.Items = append(result.Items, quizResultItem{Question: question, SelectedAnswer: selected, Correct: correct})
	}
	return result, nil
}

func (s *quizStore) createAttempt(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input quizSubmission
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Provide a completed quiz as JSON.", http.StatusBadRequest)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "Provide exactly one JSON object.", http.StatusBadRequest)
		return
	}
	result, err := validateQuizSubmission(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	attempt, created, err := s.saveAttempt(r.Context(), input, result)
	if errors.Is(err, errQuizAttemptConflict) {
		http.Error(w, "This attempt ID is already saved with different contents.", http.StatusConflict)
		return
	}
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("save quiz attempt: %v", err)
		http.Error(w, "Could not save this quiz. Your results are still available here; retry saving.", 500)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(attempt); err != nil {
		log.Printf("write quiz attempt: %v", err)
	}
}

func (s *quizStore) getAttempt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !quizAttemptID.MatchString(id) {
		http.Error(w, "Invalid attempt ID.", 400)
		return
	}
	attempt, err := s.loadAttempt(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Saved quiz attempt was not found.", 404)
		return
	}
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("load quiz attempt: %v", err)
		http.Error(w, "Could not load the saved attempt.", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(attempt); err != nil {
		log.Printf("write saved quiz: %v", err)
	}
}

func (s *quizStore) listAttempts(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "" && !validQuizMode(mode) {
		http.Error(w, "Mode must be meaning, reading, or sentence.", 400)
		return
	}
	attempts, err := s.recentAttempts(r.Context(), mode)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("list quiz attempts: %v", err)
		http.Error(w, "Could not load quiz history.", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(struct {
		Attempts []quizAttemptSummary `json:"attempts"`
	}{attempts}); err != nil {
		log.Printf("write quiz history: %v", err)
	}
}
