// Package quiz handles question catalogs, completed attempts, coverage, and practice.
package quiz

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes loads the bundled catalogs and registers the quiz API.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) error {
	catalog, err := loadQuizCatalog(quizData, n2QuizData, exerciseQuizData)
	if err != nil {
		return err
	}
	store := quizStore{db: db}
	mux.HandleFunc("POST /quiz/attempts", store.createAttempt)
	mux.HandleFunc("GET /quiz/questions", catalog.listQuestions)
	mux.HandleFunc("GET /quiz/coverage", store.listCoverage(&catalog))
	mux.HandleFunc("GET /quiz/attempts", store.listAttempts)
	mux.HandleFunc("GET /quiz/practice", store.listPracticeCandidates)
	mux.HandleFunc("GET /quiz/attempts/{id}", store.getAttempt)
	return nil
}
