package quiz

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

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
