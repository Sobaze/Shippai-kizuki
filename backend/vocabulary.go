package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
)

const vocabularyCountsSQL = `
	SELECT vocabulary, COUNT(*),
		SUM(CASE WHEN rating = 'again' THEN 1 ELSE 0 END),
		SUM(CASE WHEN rating = 'hard' THEN 1 ELSE 0 END),
		SUM(CASE WHEN rating = 'good' THEN 1 ELSE 0 END),
		SUM(CASE WHEN rating = 'easy' THEN 1 ELSE 0 END)
	FROM review_events WHERE rating IS NOT NULL
	GROUP BY vocabulary`

const (
	againWeight             = 1.0
	hardWeight              = 0.5
	practiceThreshold       = 50
	minimumReviewsForAdvice = 3
)

type ratingCounts struct {
	Again int `json:"again"`
	Hard  int `json:"hard"`
	Good  int `json:"good"`
	Easy  int `json:"easy"`
}

type vocabularyStats struct {
	Vocabulary     string       `json:"vocabulary"`
	ReviewCount    int          `json:"reviewCount"`
	RatingCounts   ratingCounts `json:"ratingCounts"`
	Difficulty     int          `json:"difficulty"`
	Recommendation string       `json:"recommendation"`
}

func scoreVocabularyStats(results []vocabularyStats) []vocabularyStats {
	for i := range results {
		stats := &results[i]
		weightedDifficulty := float64(stats.RatingCounts.Again)*againWeight + float64(stats.RatingCounts.Hard)*hardWeight
		stats.Difficulty = int(math.Round(100 * weightedDifficulty / float64(stats.ReviewCount)))

		switch {
		case stats.ReviewCount < minimumReviewsForAdvice:
			stats.Recommendation = "limited_evidence"
		case stats.Difficulty >= practiceThreshold:
			stats.Recommendation = "suggested_practice"
		default:
			stats.Recommendation = "no_extra_practice"
		}
	}

	// Sort scores, then words to make ties deterministic.
	sort.Slice(results, func(i, j int) bool {
		if results[i].Difficulty == results[j].Difficulty {
			return results[i].Vocabulary < results[j].Vocabulary
		}
		return results[i].Difficulty > results[j].Difficulty
	})
	return results
}

// Consume and close the result before another query uses the connection.
func readVocabularyCounts(rows *sql.Rows) ([]vocabularyStats, error) {
	defer rows.Close()
	results := []vocabularyStats{}
	for rows.Next() {
		var stats vocabularyStats
		if err := rows.Scan(&stats.Vocabulary, &stats.ReviewCount, &stats.RatingCounts.Again,
			&stats.RatingCounts.Hard, &stats.RatingCounts.Good, &stats.RatingCounts.Easy); err != nil {
			return nil, err
		}
		results = append(results, stats)
	}
	return results, rows.Err()
}

func (s *reviewStore) loadVocabularyStats(ctx context.Context) ([]vocabularyStats, error) {
	rows, err := s.db.QueryContext(ctx, vocabularyCountsSQL)
	if err != nil {
		return nil, err
	}
	counts, err := readVocabularyCounts(rows)
	if err != nil {
		return nil, err
	}
	return scoreVocabularyStats(counts), nil
}

func (s *reviewStore) listVocabularyStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.loadVocabularyStats(r.Context())
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("load vocabulary reviews: %v", err)
		http.Error(w, "Could not load saved vocabulary reviews.", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("write vocabulary stats response: %v", err)
	}
}
