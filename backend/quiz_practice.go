package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"
)

type quizCandidate struct {
	Question        quizQuestion `json:"question"`
	Streak          int          `json:"streak"`
	LastPracticedAt string       `json:"lastPracticedAt"`
}

// Replay every saved attempt, including attempts outside the recent-history list.
// Equal completion timestamps are ordered by attempt ID for reproducible replay.
func (s *quizStore) practiceCandidates(ctx context.Context, mode string) ([]quizCandidate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, completed_at, result_json FROM quiz_attempts WHERE mode = ?`, mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type event struct {
		id, completed string
		at            time.Time
		result        quizResult
	}
	events := []event{}
	for rows.Next() {
		var e event
		var encoded string
		if err := rows.Scan(&e.id, &e.completed, &encoded); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(encoded), &e.result); err != nil {
			return nil, err
		}
		e.at, err = time.Parse(time.RFC3339Nano, e.completed)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].at.Equal(events[j].at) {
			return events[i].id < events[j].id
		}
		return events[i].at.Before(events[j].at)
	})
	active := map[string]quizCandidate{}
	for _, e := range events {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, item := range e.result.Items {
			candidate, exists := active[item.Question.ID]
			if item.Correct && !exists {
				continue
			}
			candidate.Question = item.Question
			candidate.LastPracticedAt = e.completed
			if item.Correct {
				candidate.Streak++
			} else {
				candidate.Streak = 0
			}
			if candidate.Streak >= 2 {
				delete(active, item.Question.ID)
			} else {
				active[item.Question.ID] = candidate
			}
		}
	}
	candidates := make([]quizCandidate, 0, len(active))
	for _, candidate := range active {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, candidates[i].LastPracticedAt)
		b, _ := time.Parse(time.RFC3339Nano, candidates[j].LastPracticedAt)
		if a.Equal(b) {
			return candidates[i].Question.ID < candidates[j].Question.ID
		}
		return a.Before(b)
	})
	return candidates, nil
}

func (s *quizStore) listPracticeCandidates(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if !validQuizMode(mode) {
		http.Error(w, "Mode must be meaning, reading, or sentence.", 400)
		return
	}
	candidates, err := s.practiceCandidates(r.Context(), mode)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("list quiz practice candidates: %v", err)
		http.Error(w, "Could not load practice candidates.", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(struct {
		Candidates []quizCandidate `json:"candidates"`
	}{candidates}); err != nil {
		log.Printf("write quiz practice candidates: %v", err)
	}
}
