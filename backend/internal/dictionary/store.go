package dictionary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func getStoredKanji(ctx context.Context, db *sql.DB, character string) (kanjiDetails, bool, error) {
	var details kanjiDetails
	var meanings, onReadings, kunReadings, nameReadings string
	err := db.QueryRowContext(ctx, `
		SELECT kanji, meanings, on_readings, kun_readings, name_readings, stroke_count
		FROM kanji_entries WHERE kanji = ?
	`, character).Scan(
		&details.Kanji, &meanings, &onReadings, &kunReadings, &nameReadings, &details.StrokeCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return kanjiDetails{}, false, nil
	}
	if err != nil {
		return kanjiDetails{}, false, fmt.Errorf("read stored kanji: %w", err)
	}
	// Decode each JSON column back into its corresponding Go slice.
	encoded := []string{meanings, onReadings, kunReadings, nameReadings}
	lists := []*[]string{&details.Meanings, &details.OnReadings, &details.KunReadings, &details.NameReadings}
	for i, list := range lists {
		if err := json.Unmarshal([]byte(encoded[i]), list); err != nil {
			return kanjiDetails{}, false, fmt.Errorf("decode stored kanji list: %w", err)
		}
		if *list == nil {
			*list = []string{}
		}
	}
	return details, true, nil
}

func saveKanji(ctx context.Context, db *sql.DB, details kanjiDetails, source string) error {
	lists := [][]string{details.Meanings, details.OnReadings, details.KunReadings, details.NameReadings}
	encoded := make([]string, len(lists))
	for i, list := range lists {
		if list == nil {
			list = []string{}
		}
		data, err := json.Marshal(list)
		if err != nil {
			return fmt.Errorf("encode kanji list: %w", err)
		}
		encoded[i] = string(data)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO kanji_entries
			(kanji, meanings, on_readings, kun_readings, name_readings, stroke_count, source, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(kanji) DO UPDATE SET
			meanings = excluded.meanings,
			on_readings = excluded.on_readings,
			kun_readings = excluded.kun_readings,
			name_readings = excluded.name_readings,
			stroke_count = excluded.stroke_count,
			source = excluded.source,
			fetched_at = excluded.fetched_at
	`, details.Kanji, encoded[0], encoded[1], encoded[2], encoded[3], details.StrokeCount,
		source, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save kanji: %w", err)
	}
	return nil
}
