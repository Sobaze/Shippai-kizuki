package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
)

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

func calculateVocabularyStats(reviews []review) []vocabularyStats {
	// A map lets every review of the same word contribute to one result.
	byWord := make(map[string]*vocabularyStats)
	for _, entry := range reviews {
		stats, exists := byWord[entry.Vocabulary]
		if !exists {
			stats = &vocabularyStats{Vocabulary: entry.Vocabulary}
			byWord[entry.Vocabulary] = stats
		}
		stats.ReviewCount++
		switch entry.Rating {
		case "again":
			stats.RatingCounts.Again++
		case "hard":
			stats.RatingCounts.Hard++
		case "good":
			stats.RatingCounts.Good++
		case "easy":
			stats.RatingCounts.Easy++
		}
	}

	results := make([]vocabularyStats, 0, len(byWord))
	for _, stats := range byWord {
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
		results = append(results, *stats)
	}

	// Map iteration has no defined order. Sort scores, then words to break ties.
	sort.Slice(results, func(i, j int) bool {
		if results[i].Difficulty == results[j].Difficulty {
			return results[i].Vocabulary < results[j].Vocabulary
		}
		return results[i].Difficulty > results[j].Difficulty
	})
	return results
}

func (s *reviewStore) listVocabularyStats(w http.ResponseWriter, r *http.Request) {
	reviews, err := s.loadReviews(r.Context(), 0)
	if err != nil {
		log.Printf("load vocabulary reviews: %v", err)
		http.Error(w, "Could not load saved vocabulary reviews.", 500)
		return
	}

	stats := calculateVocabularyStats(reviews)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("write vocabulary stats response: %v", err)
	}
}
