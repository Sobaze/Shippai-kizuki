package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const vocabularyDeck = "JP Vocabulary"

type ankiField struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

type ankiCard struct {
	CardID     int64                `json:"cardId"`
	NoteID     int64                `json:"note"`
	DeckName   string               `json:"deckName"`
	ModelName  string               `json:"modelName"`
	Fields     map[string]ankiField `json:"fields"`
	Vocabulary string               `json:"-"`
	Reading    string               `json:"-"`
	AnswerText string               `json:"-"`
}

type ankiReview struct {
	ID       int64 `json:"id"`
	USN      int   `json:"usn"`
	Ease     int   `json:"ease"`
	Interval int   `json:"ivl"`
	LastIvl  int   `json:"lastIvl"`
	Factor   int   `json:"factor"`
	Time     int   `json:"time"`
	Type     int   `json:"type"`
}

type importSummary struct {
	Deck                string  `json:"deck"`
	CardCount           int     `json:"cardCount"`
	ReviewCount         int     `json:"reviewCount"`
	UnscoredReviewCount int     `json:"unscoredReviewCount"`
	MissingReadingCount int     `json:"missingReadingCount"`
	ImportedAt          *string `json:"importedAt"`
}

type ankiImporter struct {
	anki *ankiClient
	db   *sql.DB
	mu   sync.Mutex
}

func (i *ankiImporter) importVocabulary(w http.ResponseWriter, r *http.Request) {
	if !i.mu.TryLock() {
		http.Error(w, "A vocabulary import is already running. Please wait for it to finish.", http.StatusConflict)
		return
	}
	defer i.mu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cards, reviews, err := i.readVocabulary(ctx)
	if err == nil {
		err = saveVocabularySnapshot(ctx, i.db, cards, reviews)
	}
	if err != nil {
		if r.Context().Err() == nil {
			writeAnkiError(w, err)
		}
		return
	}
	i.importStatus(w, r)
}

func (i *ankiImporter) readVocabulary(ctx context.Context) ([]ankiCard, map[int64][]ankiReview, error) {
	var decks []string
	if err := i.anki.call(ctx, "deckNames", map[string]any{}, &decks); err != nil {
		return nil, nil, err
	}
	found := false
	for _, deck := range decks {
		found = found || deck == vocabularyDeck
	}
	if !found {
		return nil, nil, &ankiError{404, "JP Vocabulary was not found in the active Anki profile."}
	}
	var ids []int64
	if err := i.anki.call(ctx, "findCards", map[string]any{"query": `deck:"JP Vocabulary"`}, &ids); err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, nil, &ankiError{400, "JP Vocabulary has no cards to import. Saved data was not replaced."}
	}

	cards := make([]ankiCard, 0, len(ids))
	reviews := make(map[int64][]ankiReview)
	for offset := 0; offset < len(ids); offset += 100 {
		end := min(offset+100, len(ids))
		batch := ids[offset:end]
		params := map[string]any{"cards": batch}
		var info []ankiCard
		if err := i.anki.call(ctx, "cardsInfo", params, &info); err != nil {
			return nil, nil, err
		}
		var history map[string][]ankiReview
		if err := i.anki.call(ctx, "getReviewsOfCards", params, &history); err != nil {
			return nil, nil, err
		}
		if len(info) != len(batch) {
			return nil, nil, &ankiError{502, "Anki returned an incomplete card batch. Please retry after syncing."}
		}
		requested := make(map[int64]bool, len(batch))
		for _, id := range batch {
			requested[id] = true
		}
		for _, card := range info {
			if !requested[card.CardID] || card.NoteID <= 0 || (card.DeckName != vocabularyDeck && !strings.HasPrefix(card.DeckName, vocabularyDeck+"::")) {
				return nil, nil, &ankiError{502, "Anki returned an unexpected card. Saved data was not replaced."}
			}
			delete(requested, card.CardID)
			front, frontOK := card.Fields["Framsida"]
			back, backOK := card.Fields["Baksida"]
			if !frontOK || !backOK {
				return nil, nil, &ankiError{422, fmt.Sprintf("Card %d does not have Framsida and Baksida fields. A field mapping is needed.", card.CardID)}
			}
			card.Vocabulary = strings.Join(strings.Fields(plainAnkiText(front.Value)), " ")
			card.AnswerText = plainAnkiText(back.Value)
			card.Reading = extractAnkiReading(card.Vocabulary, card.AnswerText)
			if card.Vocabulary == "" {
				return nil, nil, &ankiError{422, fmt.Sprintf("Card %d has an empty vocabulary front. Saved data was not replaced.", card.CardID)}
			}
			entries, ok := history[strconv.FormatInt(card.CardID, 10)]
			if !ok {
				return nil, nil, &ankiError{502, "Anki returned incomplete review history. Saved data was not replaced."}
			}
			for _, entry := range entries {
				if entry.ID <= 0 {
					return nil, nil, &ankiError{502, "Anki returned an invalid review identifier."}
				}
			}
			cards = append(cards, card)
			reviews[card.CardID] = entries
		}
	}
	return cards, reviews, nil
}

func ankiRating(entry ankiReview) string {
	// Manual/rescheduling records are preserved but not treated as answers.
	if entry.Type < 0 || entry.Type > 3 {
		return ""
	}
	switch entry.Ease {
	case 1:
		return "again"
	case 2:
		return "hard"
	case 3:
		return "good"
	case 4:
		return "easy"
	default:
		return ""
	}
}

func saveVocabularySnapshot(ctx context.Context, db *sql.DB, cards []ankiCard, history map[int64][]ankiReview) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// There is one supported import source in this iteration. Replace its snapshot
	// atomically, preserving manual events and removing cards deleted from Anki.
	if _, err := tx.ExecContext(ctx, "DELETE FROM review_events WHERE source = 'anki'"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM anki_cards"); err != nil {
		return err
	}
	cardStatement, err := tx.PrepareContext(ctx, `INSERT INTO anki_cards
		(card_id, note_id, deck_name, model_name, vocabulary, reading, answer_text, fields_json, imported_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer cardStatement.Close()
	reviewStatement, err := tx.PrepareContext(ctx, `INSERT INTO review_events
		(anki_review_id, anki_card_id, vocabulary, rating, reviewed_at_ms, source, raw_json)
		VALUES (?, ?, ?, ?, ?, 'anki', ?)`)
	if err != nil {
		return err
	}
	defer reviewStatement.Close()
	importedAt := time.Now().UTC().Format(time.RFC3339)
	for _, card := range cards {
		fields, err := json.Marshal(card.Fields)
		if err != nil {
			return err
		}
		if _, err := cardStatement.ExecContext(ctx, card.CardID, card.NoteID, card.DeckName, card.ModelName, card.Vocabulary, card.Reading, card.AnswerText, string(fields), importedAt); err != nil {
			return err
		}
		for _, entry := range history[card.CardID] {
			raw, err := json.Marshal(entry)
			if err != nil {
				return err
			}
			var rating any
			if value := ankiRating(entry); value != "" {
				rating = value
			}
			if _, err := reviewStatement.ExecContext(ctx, entry.ID, card.CardID, card.Vocabulary, rating, entry.ID, string(raw)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (i *ankiImporter) importStatus(w http.ResponseWriter, r *http.Request) {
	var summary importSummary
	summary.Deck = vocabularyDeck
	var importedAt string
	err := i.db.QueryRowContext(r.Context(), `SELECT COUNT(*), COALESCE(MAX(imported_at), ''),
		COALESCE(SUM(CASE WHEN reading = '' THEN 1 ELSE 0 END), 0),
		(SELECT COUNT(*) FROM review_events WHERE source = 'anki' AND rating IS NOT NULL),
		(SELECT COUNT(*) FROM review_events WHERE source = 'anki' AND rating IS NULL)
		FROM anki_cards`).Scan(&summary.CardCount, &importedAt, &summary.MissingReadingCount, &summary.ReviewCount, &summary.UnscoredReviewCount)
	if err != nil {
		log.Printf("read import status: %v", err)
		http.Error(w, "Could not read the saved import status.", 500)
		return
	}
	if importedAt != "" {
		summary.ImportedAt = &importedAt
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Printf("write import status: %v", err)
	}
}
