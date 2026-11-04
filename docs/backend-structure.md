# Finding your way around the Go backend

The backend is organised by feature. Start it from `backend/` with:

```bash
go run .
```

## Start with `main.go`

`main.go` opens SQLite with `storage.Open`, creates the router, passes database
and HTTP-client dependencies into the features, and starts `net/http`. The health
endpoint is here too. Each feature exposes a small `RegisterRoutes` function;
its handlers, SQL queries, and model types stay private to its package.

| Package | Responsibility | Good starting files |
| --- | --- | --- |
| `internal/quiz` | Questions, attempts, saved results, coverage, mistake queues | `routes.go`, `handler.go`, `validation.go` |
| `internal/anki` | Read-only AnkiConnect requests and transactional vocabulary import | `routes.go`, `client.go`, `import.go`, `text.go` |
| `internal/dictionary` | Kanji lookups with SQLite caching | `routes.go`, `handler.go`, `store.go` |
| `internal/reviews` | Manual/imported review history and review-derived vocabulary/kanji statistics | `routes.go`, `handler.go`, `store.go`, `vocabulary.go`, `kanji.go` |
| `internal/storage` | SQLite connection, shared schema, and migrations | `database.go`, `migrations.go` |
| `internal/testutil` | Temporary SQLite database setup for tests | `database.go` |

Go packages follow directories. Files within `internal/quiz`, for example, share
`package quiz` and can use each other's unexported names. Other packages use the
small exported API instead. Go's `internal/` rule also restricts these packages
to Go code within the `backend/` tree.

## Follow one request

For a completed quiz:

1. `main.go` calls `quiz.RegisterRoutes` from `internal/quiz/routes.go`.
2. The registered handler in `handler.go` decodes the HTTP request.
3. `validation.go` checks completion and grades the submitted snapshots.
4. `store.go` saves or loads the immutable attempt using `database/sql`.
5. `types.go` defines the private JSON request/result shapes used by these files.

For an Insights score, use `internal/reviews/vocabulary.go` or `kanji.go`.
For an Anki parsing issue, use `internal/anki/text.go`. For a dictionary-cache
issue, use `internal/dictionary/store.go`.

`storage` owns table definitions and migrations. Feature packages own their SQL
queries and domain rules. Production feature packages do not import each other;
they work with the database and dependencies supplied by startup.

## The two data directories

- `backend/data/`: your runtime SQLite database, created locally and ignored by Git.
- `backend/internal/quiz/data/`: versioned question catalogs and audit evidence.

The quiz package embeds `questions.json`, `n2.generated.json`, and `exercises.json`
from its own `data/` directory with `//go:embed`. The audit file is for offline
verification. Embedded catalogs are available without runtime file-path lookups
or external network requests.

`backend/scripts/` contains development-only N2 data-building/checking tools.
Their paths and the frontend's test fixtures point to the quiz package's data
directory. Rebuild/restart Go after changing embedded catalog contents.

## Tests

Tests live next to the code they exercise. Package-local tests can inspect
private feature types. Cross-feature behaviour uses registered HTTP routes:
the Anki import tests check the review API, and root integration tests check
application route wiring, HTTP exercise persistence, and cancellation handling.
`internal/testutil` provides isolated SQLite fixtures; tests use temporary files.

From `backend/`:

```bash
go test ./...
go vet ./...
node --test scripts/build-n2.test.mjs

# Focus on one feature.
go test ./internal/quiz

# Measure review scoring at collection scale.
go test -run '^$' -bench '^BenchmarkVocabularyStats$' -benchmem ./internal/reviews
```

Use `./...` for the full Go test suite: the backend root now holds application
integration tests, while feature tests live in the corresponding subpackages.
