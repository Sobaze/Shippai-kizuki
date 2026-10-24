# Shippai Kizuki 失敗気づき

A personal Japanese learning and analytics application built around:

**Mistake → Insight → Learning → Improvement**

The project starts with a React + TypeScript frontend and a Go backend. Vocabulary
and kanji mistake analysis, SQL persistence, and Anki integration will be explored
in small iterations.

## Tech stack

- **Frontend:** React + TypeScript, with Vite for development and builds
- **Backend:** Go
- **Database:** SQL-based; selection planned for a later iteration

## Installation and setup

### 1. Install prerequisites

Install these tools on your computer:

| Tool | Version used during setup | Download |
| --- | --- | --- |
| Git | Any recent version | [git-scm.com](https://git-scm.com/downloads) |
| Node.js (includes npm) | 24 | [nodejs.org](https://nodejs.org/) |
| Go | 1.27 | [go.dev](https://go.dev/dl/) |

On Windows, you can install them with Windows Package Manager:

```powershell
winget install --id Git.Git --exact --source winget
winget install --id OpenJS.NodeJS.LTS --exact --source winget
winget install --id GoLang.Go --exact --source winget
```

Skip tools you already have installed. After installation, reopen your terminal
(and restart VS Code if using its integrated terminal) to pick up PATH changes.

Check that the tools are available:

```powershell
git --version
node --version
npm --version
go version
```

### 2. Clone the repository

```powershell
git clone https://github.com/Sobaze/Shippai-kizuki.git
cd Shippai-kizuki
```

If you already have the repository locally, open a terminal at its root instead.
Both `frontend/` and `backend/` belong to this single Git repository.

### 3. Install frontend dependencies

From the repository root:

```powershell
cd frontend
npm ci
```

`npm ci` installs the dependency versions recorded in `package-lock.json`.
Use it for a fresh checkout. When intentionally adding or updating dependencies,
use `npm install` instead.

The backend currently uses only Go's standard library, so it requires no
third-party dependency installation. Its Go module is already initialized.

## Running locally

### Frontend

From the repository root:

```powershell
cd frontend
npm run dev
```

Open the local URL printed by Vite (normally **http://localhost:5173**).
The frontend shows vocabulary difficulty scores fetched from the Go backend. Keep this
terminal running while using the frontend; press **Ctrl+C** to stop the development
server. Run the backend in a second terminal using the instructions below.

With both running, the page displays one row per word, sorted by difficulty, with
review counts, rating breakdowns, and a practice suggestion. Use **Refresh scores**
to recalculate after adding reviews through the API. Expand **View individual
review history** to inspect the original rating events; that list has its own
Refresh button. Loading, empty-list, and request-error states are included.

### Backend

In a separate terminal, from the repository root:

```powershell
cd backend
go run .
```

The backend starts a local HTTP server on **http://localhost:8080** and logs a
startup message. Keep the terminal running; press **Ctrl+C** to stop the server.

Open **http://localhost:8080/health** in your browser, or use another PowerShell
terminal:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/health
```

The endpoint returns HTTP `200 OK` with this JSON body (PowerShell displays it as
an object):

```json
{"status":"ok"}
```

This confirms that the server can receive a request and return JSON. The endpoint
does not check a database or other services.

### How the frontend connects

```text
React fetch('/api/reviews') → Vite development proxy → Go GET /reviews
React fetch('/api/vocabulary-stats') → Vite development proxy → Go GET /vocabulary-stats
```

The proxy is configured in `frontend/vite.config.ts`. It forwards `/api` requests
to `http://localhost:8080` and removes the `/api` prefix. The browser uses the
frontend's origin, so this local setup does not require backend CORS configuration.
This proxy is a development-server feature; production hosting will need its own
API routing when deployment is introduced.

### Current API

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/health` | Return `{"status":"ok"}` to confirm the server responds |
| GET | `/reviews` | Return recorded reviews as a JSON array (`200 OK`) |
| POST | `/reviews` | Record a vocabulary review and return it (`201 Created`) |
| GET | `/vocabulary-stats` | Group reviews by vocabulary and return difficulty scores (`200 OK`) |

#### List reviews

Open **http://localhost:8080/reviews**, or run:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/reviews
```

The server starts with 18 demo reviews across five vocabulary items, defined in
`backend/seed.go`. The histories deliberately include failures, difficult recalls,
successful recalls, and limited evidence so we can inspect the scoring behavior.

These are sample reviews, not imported Anki data. Reviews are stored in memory:
restarting the backend discards added entries and restores the demo histories.

#### Vocabulary difficulty scores

Open **http://localhost:8080/vocabulary-stats**, or run:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/vocabulary-stats
```

The backend groups review events by their exact vocabulary text. Each word gets
its own review count, rating counts, integer difficulty score, and recommendation.
Scores are calculated on request from the original reviews; they are not stored
separately.

The initial weights are **Again = 1**, **Hard = 0.5**, **Good = 0**, **Easy = 0**:

```text
difficulty = round(100 × (Again count + 0.5 × Hard count) / review count)
```

Good and Easy still influence the result: they increase the review count without
adding difficulty weight, reducing the average. Difficulty ranges from 0 to 100,
with higher scores indicating more recorded difficulty.

Recommendations use the rounded score:

- Fewer than 3 reviews: `limited_evidence`, regardless of score.
- At least 3 reviews and difficulty at least 50: `suggested_practice`.
- Otherwise: `no_extra_practice`.

The demo results, ordered from highest difficulty to lowest, are:

| Vocabulary | Again | Hard | Good | Easy | Reviews | Difficulty | Suggestion |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 書く | 3 | 2 | 0 | 0 | 5 | 80 | Suggested practice |
| 図書館 | 2 | 1 | 1 | 0 | 4 | 63 | Suggested practice |
| 読書 | 0 | 4 | 0 | 0 | 4 | 50 | Suggested practice |
| 食べる | 0 | 1 | 2 | 1 | 4 | 13 | No extra practice suggested |
| 日本語 | 0 | 0 | 0 | 1 | 1 | 0 | Limited evidence |

The calculation and thresholds are experimental heuristics, not calibrated
measurements of mastery. Three reviews are only a provisional minimum, not a
statistical confidence guarantee. All reviews currently have equal influence,
regardless of age. There are no timestamps, recency weights, or kanji scores yet.
The adjustable scoring constants live in `backend/vocabulary.go`.

An empty history returns `[]`. Equal scores are ordered by vocabulary text to
keep the response deterministic.

#### Record a review

Send a JSON object with these fields:

| Field | Allowed values |
| --- | --- |
| `vocabulary` | Non-empty text; surrounding whitespace is trimmed |
| `rating` | Exactly `again`, `hard`, `good`, or `easy` |

For example, in PowerShell:

```powershell
$body = '{"vocabulary":"図書館","rating":"hard"}'
$utf8Body = [System.Text.Encoding]::UTF8.GetBytes($body)
Invoke-RestMethod -Uri http://localhost:8080/reviews -Method Post -ContentType 'application/json; charset=utf-8' -Body $utf8Body
```

Encoding the request body as UTF-8 preserves Japanese text in Windows PowerShell.
The server returns `201 Created` with the saved review:

```json
{ "vocabulary": "図書館", "rating": "hard" }
```

Requesting `GET /reviews` afterward returns the demo reviews followed by this
new entry. Click **Refresh scores** to see the updated vocabulary score, or expand
the individual history and click its **Refresh** button to see the new event.

To try improving a score, submit a `good` review for 図書館 after a fresh backend
start. Its difficulty changes from 63 to 50 because the same total difficulty
weight is now divided by five reviews rather than four.

Each valid POST adds a separate review, including repeated reviews of the same
word. Ratings are preserved for future analysis; kanji extraction and scoring
are not implemented yet.

Invalid JSON, missing or invalid fields, unknown fields, multiple JSON values,
and request bodies larger than 1 MiB return `400 Bad Request` with a plain-text
error message. Rejected requests do not add a review.

The previous `/mistakes` example endpoint has been replaced by `/reviews`.
After changing Go code, stop the server with Ctrl+C and run `go run .` again
to load the changes.

## Development checks

Inside `frontend/`:

```powershell
npm run build
npm run lint
```

- `npm run build` checks TypeScript and produces a production build in `frontend/dist/`.
- `npm run lint` checks the frontend code for potential issues.

Inside `backend/`:

```powershell
go test ./...
go vet ./...
```

`go test` checks review creation, listing, input validation, vocabulary grouping,
score calculations, threshold boundaries, limited evidence, ordering, and score
updates after new reviews. The HTTP tests run in process and do not need a running server
or a free port.
`go vet` checks Go code for common mistakes.

## Project structure

```text
Shippai-kizuki/
├── backend/
│   ├── go.mod          # Go module identity and Go version
│   ├── main.go         # Server startup, routing, and health endpoint
│   ├── reviews.go      # Review type, in-memory storage, and GET/POST handlers
│   ├── reviews_test.go # Review API tests
│   ├── seed.go         # Demo histories loaded on startup
│   ├── vocabulary.go  # Vocabulary aggregation, scoring, and stats endpoint
│   └── vocabulary_test.go # Scoring and stats API tests
├── frontend/
│   ├── public/         # Static assets
│   ├── src/
│   │   ├── features/
│   │   │   ├── reviews/
│   │   │   │   ├── ReviewsPage.tsx # Individual review history
│   │   │   │   ├── api.ts          # Review API calls
│   │   │   │   ├── types.ts        # Review and rating types
│   │   │   │   └── reviews.css
│   │   │   └── vocabulary/
│   │   │       ├── VocabularyPage.tsx # Scores, evidence, and suggestions
│   │   │       ├── api.ts             # Vocabulary stats API call
│   │   │       ├── types.ts           # Stats and recommendation types
│   │   │       └── vocabulary.css
│   │   ├── App.tsx                # Application shell
│   │   ├── App.css                # Application shell styles
│   │   ├── index.css              # Global styles
│   │   └── main.tsx               # React entry point
│   ├── vite.config.ts  # Development server and API proxy
│   ├── package.json    # Frontend dependencies and npm commands
│   └── package-lock.json
├── .gitignore
└── README.md
```

The Go module path matches this repository's GitHub location; running it locally
does not require publishing it.

Frontend code is grouped by feature. Keep feature-specific UI, API calls, types,
and styles together; extract shared code when it is needed by multiple features.

## Future Anki integration

AnkiConnect is the likely starting point for reading your own decks and card
information. The intended flow is for the Go backend to communicate with Anki
and expose data to the frontend. Connection setup, available review history, and
field mapping will be explored in a later iteration. The current page uses only
local demo and manually submitted reviews.

## Troubleshooting

- **`go`, `node`, or `npm` is not recognized:** reopen your terminal after installing
  the tool. If using VS Code, restart it too. Verify the installation using the
  version commands above.
- **Frontend dependencies are missing:** run `npm ci` inside `frontend/`.
- **Port 5173 is already in use:** use the alternative local URL printed by Vite.
- **The backend reports that port 8080 is already in use:** stop any previous
  instance of the backend with Ctrl+C before starting it again.
- **The review list shows a request error:** check that the Go backend is running
  on port 8080, then click Refresh. Restart Vite after changing its proxy settings.
- **The demo reviews do not appear:** restart the backend with `go run .` so it
  loads the latest startup data.
- **Commands cannot find `package.json` or `go.mod`:** run npm commands inside
  `frontend/` and Go commands inside `backend/`.
