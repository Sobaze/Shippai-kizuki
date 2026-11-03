package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type quizQuestionCoverage struct {
	QuestionID      string  `json:"questionId"`
	Attempts        int     `json:"attempts"`
	LastPracticedAt *string `json:"lastPracticedAt"`
}

type quizCoverage struct {
	Mode       string                 `json:"mode"`
	Level      string                 `json:"level"`
	Total      int                    `json:"total"`
	Attempted  int                    `json:"attempted"`
	Unseen     int                    `json:"unseen"`
	Questions  []quizQuestionCoverage `json:"questions"`
	ExerciseID string                 `json:"exerciseId,omitempty"`
}

// Match stable IDs to the current pool, rather than relying on historical level
// metadata. This includes reclassified starter words but excludes retired items.
func (s *quizStore) poolCoverage(ctx context.Context, pool quizPool) (quizCoverage, error) {
	coverage := quizCoverage{Mode: pool.Mode, Level: pool.Level, ExerciseID: pool.ExerciseID, Total: len(pool.Questions),
		Questions: make([]quizQuestionCoverage, len(pool.Questions))}
	positions := make(map[string]int, len(pool.Questions))
	for i, question := range pool.Questions {
		coverage.Questions[i].QuestionID = question.ID
		positions[question.ID] = i
	}
	rows, err := s.db.QueryContext(ctx, `SELECT completed_at, result_json FROM quiz_attempts WHERE mode = ?`, pool.Mode)
	if err != nil {
		return quizCoverage{}, err
	}
	defer rows.Close()
	latest := make(map[string]time.Time, len(pool.Questions))
	for rows.Next() {
		var completed, encoded string
		if err := rows.Scan(&completed, &encoded); err != nil {
			return quizCoverage{}, err
		}
		at, err := time.Parse(time.RFC3339Nano, completed)
		if err != nil {
			return quizCoverage{}, err
		}
		var result quizResult
		if err := json.Unmarshal([]byte(encoded), &result); err != nil {
			return quizCoverage{}, err
		}
		for _, item := range result.Items {
			index, exists := positions[item.Question.ID]
			if !exists {
				continue
			}
			entry := &coverage.Questions[index]
			entry.Attempts++
			if entry.LastPracticedAt == nil || at.After(latest[item.Question.ID]) {
				entry.LastPracticedAt = &completed
				latest[item.Question.ID] = at
			}
		}
	}
	if err := rows.Err(); err != nil {
		return quizCoverage{}, err
	}
	if err := ctx.Err(); err != nil {
		return quizCoverage{}, err
	}
	for _, entry := range coverage.Questions {
		if entry.Attempts > 0 {
			coverage.Attempted++
		}
	}
	coverage.Unseen = coverage.Total - coverage.Attempted
	return coverage, nil
}

func (s *quizStore) listCoverage(catalog *quizCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode, level := r.URL.Query().Get("mode"), r.URL.Query().Get("level")
		for _, pool := range catalog.Pools {
			if pool.Mode != mode || pool.Level != level || pool.ExerciseID != r.URL.Query().Get("exercise") {
				continue
			}
			coverage, err := s.poolCoverage(r.Context(), pool)
			if err != nil {
				if r.Context().Err() != nil {
					return
				}
				log.Printf("load quiz coverage: %v", err)
				http.Error(w, "Could not load saved quiz coverage.", 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			if err := json.NewEncoder(w).Encode(coverage); err != nil {
				log.Printf("write quiz coverage: %v", err)
			}
			return
		}
		http.Error(w, "Choose a supported mode and level; sentence exercises also need an exercise ID.", 400)
	}
}
