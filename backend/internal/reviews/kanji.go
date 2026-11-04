package reviews

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"unicode"
)

const minimumWordsForKanjiAdvice = 2

type kanjiEvidence struct {
	Vocabulary  string `json:"vocabulary"`
	Reading     string `json:"reading,omitempty"`
	Difficulty  int    `json:"difficulty"`
	ReviewCount int    `json:"reviewCount"`
	Qualifies   bool   `json:"qualifies"`
}

type kanjiStats struct {
	Kanji              string          `json:"kanji"`
	Difficulty         *int            `json:"difficulty"`
	WordCount          int             `json:"wordCount"`
	QualifiedWordCount int             `json:"qualifiedWordCount"`
	Recommendation     string          `json:"recommendation"`
	Evidence           []kanjiEvidence `json:"evidence"`
}

func calculateKanjiStats(words []vocabularyStats, mapping map[string][]string) []kanjiStats {
	byKanji := make(map[string]*kanjiStats)
	for _, word := range words {
		// A repeated character in a word must not count as another context.
		seen := make(map[string]bool)
		for _, character := range mapping[word.Vocabulary] {
			if seen[character] {
				continue
			}
			seen[character] = true
			stats, exists := byKanji[character]
			if !exists {
				stats = &kanjiStats{Kanji: character}
				byKanji[character] = stats
			}
			stats.Evidence = append(stats.Evidence, kanjiEvidence{
				Vocabulary:  word.Vocabulary,
				Difficulty:  word.Difficulty,
				ReviewCount: word.ReviewCount,
				Qualifies:   word.ReviewCount >= minimumReviewsForAdvice,
			})
		}
	}

	results := make([]kanjiStats, 0, len(byKanji))
	for _, stats := range byKanji {
		stats.WordCount = len(stats.Evidence)
		total := 0
		for _, evidence := range stats.Evidence {
			if evidence.Qualifies {
				total += evidence.Difficulty
				stats.QualifiedWordCount++
			}
		}
		if stats.QualifiedWordCount > 0 {
			score := int(math.Round(float64(total) / float64(stats.QualifiedWordCount)))
			stats.Difficulty = &score
		}

		switch {
		case stats.QualifiedWordCount == 0:
			stats.Recommendation = "limited_review_evidence"
		case stats.QualifiedWordCount < minimumWordsForKanjiAdvice:
			stats.Recommendation = "limited_cross_word_evidence"
		case *stats.Difficulty >= practiceThreshold:
			stats.Recommendation = "suggested_practice"
		default:
			stats.Recommendation = "no_extra_practice"
		}
		sort.Slice(stats.Evidence, func(i, j int) bool {
			return stats.Evidence[i].Vocabulary < stats.Evidence[j].Vocabulary
		})
		results = append(results, *stats)
	}

	// Scored kanji come first; unscored kanji follow. Break ties by character.
	sort.Slice(results, func(i, j int) bool {
		left, right := -1, -1
		if results[i].Difficulty != nil {
			left = *results[i].Difficulty
		}
		if results[j].Difficulty != nil {
			right = *results[j].Difficulty
		}
		if left == right {
			return results[i].Kanji < results[j].Kanji
		}
		return left > right
	})
	return results
}

func (s *reviewStore) listKanjiStats(w http.ResponseWriter, r *http.Request) {
	words, readings, err := s.loadKanjiVocabulary(r.Context())
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("load kanji reviews: %v", err)
		http.Error(w, "Could not load saved kanji reviews.", 500)
		return
	}

	stats := calculateKanjiStats(words, kanjiMappingForWords(words))
	for i := range stats {
		for j := range stats[i].Evidence {
			word := &stats[i].Evidence[j]
			word.Reading = readings[word.Vocabulary]
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("write kanji stats response: %v", err)
	}
}

// Keep counts and readings in one read snapshot so an import cannot change
// the collection between these queries. No network requests run in this transaction.
func (s *reviewStore) loadKanjiVocabulary(ctx context.Context) ([]vocabularyStats, map[string]string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, vocabularyCountsSQL)
	if err != nil {
		return nil, nil, fmt.Errorf("read vocabulary counts: %w", err)
	}
	words, err := readVocabularyCounts(rows)
	if err != nil {
		return nil, nil, fmt.Errorf("decode vocabulary counts: %w", err)
	}
	// Conflicting readings of the same written word are not guessed or merged.
	rows, err = tx.QueryContext(ctx, `SELECT vocabulary,
		CASE WHEN COUNT(DISTINCT NULLIF(reading, '')) = 1 THEN MAX(reading) ELSE '' END
		FROM anki_cards GROUP BY vocabulary`)
	if err != nil {
		return nil, nil, fmt.Errorf("read vocabulary readings: %w", err)
	}
	readings := make(map[string]string)
	for rows.Next() {
		var word, reading string
		if err = rows.Scan(&word, &reading); err != nil {
			break
		}
		readings[word] = reading
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("decode vocabulary readings: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return scoreVocabularyStats(words), readings, nil
}

func kanjiMappingForWords(words []vocabularyStats) map[string][]string {
	mapping := make(map[string][]string)
	for _, entry := range words {
		characters := []string{}
		for _, character := range entry.Vocabulary {
			if unicode.Is(unicode.Han, character) && character != '々' {
				characters = append(characters, string(character))
			}
		}
		mapping[entry.Vocabulary] = characters
	}
	return mapping
}
