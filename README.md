# Shippai Kizuki 失敗気づき

A personal Japanese learning and analytics application built around:

**Mistake → Insight → Learning → Improvement**

The project starts with a React + TypeScript frontend and a Go backend. Vocabulary
and kanji mistake analysis, SQL persistence, and Anki integration will be explored
in small iterations.

## Tech stack

- **Frontend:** React + TypeScript, with Vite for development and builds
- **Backend:** Go, using the standard-library HTTP server and SQL API
- **Database:** SQLite for imported Anki cards, answer history, manual reviews, and kanji dictionary entries

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

The backend uses Go's `database/sql` package with the pure-Go
`modernc.org/sqlite` driver. Go downloads its dependencies automatically on the
first run. To download them ahead of time, run this from the repository root:

```bash
cd backend
go mod download
```

The SQLite driver works without a C compiler or a separate database server.
`go.mod` records dependency versions and `go.sum` records their checksums.

## Running locally

### Frontend

From the repository root:

```powershell
cd frontend
npm run dev
```

Open the local URL printed by Vite (normally **http://localhost:5173**).
The frontend shows kanji insights and vocabulary difficulty scores fetched from the Go backend. Keep this
terminal running while using the frontend; press **Ctrl+C** to stop the development
server. Run the backend in a second terminal using the instructions below.

With both running, choose **Import JP Vocabulary** in the Anki panel to save your
cards and answer history. Importing automatically refreshes the analysis panels.
**Connect to Anki** separately lists decks from your running desktop collection.
The rest of the page displays kanji cards with expandable evidence.
Use **Refresh kanji** to fetch their latest scores. Below those, it displays one row per word, sorted by difficulty, with
review counts, rating breakdowns, and a practice suggestion. Use **Refresh scores**
to recalculate after adding reviews through the API. Expand **View individual
review history** to inspect the original rating events; that list has its own
Refresh button. Loading, empty-list, and request-error states are included.
The lists initially show 24 kanji, 50 vocabulary items, and the latest 100 review
events to keep large imports responsive. **Show more** expands kanji/vocabulary
lists; all saved scored events contribute to difficulty calculations.

Choose **Meanings & readings →** on a kanji card to open its detail popup. Go
looks up its meanings, on’yomi, kun’yomi, name readings (when available), and
stroke count from SQLite when an entry is saved. Missing entries are fetched from
kanjiapi.dev and saved for later. Only the initial lookup of an uncached kanji
requires internet access; no API key is needed.

### Backend

In a separate terminal, from the repository root:

```powershell
cd backend
go run .
```

The backend starts a local HTTP server on **http://localhost:8080** and logs a
startup message. Keep the terminal running; press **Ctrl+C** to stop the server.

On startup, the backend creates `data/shippai-kizuki.sqlite` and the
`kanji_entries`, `anki_cards`, and `review_events` tables if they do not exist.
It loads saved data rather than seeding demo reviews. A fresh database starts empty.
Run Go from `backend/` as shown above:
the database path is relative to the server's working directory.

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
React fetch('/api/kanji-stats') → Vite development proxy → Go GET /kanji-stats
React fetch('/api/kanji/書') → Vite development proxy → Go GET /kanji/書 → SQLite
                                                                     ↓ if missing
                                                                  kanjiapi.dev
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
| GET | `/reviews` | Return the latest 100 scored events, newest first (`200 OK`) |
| POST | `/reviews` | Record a vocabulary review and return it (`201 Created`) |
| GET | `/vocabulary-stats` | Group reviews by vocabulary and return difficulty scores (`200 OK`) |
| GET | `/kanji-stats` | Return kanji difficulty and contributing vocabulary evidence (`200 OK`) |
| GET | `/kanji/{character}` | Read a saved kanji entry or fetch and save it (`200 OK`) |
| GET | `/anki/decks` | Read deck names from the active Anki Desktop profile (`200 OK`) |
| GET | `/anki/import-status` | Read the saved JP Vocabulary import summary from SQLite |
| POST | `/anki/import-vocabulary` | Import/re-import JP Vocabulary cards and answer history |

#### List reviews

Open **http://localhost:8080/reviews**, or run:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/reviews
```

The server returns saved Anki answers and manually recorded reviews from SQLite.
Each event includes `vocabulary`, `rating`, `source` (`anki` or `manual`), and
`reviewedAt` (Unix milliseconds). The history endpoint is limited to 100 recent
answers; both scoring endpoints load the complete scored history. Restarting Go
preserves the data. Demo histories exist only as fixtures in `backend/seed_test.go`.

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

As an illustrative calculation, Again, Hard, Good, Again gives
`round(100 × 2.5 / 4) = 63`. The actual page uses your saved history, not these
example ratings. Multiple cards with the same cleaned front text contribute to
the same vocabulary item. Conjugated forms and other differently written fronts
remain distinct; this iteration does not perform lemmatization.

The calculation and thresholds are experimental heuristics, not calibrated
measurements of mastery. Three reviews are only a provisional minimum, not a
statistical confidence guarantee. All reviews currently have equal influence,
regardless of age. Anki review timestamps and original record types are preserved,
but recency weighting has not been introduced yet.
The adjustable scoring constants live in `backend/vocabulary.go`.

An empty history returns `[]`. Equal scores are ordered by vocabulary text to
keep the response deterministic.

#### Kanji insights

Open **http://localhost:8080/kanji-stats**, or in Git Bash run:

```bash
curl http://localhost:8080/kanji-stats
```

The backend extracts Han characters from the saved vocabulary text. For example:

| Word | Component kanji |
| --- | --- |
| 図書館 | 図, 書, 館 |
| 書く | 書 |
| 読書 | 読, 書 |
| 食べる | 食 |
| 日本語 | 日, 本, 語 |

Kana-only vocabulary remains in the vocabulary scores but contributes no kanji.
Repeating a kanji within one word does not create another context, and the
iteration mark 々 is not scored as a separate kanji. This is character extraction,
not a dictionary-based morphological analysis.

For each kanji:

1. List its distinct vocabulary items and their current difficulty/review counts.
2. A word qualifies for scoring once it has **at least 3 reviews**.
3. Calculate the rounded arithmetic mean of qualifying vocabulary scores. Each
   qualifying word has equal influence, regardless of its review count. Vocabulary
   scores are already rounded integers.
4. Show less-reviewed words as **early evidence**, excluded from the calculation.
5. With no qualifying words, return `difficulty: null`, displayed as **Not scored yet**.

Recommendations are:

| Evidence | Recommendation |
| --- | --- |
| No qualifying words | `limited_review_evidence` |
| One qualifying word | `limited_cross_word_evidence` |
| At least two qualifying words and score >= 50 | `suggested_practice` |
| At least two qualifying words and score < 50 | `no_extra_practice` |

For illustration, if 書く, 図書館, and 読書 have qualifying scores of 80, 63,
and 50, 書 would score 64. Your actual results depend on your imported history.

Expand **Inspect contributing words** on a card to see each word's score, review
count, and whether it is included or early evidence. Results are ordered by score
descending, with unscored kanji last and character order breaking ties. Empty or
entirely kana-only review histories return `[]`.

After an import, the frontend automatically reloads both scoring sections. For
manually submitted reviews, use the individual Refresh controls or reload the page.

These scores describe difficulty associated with a kanji's vocabulary contexts,
not proof that the kanji caused the difficulty. Evidence thresholds are provisional
heuristics. Graded practice exercises and scoring refinements are future iterations.

#### Kanji dictionary details

Select **Meanings & readings →** on any kanji card. A modal popup opens above the
existing overview, with a dimmed, blurred backdrop. Close it with **Close ×**,
**Escape**, or a click outside the popup. The native HTML dialog makes the
background non-interactive while open; closing returns focus to the opener and
preserves the overview. The popup scrolls independently of the background page.

The detail view includes:

- A large kanji with its meanings, stroke count, and score on the left.
- Dictionary-listed on’yomi (音読み) above kun’yomi (訓読み) on the right.
- Expandable name readings (名乗り) and dictionary notation help.
- Your existing difficulty score and contributing vocabulary evidence.
- Per-word **Show furigana / Hide furigana** buttons below the kanji/readings.
- Loading, error, and retry behavior, plus dictionary source attribution.

##### Small reading-recall practice

Vocabulary readings start hidden each time the popup opens. Try recalling a
word's reading, then choose **Show furigana** to display the full-word reading
above it using HTML ruby text. Each word can be revealed and hidden independently.
Revealing a reading does not submit a review or change a score.

Readings are extracted from the imported `Baksida` field. `GET /kanji-stats`
includes the optional `reading` field on vocabulary evidence. This display data
is added after scoring and does not influence the calculation. If cards sharing
the same written front have conflicting readings, the evidence leaves its reading
unset instead of picking one arbitrarily.

These whole-word readings are separate from the live kanji dictionary lookup;
they are not assembled from on’yomi/kun’yomi or aligned to individual characters.
If a word has no reading, the popup shows **Reading not available** instead of
guessing. The original answer text and field HTML are retained in SQLite so the
parser can be refined later without losing the source material.

For example, open **http://localhost:8080/kanji/書** or use this ASCII-only
Git Bash command (the URL encodes 書):

```bash
curl -i 'http://localhost:8080/kanji/%E6%9B%B8'
```

Expected response:

```json
{
  "kanji": "書",
  "meanings": ["write"],
  "onReadings": ["ショ"],
  "kunReadings": ["か.く", "-が.き", "-がき"],
  "nameReadings": ["かき"],
  "strokeCount": 10
}
```

Dictionary entries can be updated by the source. A dot marks where okurigana
(the kana ending) begins, for example `か.く` for 書く; hyphens indicate
prefix/suffix forms. We preserve the dictionary notation and explain it in the
view. A whole vocabulary word's reading cannot always be inferred simply by
combining individual kanji readings.

`backend/dictionary.go` first checks the local `kanji_entries` table. A saved
entry is returned immediately without contacting the external dictionary. If
missing, Go calls `https://kanjiapi.dev/v1/kanji/{character}`, validates the
response, saves it to SQLite, and returns it. The external HTTP client has an
8-second timeout.

A lookup does not add reviews or change scores. The response body is the same
whether the entry came from SQLite or the external API. The `X-Dictionary-Source`
header is `kanjiapi` for a newly fetched entry and `sqlite` for a stored one.

The dictionary endpoint can look up individual Han characters whether or not they
are present in your cards. A dictionary lookup alone creates no scoring evidence.

Errors return a plain-text message:

| Status | Meaning |
| --- | --- |
| `400` | Input is not exactly one Han character |
| `404` | No dictionary entry exists for that character |
| `500` | Local database read or write failed |
| `502` | Dictionary service is unavailable or returned an invalid entry |
| `504` | Dictionary connection/request timed out |

Meanings and readings are dictionary data; your difficulty and evidence are
calculated from local review history. Meaning illustrations and stroke-order
graphics are not included in this iteration.

#### Local dictionary storage

`backend/database.go` opens SQLite and creates the initial table on startup.
When run from `backend/`, the file is `backend/data/shippai-kizuki.sqlite`.
The local data directory and SQLite journal files are ignored by Git.

Each row in `kanji_entries` contains:

| Column | Stored value |
| --- | --- |
| `kanji` | The character; primary key, one row per kanji |
| `meanings` | JSON array of English meanings |
| `on_readings` | JSON array of on’yomi |
| `kun_readings` | JSON array of kun’yomi |
| `name_readings` | JSON array of name readings |
| `stroke_count` | Integer stroke count |
| `source` | The upstream dictionary entry URL |
| `fetched_at` | UTC fetch timestamp in RFC 3339 format |

Parameterized SQL queries bind values using `?` placeholders. Inserts use an
upsert so concurrent first-time lookups cannot create duplicate rows. HTTP
requests share a `database/sql` handle configured with one open connection to
serialize these small local database operations. No SQL transaction is held
during an external API request.

Only validated successful entries are stored. Failed requests, missing entries,
and malformed responses are not cached. Database errors are reported rather than
silently treating them as a cache miss. A failed save returns `500`, so a successful
lookup response means the entry is available for later local reads.

Saved entries survive browser refreshes and backend restarts and are available
offline. This first version fetches only missing entries; it does not refresh all
kanji at startup or automatically expire saved entries. Fetch timestamps are
recorded for a later refresh policy. Imported cards and review events are also
persisted in SQLite, in separate tables from dictionary definitions.

##### Verify persistent lookups

With the backend running, use Git Bash:

```bash
curl -i 'http://localhost:8080/kanji/%E6%9B%B8'
```

For a kanji not previously saved, the response includes:

```text
X-Dictionary-Source: kanjiapi
```

Repeat the command; the header should be:

```text
X-Dictionary-Source: sqlite
```

Restart Go with Ctrl+C followed by `go run .` and repeat the command. It should
still report `sqlite`. If the entry was already saved before this check, the
first command will report `sqlite` too. The popup uses the same endpoint, so
opening it also populates and reuses this local cache.

##### Browser-session cache and popup loading

`frontend/src/features/kanji/api.ts` also retains successful dictionary responses
in memory for the current page session. These two caches serve different purposes:

- **Browser memory:** reopening a viewed kanji renders its existing details
  immediately, without another HTTP request or loading-state reset.
- **SQLite:** entries survive page reloads and backend restarts, avoiding another
  external dictionary lookup when the browser needs them again.

Pending requests for the same character are shared. This includes concurrent
hover/focus preloads, opening the popup, and React Strict Mode's development-only
effect setup/cleanup cycle. Strict Mode remains enabled.

Hovering or focusing **Meanings & readings →** preloads that kanji's entry; the
app does not fetch every dictionary entry at startup. Closing a popup stops its
state updates but lets the shared request finish, so it can populate SQLite and
the session cache. Failed requests are removed from the pending-request map and
are not cached; Retry can make a new attempt.

A genuine first lookup can still show loading, especially if its entry is not
yet in SQLite. Reloading the page clears the browser cache but retains SQLite;
development hot reloads can also reset the in-memory cache. Review scores and
furigana visibility are not stored in the dictionary session cache.

A browser investigation of three consecutive openings of a stored 書 entry found
six fetch attempts before this optimization (three cancelled by Strict Mode),
with loading rendered on each opening. Afterward, the same scenario used one
shared request, with no loading state on subsequent openings. In the baseline
browser sample, completed SQLite-backed HTTP requests took about 3–5 ms; this
identified repeated frontend loading/request setup as the main cause of the
visible flash rather than an external API call on each opening.

#### Record a review

Send a JSON object with these fields:

| Field | Allowed values |
| --- | --- |
| `vocabulary` | Non-empty text; surrounding whitespace is trimmed |
| `rating` | Exactly `again`, `hard`, `good`, or `easy` |

For example, in **Git Bash on Windows**:

```bash
curl -i http://localhost:8080/reviews \
  -H 'Content-Type: application/json; charset=utf-8' \
  --data-raw '{"vocabulary":"\u56f3\u66f8\u9928","rating":"hard"}'
```

The JSON Unicode escapes represent 図書館. Using an ASCII-only request here avoids
Windows terminal/native-command encoding issues that can replace Japanese
characters with `?`. Go's JSON decoder converts the escapes back to Japanese.

Alternatively, in **PowerShell**:

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

Requesting `GET /reviews` afterward includes the new manual event among the
latest answers. Click **Refresh scores** to see the updated vocabulary score, or expand
the individual history and click its **Refresh** button to see the new event.

The size of a score change depends on the word's existing history. One Good answer
has less influence on a word with many previous answers than one with only a few.

If a POST returns `201` but the original word's score does not change, check for
a separate `???` vocabulary entry. This means the submitted text was altered
before reaching the API. Use the Unicode-escaped Git Bash example above. Reviews
now persist, so restarting the backend does not remove manually submitted events.

Each valid POST adds a separate review, including repeated reviews of the same
word. Ratings are preserved for analysis, including vocabulary and mapped-kanji
scores. Kanji are extracted from the saved word text.

Invalid JSON, missing or invalid fields, unknown fields, multiple JSON values,
and request bodies larger than 1 MiB return `400 Bad Request` with a plain-text
error message. Rejected requests do not add a review.

The previous `/mistakes` example endpoint has been replaced by `/reviews`.
After changing Go code, stop the server with Ctrl+C and run `go run .` again
to load the changes.

## Development checks

Inside `frontend/`:

```powershell
npm test
npm run build
npm run lint
```

- `npm test` uses Node.js 24's built-in test runner to check dictionary cache hits,
  shared requests, preloading, independent kanji entries, and retry after errors.
- `npm run build` checks TypeScript and produces a production build in `frontend/dist/`.
- `npm run lint` checks the frontend code for potential issues.

Inside `backend/`:

```powershell
go test ./...
go vet ./...
```

`go test` checks review creation, listing, input validation, vocabulary grouping,
score calculations, threshold boundaries, limited evidence, ordering, and score
updates after new reviews. Kanji tests cover equal word weighting, early-evidence
exclusion, cross-word thresholds, repeated characters, unmapped words, and the
transition from unscored to scored as reviews accumulate. Dictionary tests cover
reading/meaning conversion, missing entries, empty reading lists, invalid input,
service failures, malformed responses, and timeouts. Database tests use isolated
temporary SQLite files and verify JSON round trips, upserts, corrupt-data errors,
cache hits, source/timestamp metadata, persistence across reopening with the
upstream offline, and failed-response handling. They use local HTTP test servers
rather than the live dictionary, so the tests work offline after dependencies
are downloaded and do not need the application backend running on port 8080.
Anki tests use a simulated AnkiConnect server to verify the deck-name request,
empty lists, action errors inside HTTP 200 responses, connection failures, and
timeouts. Import tests cover both answer-field formats, real rating conversion,
repeat imports, removals, preserved manual events, transaction rollback, and
kanji evidence from imported readings. They do not require Anki Desktop running.
`go vet` checks Go code for common mistakes.

## Project structure

```text
Shippai-kizuki/
├── backend/
│   ├── go.mod          # Go module identity and Go version
│   ├── go.sum          # Dependency checksums
│   ├── main.go         # Server startup, routing, and health endpoint
│   ├── anki.go         # Read-only AnkiConnect deck-name request
│   ├── anki_test.go    # Offline AnkiConnect HTTP tests
│   ├── anki_import.go  # JP Vocabulary batches, transactional import, and status
│   ├── anki_import_test.go # Import correctness and rollback tests
│   ├── anki_text.go    # Plain-text extraction and conservative reading parsing
│   ├── database.go     # SQLite schema and dictionary reads/writes
│   ├── database_test.go # Isolated SQLite storage tests
│   ├── data/           # Local SQLite file; created at runtime, ignored by Git
│   ├── dictionary.go   # Database-first dictionary lookup and response conversion
│   ├── dictionary_test.go # Offline dictionary HTTP tests
│   ├── reviews.go      # SQLite review queries and manual review recording
│   ├── reviews_test.go # Review API tests
│   ├── seed_test.go    # Example data used only by automated tests
│   ├── kanji.go        # Kanji aggregation, evidence rules, and stats endpoint
│   ├── kanji_test.go   # Kanji scoring and evidence API tests
│   ├── vocabulary.go  # Vocabulary aggregation, scoring, and stats endpoint
│   └── vocabulary_test.go # Scoring and stats API tests
├── frontend/
│   ├── public/         # Static assets
│   ├── src/
│   │   ├── features/
│   │   │   ├── anki/
│   │   │   │   ├── AnkiConnection.tsx # Connection status and deck names
│   │   │   │   ├── VocabularyImport.tsx # Import action and saved summary
│   │   │   │   ├── api.ts
│   │   │   │   ├── types.ts
│   │   │   │   └── anki.css
│   │   │   ├── kanji/
│   │   │   │   ├── KanjiPage.tsx # Kanji cards and expandable word evidence
│   │   │   │   ├── KanjiDetail.tsx # Meanings, readings, and personal evidence
│   │   │   │   ├── KanjiEvidenceTable.tsx # Evidence UI shared by cards/details
│   │   │   │   ├── labels.ts     # Shared recommendation labels
│   │   │   │   ├── kanji-detail.css
│   │   │   │   ├── api.ts
│   │   │   │   ├── types.ts
│   │   │   │   └── kanji.css
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
│   ├── tests/kanji-api.test.ts # Browser dictionary-cache logic tests (Node)
│   ├── package.json    # Frontend dependencies and npm commands
│   └── package-lock.json
├── .gitignore
└── README.md
```

The Go module path matches this repository's GitHub location; running it locally
does not require publishing it.

Frontend code is grouped by feature. Keep feature-specific UI, API calls, types,
and styles together; extract shared code when it is needed by multiple features.

## Dictionary data sources

Kanji meanings and readings use the [KANJIDIC2 dictionary](https://www.edrdg.org/wiki/KANJIDIC_Project.html),
copyright James William Breen and the Electronic Dictionary Research and
Development Group (EDRDG), accessed via [kanjiapi.dev](https://kanjiapi.dev/).
KANJIDIC2 data is provided under [Creative Commons Attribution-ShareAlike 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
See the [EDRDG dictionary licence and documentation](https://www.edrdg.org/edrdg/licence.html).
The detail view includes source and licence links. Personal review history,
derived difficulty scores, and vocabulary field data are separate from this
dictionary data.

## Local Anki connection

The integration lists your real deck names and imports JP Vocabulary. The
connection is checked when you click **Connect to Anki**, **Refresh decks**, or
**Retry connection**. There is no automatic polling. The displayed connection
status and time describe the last check, not a continuously monitored connection.

### Desktop setup

1. Sync your phone's collection to your AnkiWeb account.
2. Install [Anki Desktop](https://apps.ankiweb.net/) and sign in to the same
   AnkiWeb account. For a fresh empty desktop profile, choose **Download** on the
   initial sync to retrieve the collection already uploaded from your phone.
3. Open **Tools → Add-ons → Get Add-ons**, enter `2055492159`, and restart Anki.
4. Keep Anki Desktop open with the intended profile loaded.

The add-on is [AnkiConnect](https://ankiweb.net/shared/info/2055492159). Go talks
to its default local endpoint, `http://127.0.0.1:8765/`, using the legacy v6
`deckNames` action verified during setup. This iteration assumes the default
configuration with no API key.

```text
React GET /api/anki/decks → Go GET /anki/decks → AnkiConnect POST deckNames
```

Restart Go after adding this route, then click **Connect to Anki** on the frontend.
You can also check our backend directly from Git Bash:

```bash
curl -i http://localhost:8080/anki/decks
```

Example response:

```json
{
  "decks": ["JP Grammar", "JP Sentences", "JP Vocabulary", "Standard"]
}
```

Go uses a five-second HTTP timeout. Connection failures return `503`, timeouts
return `504`, and invalid/action-error responses from AnkiConnect return `502`,
with a readable error message. The frontend displays the message and a retry
button. Successful responses are marked `Cache-Control: no-store` so refreshing
checks the current active collection.

### Import JP Vocabulary

1. Sync your phone, then sync Anki Desktop with the same AnkiWeb account.
2. Keep the intended Anki Desktop profile open.
3. Choose **Import JP Vocabulary**. It works independently of the deck-list button.
4. After success, the summary and scoring panels refresh automatically. Saved
   analysis remains available when Anki is closed or Go is restarted.

The import is deliberately limited to `JP Vocabulary` and its subdecks, from one
active desktop profile. It calls only read operations: `deckNames`, `findCards`,
`cardsInfo`, and `getReviewsOfCards`. `JP Grammar`, `JP Sentences`, and `Standard`
are not imported. Anki itself is not modified.

The current field mapping matches this deck's Basic/Grundläggande notes:

| Field | Use |
| --- | --- |
| `Framsida` | Cleaned written vocabulary; formatting and ruby hints removed |
| `Baksida` | Full answer text retained; reading extracted when recognizable |

Readings are accepted from a kana first line or a `【読み】` section. Kana-only
fronts can supply their own reading. Parenthesized kana such as `うわき(する)`
is preserved. Unknown formats remain unparsed rather than guessing. The HTML
tokenizer comes from `golang.org/x/net/html`; it safely extracts text and line
breaks while the original fields remain saved as JSON. Raw Anki HTML is not
rendered in the website.

Cards are fetched in batches of 100. Their original card/note IDs, deck name,
note type, fields, parsed vocabulary/reading, and import time go into `anki_cards`.
Review IDs, card IDs, timestamps, source, and original review records go into
`review_events`.
Cards with no answer history are saved but do not receive difficulty scores
until there is a scored answer to analyze.

Anki ease **1 → Again**, **2 → Hard**, **3 → Good**, **4 → Easy**. Learning,
review, relearning, and filtered answer events (types 0–3) count equally in the
first scoring model. Manual/rescheduling or unsupported records are retained
with a null rating and excluded from scoring. Scores are never inferred from
the card's current interval, ease factor, lapse count, or total repetitions.

All network reads and validation finish before the database transaction starts.
The transaction replaces the previous Anki snapshot and commits only after all
cards and reviews have been saved. Repeat imports therefore do not duplicate
answer history and remove cards that have left this deck. Manual site-recorded
events and cached dictionary entries are preserved. An empty/missing deck, an
unexpected card field layout, a failed batch, or a failed database write leaves
the prior snapshot intact. Only one import runs at a time; its overall timeout
is 60 seconds, with the existing five-second limit per AnkiConnect request.

You can also trigger and inspect the import through Go:

```bash
curl -X POST http://localhost:8080/anki/import-vocabulary
curl http://localhost:8080/anki/import-status
```

A fresh database starts with zero imported cards and no demo reviews. Existing
dictionary cache entries are retained when the new tables are initialized. The
import status includes card/answer counts, unscored record count, missing-reading
count, and last import time. Actual counts change as you study and add cards.

For ongoing phone use, sync the phone to AnkiWeb and then sync Anki Desktop before
re-importing JP Vocabulary. Refreshing deck names alone does not update the saved
review history. AnkiConnect reads the desktop collection;
it does not contact AnkiWeb directly.

## Troubleshooting

- **`go`, `node`, or `npm` is not recognized:** reopen your terminal after installing
  the tool. If using VS Code, restart it too. Verify the installation using the
  version commands above.
- **Frontend dependencies are missing:** run `npm ci` inside `frontend/`.
- **Port 5173 is already in use:** use the alternative local URL printed by Vite.
- **The backend reports that port 8080 is already in use:** stop any previous
  instance of the backend with Ctrl+C before starting it again.
- **Anki cannot be reached:** keep Anki Desktop open, confirm AnkiConnect is
  enabled, and use **Retry connection**. The default endpoint is
  `http://127.0.0.1:8765/`. Restart Anki after installing the add-on.
- **The Anki panel reports `404 page not found`:** restart the Go backend so it
  loads the new `/anki/decks` route.
- **The review list shows a request error:** check that the Go backend is running
  on port 8080, then click Refresh. Restart Vite after changing its proxy settings.
- **The analysis is empty:** import JP Vocabulary. Production no longer seeds
  demo reviews; its history comes from SQLite.
- **Dictionary details fail to load:** confirm the backend has been restarted
  and, for a kanji not yet saved, that it can reach `https://kanjiapi.dev`. Use
  **Retry dictionary lookup** after a temporary connection or service error.
- **A kanji briefly loads on its first opening:** it may not be in browser memory
  yet. Check `X-Dictionary-Source` on its network request: `sqlite` is a local read,
  while `kanjiapi` is a first-time external fetch. Reopening the same successful
  entry in the page session should make no new dictionary request.
- **The backend cannot open or write the database:** confirm `backend/data/` is
  writable and run `go run .` from `backend/` so it uses the expected database file.
- **A word has no furigana:** its answer format may be unrecognized, or cards
  sharing that written word may disagree on the reading. The import summary
  reports cards with missing parsed readings; their review history still counts.
- **The import reports missing Framsida/Baksida fields:** the active deck has an
  unsupported note layout. Its field mapping needs to be added before importing.
- **Commands cannot find `package.json` or `go.mod`:** run npm commands inside
  `frontend/` and Go commands inside `backend/`.
