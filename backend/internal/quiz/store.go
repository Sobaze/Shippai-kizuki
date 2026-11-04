package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type quizStore struct{ db *sql.DB }

var errQuizAttemptConflict = errors.New("quiz attempt contents conflict")

func resultExerciseID(result quizResult) string {
	if len(result.Items) == 0 {
		return ""
	}
	id := result.Items[0].Question.ExerciseID
	for _, item := range result.Items {
		if item.Question.ExerciseID != id {
			return ""
		}
	}
	return id
}

func (s *quizStore) saveAttempt(ctx context.Context, input quizSubmission, result quizResult) (savedQuizAttempt, bool, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return savedQuizAttempt{}, false, err
	}
	completed, err := time.Parse(time.RFC3339Nano, input.CompletedAt)
	if err != nil {
		return savedQuizAttempt{}, false, err
	}
	write, err := s.db.ExecContext(ctx, `INSERT INTO quiz_attempts
		(id, mode, kind, parent_attempt_id, started_at, completed_at, completed_at_ms, correct_count, total, result_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`,
		input.ID, input.Mode, input.Kind, input.ParentAttemptID, input.StartedAt, input.CompletedAt,
		completed.UnixMilli(), result.CorrectCount, result.Total, string(encoded))
	if err != nil {
		return savedQuizAttempt{}, false, err
	}
	count, err := write.RowsAffected()
	if err != nil {
		return savedQuizAttempt{}, false, err
	}
	attempt, err := s.loadAttempt(ctx, input.ID)
	if err != nil {
		return savedQuizAttempt{}, false, err
	}
	if count == 0 {
		stored, err := json.Marshal(attempt.Result)
		if err != nil {
			return savedQuizAttempt{}, false, err
		}
		parentsMatch := input.ParentAttemptID == nil && attempt.ParentAttemptID == nil
		if input.ParentAttemptID != nil && attempt.ParentAttemptID != nil {
			parentsMatch = *input.ParentAttemptID == *attempt.ParentAttemptID
		}
		if attempt.Mode != input.Mode || attempt.Kind != input.Kind || attempt.StartedAt != input.StartedAt || attempt.CompletedAt != input.CompletedAt || !parentsMatch || string(stored) != string(encoded) {
			return savedQuizAttempt{}, false, errQuizAttemptConflict
		}
	}
	return attempt, count == 1, nil
}

func (s *quizStore) loadAttempt(ctx context.Context, id string) (savedQuizAttempt, error) {
	var attempt savedQuizAttempt
	var encoded string
	err := s.db.QueryRowContext(ctx, `SELECT id, mode, kind, completed_at, correct_count, total,
		started_at, parent_attempt_id, result_json FROM quiz_attempts WHERE id = ?`, id).Scan(
		&attempt.ID, &attempt.Mode, &attempt.Kind, &attempt.CompletedAt, &attempt.CorrectCount, &attempt.Total,
		&attempt.StartedAt, &attempt.ParentAttemptID, &encoded)
	if err != nil {
		return savedQuizAttempt{}, err
	}
	if err := json.Unmarshal([]byte(encoded), &attempt.Result); err != nil {
		return savedQuizAttempt{}, err
	}
	attempt.ExerciseID = resultExerciseID(attempt.Result)
	return attempt, nil
}

func (s *quizStore) recentAttempts(ctx context.Context, mode string) ([]quizAttemptSummary, error) {
	query := `SELECT id, mode, kind, completed_at, correct_count, total, result_json FROM quiz_attempts`
	var args []any
	if mode != "" {
		query += ` WHERE mode = ?`
		args = append(args, mode)
	}
	query += ` ORDER BY completed_at_ms DESC, id DESC LIMIT 20`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attempts := []quizAttemptSummary{}
	for rows.Next() {
		var attempt quizAttemptSummary
		var encoded string
		if err := rows.Scan(&attempt.ID, &attempt.Mode, &attempt.Kind, &attempt.CompletedAt, &attempt.CorrectCount, &attempt.Total, &encoded); err != nil {
			return nil, err
		}
		if attempt.Mode == "sentence" {
			var result quizResult
			if err := json.Unmarshal([]byte(encoded), &result); err != nil {
				return nil, err
			}
			attempt.ExerciseID = resultExerciseID(result)
		}
		attempts = append(attempts, attempt)
	}
	return attempts, rows.Err()
}
