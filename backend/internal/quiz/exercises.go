package quiz

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Embedded exercise definitions are immutable for the lifetime of the server.
// Decode their lengths once, rather than decoding every passage for each answer.
var exerciseLengths = sync.OnceValue(func() map[string]int {
	lengths := map[string]int{}
	for _, data := range [][]byte{exerciseQuizData, additionalExerciseQuizData, originalQuizData} {
		var catalog quizCatalog
		if json.Unmarshal(data, &catalog) != nil {
			return nil
		}
		for _, pool := range catalog.Pools {
			lengths[pool.ExerciseID] = len(pool.Questions)
		}
	}
	return lengths
})

func exerciseLength(id string) int { return exerciseLengths()[id] }

type exerciseSummary struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Level              string `json:"level"`
	QuestionCount      int    `json:"questionCount"`
	ReadingCount       int    `json:"readingCount"`
	SpellingCount      int    `json:"spellingCount"`
	ComprehensionCount int    `json:"comprehensionCount"`
	Category           string `json:"category"`
	Original           bool   `json:"original,omitempty"`
}

func poolCategory(pool quizPool) string {
	if pool.Category == "" {
		return "kanji"
	}
	return pool.Category
}

func validExerciseCategory(category string) bool {
	return category == "kanji" || category == "comprehension"
}

func (c *quizCatalog) listExercises(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	category := r.URL.Query().Get("category")
	if !validQuizLevel(level) {
		http.Error(w, "Choose a supported exercise level.", 400)
		return
	}
	if category != "" && !validExerciseCategory(category) {
		http.Error(w, "Choose kanji or comprehension practice.", 400)
		return
	}
	exercises := []exerciseSummary{}
	for _, pool := range c.Pools {
		if pool.Mode != "sentence" || pool.Level != level || (category != "" && poolCategory(pool) != category) {
			continue
		}
		summary := exerciseSummary{ID: pool.ExerciseID, Title: pool.Title, Level: pool.Level, QuestionCount: len(pool.Questions), Category: poolCategory(pool), Original: pool.Original}
		for _, question := range pool.Questions {
			if question.Task == "reading" {
				summary.ReadingCount++
			} else if question.Task == "spelling" {
				summary.SpellingCount++
			} else if question.Task == "comprehension" {
				summary.ComprehensionCount++
			}
		}
		exercises = append(exercises, summary)
	}
	sort.Slice(exercises, func(i, j int) bool {
		if exercises[i].Original != exercises[j].Original {
			return exercises[i].Original
		}
		first, firstErr := strconv.Atoi(strings.TrimPrefix(exercises[i].ID, "exercise-"))
		second, secondErr := strconv.Atoi(strings.TrimPrefix(exercises[j].ID, "exercise-"))
		if firstErr == nil && secondErr == nil {
			return first < second
		}
		return exercises[i].ID < exercises[j].ID
	})
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(struct {
		Exercises []exerciseSummary `json:"exercises"`
	}{exercises}); err != nil {
		log.Printf("write exercise list: %v", err)
	}
}
