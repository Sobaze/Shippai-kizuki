package quiz

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"
)

//go:embed data/questions.json
var quizData []byte

//go:embed data/n2.generated.json
var n2QuizData []byte

//go:embed data/exercises.json
var exerciseQuizData []byte

type quizPool struct {
	Mode       string         `json:"mode"`
	Level      string         `json:"level"`
	Questions  []quizQuestion `json:"questions"`
	ExerciseID string         `json:"exerciseId,omitempty"`
	Title      string         `json:"title,omitempty"`
}

type quizCatalog struct {
	Version   int               `json:"version"`
	CheckedAt string            `json:"checkedAt"`
	Sources   map[string]string `json:"sources"`
	Pools     []quizPool        `json:"pools"`
}

func validQuizLevel(level string) bool { return level == "N5" || level == "N4" || level == "N2" }

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

func loadQuizCatalog(data []byte, additions ...[]byte) (quizCatalog, error) {
	catalog := quizCatalog{Version: 1, Sources: map[string]string{}}
	for _, part := range append([][]byte{data}, additions...) {
		var source quizCatalog
		decoder := json.NewDecoder(bytes.NewReader(part))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&source); err != nil {
			return quizCatalog{}, fmt.Errorf("decode quiz data: %w", err)
		}
		if source.Version != 1 || source.CheckedAt == "" || len(source.Sources) == 0 || len(source.Pools) == 0 {
			return quizCatalog{}, fmt.Errorf("quiz data needs version 1, review date, sources, and pools")
		}
		catalog.Pools = append(catalog.Pools, source.Pools...)
		for key, value := range source.Sources {
			catalog.Sources[key] = value
		}
		if source.CheckedAt > catalog.CheckedAt {
			catalog.CheckedAt = source.CheckedAt
		}
	}
	seenPools, seenIDs, seenPrompts := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for index := range catalog.Pools {
		pool := &catalog.Pools[index]
		key := pool.Mode + "/" + pool.Level + "/" + pool.ExerciseID
		validSize := len(pool.Questions) >= 10
		if pool.Mode == "sentence" {
			validSize = pool.ExerciseID != "" && pool.Title != "" && pool.Level == "N2" && len(pool.Questions) >= 1 && len(pool.Questions) <= 12
		}
		if seenPools[key] || !validQuizMode(pool.Mode) || !validQuizLevel(pool.Level) || !validSize || (pool.Mode != "sentence" && pool.ExerciseID != "") {
			return quizCatalog{}, fmt.Errorf("invalid, duplicate, or undersized quiz pool %q", key)
		}
		seenPools[key] = true
		for i := range pool.Questions {
			question := &pool.Questions[i]
			promptKey := pool.Mode + "/" + pool.ExerciseID + "/" + question.Task + "/" + question.Prompt
			if seenIDs[question.ID] || seenPrompts[promptKey] || question.Mode != pool.Mode {
				return quizCatalog{}, fmt.Errorf("duplicate or mismatched question %q", question.ID)
			}
			seenIDs[question.ID], seenPrompts[promptKey] = true, true
			if pool.Mode == "meaning" {
				runes := []rune(question.Prompt)
				if len(runes) != 1 || !unicode.Is(unicode.Han, runes[0]) || !strings.Contains(question.Example.Written, question.Prompt) {
					return quizCatalog{}, fmt.Errorf("meaning question %q needs one kanji and an example containing it", question.ID)
				}
			}
			if question.EstimatedLevel != "" && question.EstimatedLevel != pool.Level {
				return quizCatalog{}, fmt.Errorf("question %q has a level that differs from its pool", question.ID)
			}
			question.EstimatedLevel = pool.Level
			if pool.Mode == "sentence" && question.ExerciseID != pool.ExerciseID {
				return quizCatalog{}, fmt.Errorf("question %q does not match its exercise", question.ID)
			}
			// Reuse submission validation to check choice uniqueness, answer keys,
			// reading/word alignment, and complete example snapshots at startup.
			input := quizSubmission{ID: "catalog-check", Mode: pool.Mode, Kind: "retry",
				StartedAt: "2026-01-01T00:00:00Z", CompletedAt: "2026-01-01T00:00:00Z",
				Questions: []quizQuestion{*question}, Answers: []quizAnswer{{QuestionID: question.ID, SelectedAnswer: question.Answer}}}
			if _, err := validateQuizSubmission(&input); err != nil {
				return quizCatalog{}, fmt.Errorf("question %q: %w", question.ID, err)
			}
			for _, r := range question.Example.Reading {
				if !unicode.Is(unicode.Hiragana, r) && r != 'ー' {
					return quizCatalog{}, fmt.Errorf("example reading for %q must be kana", question.ID)
				}
			}
			if pool.Mode == "reading" || (pool.Mode == "sentence" && question.Task == "reading") {
				for _, choice := range question.Choices {
					for _, r := range choice {
						if !unicode.Is(unicode.Hiragana, r) && r != 'ー' {
							return quizCatalog{}, fmt.Errorf("reading choices for %q must be kana", question.ID)
						}
					}
				}
			}
		}
	}
	return catalog, nil
}

func (c *quizCatalog) listQuestions(w http.ResponseWriter, r *http.Request) {
	mode, level := r.URL.Query().Get("mode"), r.URL.Query().Get("level")
	for _, pool := range c.Pools {
		if pool.Mode != mode || pool.Level != level || pool.ExerciseID != r.URL.Query().Get("exercise") {
			continue
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(struct {
			quizPool
			Version   int               `json:"version"`
			CheckedAt string            `json:"checkedAt"`
			Sources   map[string]string `json:"sources"`
		}{pool, c.Version, c.CheckedAt, c.Sources}); err != nil {
			log.Printf("write quiz pool: %v", err)
		}
		return
	}
	http.Error(w, "Choose a supported mode and level; sentence exercises also need an exercise ID.", http.StatusBadRequest)
}
