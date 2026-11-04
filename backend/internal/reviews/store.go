package reviews

import (
	"context"
	"database/sql"
)

type reviewStore struct{ db *sql.DB }

func (s *reviewStore) loadReviews(ctx context.Context, limit int) ([]review, error) {
	query := `SELECT vocabulary, rating, source, reviewed_at_ms FROM review_events WHERE rating IS NOT NULL`
	var args []any
	if limit > 0 {
		query += ` ORDER BY reviewed_at_ms DESC, id DESC LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reviews := []review{}
	for rows.Next() {
		var entry review
		if err := rows.Scan(&entry.Vocabulary, &entry.Rating, &entry.Source, &entry.ReviewedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, entry)
	}
	return reviews, rows.Err()
}
