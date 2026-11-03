package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQuizCatalogPoolsAndAPI(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		mode, level string
		count       int
	}{{"meaning", "N5", 24}, {"meaning", "N4", 24}, {"reading", "N5", 24}, {"reading", "N4", 24}} {
		response := httptest.NewRecorder()
		catalog.listQuestions(response, httptest.NewRequest("GET", "/quiz/questions?mode="+test.mode+"&level="+test.level, nil))
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		var body struct {
			quizPool
			Version   int
			CheckedAt string
			Sources   map[string]string
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Mode != test.mode || body.Level != test.level || len(body.Questions) != test.count || body.Version != 1 || body.CheckedAt == "" || len(body.Sources) == 0 {
			t.Fatalf("pool metadata: %+v", body)
		}
		for _, q := range body.Questions {
			if q.Mode != test.mode || q.EstimatedLevel != test.level {
				t.Fatalf("incorrect level: %+v", q)
			}
		}
	}
	for _, query := range []string{"", "mode=other&level=N5", "mode=meaning&level=N3", "mode=reading&level=starter", "mode=reading&level=N1", "mode=meaning", "mode=meaning&level=n5"} {
		response := httptest.NewRecorder()
		catalog.listQuestions(response, httptest.NewRequest("GET", "/quiz/questions?"+query, nil))
		if response.Code != 400 {
			t.Fatalf("accepted %q", query)
		}
	}
}

func TestQuizCatalogRejectsBrokenContent(t *testing.T) {
	for name, mutate := range map[string]func(*quizCatalog){
		"too few questions":            func(c *quizCatalog) { c.Pools[0].Questions = c.Pools[0].Questions[:9] },
		"duplicate ID":                 func(c *quizCatalog) { c.Pools[1].Questions[0].ID = c.Pools[0].Questions[0].ID },
		"duplicate kanji":              func(c *quizCatalog) { c.Pools[0].Questions[1].Prompt = c.Pools[0].Questions[0].Prompt },
		"duplicate choices":            func(c *quizCatalog) { c.Pools[0].Questions[0].Choices[1] = c.Pools[0].Questions[0].Answer },
		"missing answer":               func(c *quizCatalog) { c.Pools[0].Questions[0].Answer = "not offered" },
		"missing example":              func(c *quizCatalog) { c.Pools[0].Questions[0].Example.Reading = "" },
		"unrelated example":            func(c *quizCatalog) { c.Pools[0].Questions[0].Example.Written = "山" },
		"mode mismatch":                func(c *quizCatalog) { c.Pools[0].Questions[0].Mode = "reading" },
		"unsupported level":            func(c *quizCatalog) { c.Pools[0].Level = "N1" },
		"duplicate pool":               func(c *quizCatalog) { c.Pools = append(c.Pools, c.Pools[0]) },
		"bad reading key":              func(c *quizCatalog) { c.Pools[2].Questions[0].Example.Reading = "different" },
		"non-kana option":              func(c *quizCatalog) { c.Pools[2].Questions[0].Choices[1] = "English" },
		"conflicting level":            func(c *quizCatalog) { c.Pools[2].Questions[0].EstimatedLevel = "N4" },
		"duplicate word across levels": func(c *quizCatalog) { c.Pools[3].Questions[0] = c.Pools[2].Questions[0] },
	} {
		t.Run(name, func(t *testing.T) {
			var catalog quizCatalog
			if err := json.Unmarshal(quizData, &catalog); err != nil {
				t.Fatal(err)
			}
			mutate(&catalog)
			encoded, err := json.Marshal(catalog)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := loadQuizCatalog(encoded); err == nil {
				t.Fatal("broken catalog accepted")
			}
		})
	}
}

func TestQuizLevelSnapshotsAndLegacyQueue(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	legacy := quizFixture("meaning", "retry", 1)
	legacy.Questions[0].ID = "n4-meaning-sea"
	legacy.Answers[0] = quizAnswer{QuestionID: "n4-meaning-sea", SelectedAnswer: "read"}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacy); response.Code != 201 {
		t.Fatal(response.Body.String())
	}
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	for index, pool := range catalog.Pools[:2] {
		input := quizFixture("meaning", "full", 10)
		input.ID = "level-" + pool.Level
		input.Questions = pool.Questions[:10]
		input.Answers = nil
		for _, q := range input.Questions {
			input.Answers = append(input.Answers, quizAnswer{QuestionID: q.ID, SelectedAnswer: q.Choices[1]})
		}
		input.CompletedAt = []string{"2026-11-01T11:00:00Z", "2026-11-01T12:00:00Z"}[index]
		for _, status := range []int{201, 200} {
			response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
			if response.Code != status {
				t.Fatalf("save: %s", response.Body.String())
			}
			var saved savedQuizAttempt
			if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.Result.Items[0].Question.EstimatedLevel != pool.Level {
				t.Fatal("level snapshot lost")
			}
		}
	}
	candidates, err := store.practiceCandidates(context.Background(), "meaning")
	if err != nil || len(candidates) != 20 {
		t.Fatalf("mixed queue: %d %v", len(candidates), err)
	}
	// Historical IDs remain the same, so an old candidate is updated rather than duplicated.
	for _, candidate := range candidates {
		if candidate.Question.ID == "n4-meaning-sea" && candidate.Question.EstimatedLevel != "N4" {
			t.Fatal("latest snapshot did not replace legacy snapshot")
		}
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", legacy); response.Code != 200 {
		t.Fatal("legacy resave broke")
	}
	for _, level := range []string{"N1", "starter"} {
		input := quizFixture("meaning", "retry", 1)
		input.Questions[0].EstimatedLevel = level
		if _, err := validateQuizSubmission(&input); err == nil {
			t.Fatal("invalid estimated level accepted")
		}
	}
	input := quizFixture("reading", "retry", 1)
	input.Questions[0].EstimatedLevel = "N3"
	if _, err := validateQuizSubmission(&input); err == nil {
		t.Fatal("unsupported reading level accepted")
	}
	// Empty level metadata remains absent when a historical snapshot is re-encoded.
	encoded, _ := json.Marshal(legacy.Questions[0])
	if strings.Contains(string(encoded), "estimatedLevel") {
		t.Fatal("legacy serialization changed")
	}
}

func TestReadingLevelsPreserveLegacySnapshotsAndQueue(t *testing.T) {
	store := quizStore{db: testDatabase(t)}
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	question := catalog.Pools[2].Questions[0] // 新しい keeps its starter ID.
	if question.ID != "reading-new" || question.EstimatedLevel != "N5" {
		t.Fatal("starter target changed identity")
	}
	legacy := question
	legacy.EstimatedLevel = ""
	retired := quizQuestion{ID: "reading-seawater", Mode: "reading", Prompt: "海水", Answer: "かいすい",
		Choices: []string{"かいすい", "かいずい", "かいす", "かすい"}, Example: quizExample{Written: "海水", Reading: "かいすい", Meaning: "seawater"}}
	save := func(id, at string, q quizQuestion, correct bool) quizSubmission {
		t.Helper()
		selected := q.Answer
		if !correct {
			selected = q.Choices[1]
		}
		input := quizSubmission{ID: id, Mode: "reading", Kind: "retry", StartedAt: at, CompletedAt: at,
			Questions: []quizQuestion{q}, Answers: []quizAnswer{{QuestionID: q.ID, SelectedAnswer: selected}}}
		response := quizRequest(t, &store, "POST", "/quiz/attempts", input)
		if response.Code != 201 {
			t.Fatal(response.Body.String())
		}
		return input
	}
	old := save("old-reading", "2026-11-01T10:00:00Z", legacy, false)
	save("retired-reading", "2026-11-01T10:01:00Z", retired, false)
	save("new-reading", "2026-11-01T11:00:00Z", question, true)
	candidates, err := store.practiceCandidates(context.Background(), "reading")
	if err != nil || len(candidates) != 2 {
		t.Fatalf("legacy queue lost: %+v %v", candidates, err)
	}
	if candidates[1].Question.ID != legacy.ID || candidates[1].Streak != 1 || candidates[1].Question.EstimatedLevel != "N5" {
		t.Fatalf("legacy progress split: %+v", candidates)
	}
	saved, err := store.loadAttempt(context.Background(), old.ID)
	if err != nil || saved.Result.Items[0].Question.EstimatedLevel != "" {
		t.Fatal("old level retroactively rewritten")
	}
	if response := quizRequest(t, &store, "POST", "/quiz/attempts", old); response.Code != 200 {
		t.Fatal("legacy resave failed")
	}
	save("new-reading-2", "2026-11-01T12:00:00Z", question, true)
	candidates, err = store.practiceCandidates(context.Background(), "reading")
	if err != nil || len(candidates) != 1 || candidates[0].Question.ID != retired.ID {
		t.Fatalf("clearing changed unrelated legacy item: %+v %v", candidates, err)
	}
	// A retired starter question is still usable in snapshot-based queue practice.
	save("retired-retry", "2026-11-01T13:00:00Z", retired, true)
	candidates, err = store.practiceCandidates(context.Background(), "reading")
	if err != nil || candidates[0].Streak != 1 {
		t.Fatal("retired snapshot cannot progress")
	}
}

func TestReadingDistractorsExcludeKnownDictionaryAlternatives(t *testing.T) {
	// JMdict via kanjiapi.dev, reviewed 2026-11-03. Even uncommon valid
	// readings must not be offered as wrong answers for the same written word.
	alternatives := map[string][]string{"甘い": {"うまい"}, "案内": {"あない"}, "安心": {"あんじん"}, "安全": {"あんせん"}, "開ける": {"ひらける"}}
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	for _, pool := range catalog.Pools {
		if pool.Mode != "reading" {
			continue
		}
		for _, q := range pool.Questions {
			for _, alternative := range alternatives[q.Prompt] {
				for _, choice := range q.Choices {
					if choice == alternative && choice != q.Answer {
						t.Fatalf("%s offers valid alternative %s as wrong", q.Prompt, choice)
					}
				}
			}
		}
	}
}

func TestReadingPoolRetainsClassifiedStarterIDs(t *testing.T) {
	catalog, err := loadQuizCatalog(quizData)
	if err != nil {
		t.Fatal(err)
	}
	for id, expected := range map[string]struct{ word, level string }{
		"reading-new": {"新しい", "N5"}, "reading-newspaper": {"新聞", "N5"},
		"reading-study": {"勉強", "N5"}, "reading-near": {"近い", "N5"},
		"reading-breakfast": {"朝ご飯", "N5"}, "reading-recently": {"最近", "N4"},
	} {
		found := false
		for _, pool := range catalog.Pools {
			for _, q := range pool.Questions {
				if q.ID != id {
					continue
				}
				found = true
				if q.Mode != "reading" || q.Prompt != expected.word || q.EstimatedLevel != expected.level {
					t.Fatalf("starter identity changed: %+v", q)
				}
			}
		}
		if !found {
			t.Fatalf("lost starter ID %s", id)
		}
	}
}
