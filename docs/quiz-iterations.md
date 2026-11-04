# Quiz iterations

The goal is useful recall practice that feeds the broader
**Mistake → Insight → Learning → Improvement** loop. This is a working plan,
not a fixed specification; refine it using actual practice and feedback.

## Current priority: N2 exam practice — implemented

With the exam approaching, prioritize N2 over incremental expansion through N3.
Both modes now default to N2: **374 N2-specific kanji meanings** and **1,640
dictionary-matched N2 word readings**. The previous N5/N4 pools and history are
preserved. This covers the chosen N2-specific source data, not a complete
cumulative exam syllabus.

The N2 catalog is bundled with independent vocabulary classification and offline
dictionary evidence. Coverage/selection work over the full larger pools. Queues
default to the selected level, have an all-level/legacy toggle, and expand long
candidate lists in groups of 24. See [n2-practice.md](n2-practice.md) for sources,
scope, first-pass pacing, licensing, and reproducible build/check commands.

## 1. Small kanji meaning quiz — complete

- Estimated N4, ten distinct questions per attempt, four choices per question.
- Start with twelve authored, checked entries; shuffle questions and choices.
- Show progress and review every answer after completion.
- Keep active question state in the page session, separate from Anki difficulty scores.
- Cancel returns to setup, discards the partial run, and records no unfinished score.
- Evaluate wording, distractors, and the results experience before expanding.

Implemented at `/quiz`. Verified the full question/result flow, keyboard radio
selection and submission, 0/10 and 10/10 boundaries, duplicate-submit protection,
and 320px layout. The next step remains subject to feedback from using this version.

Level estimates were checked against
[JLPT Sensei's N4 list](https://jlptsensei.com/jlpt-n4-kanji-list/), and meanings
against [kanjiapi.dev](https://kanjiapi.dev/) / KANJIDIC2. The
[official JLPT FAQ](https://www.jlpt.jp/e/faq/#anchor28) explains why current
test-content lists are not published. This small set is not a complete syllabus
or a reproduction of official test questions.

## 2. Reading practice in vocabulary context — complete

- Bring reading practice forward based on feedback from the first meaning quiz.
- Ten distinct whole-word questions per attempt, drawn from twelve checked starter words.
- Link a kanji to common words and their meanings: for example 新しい →
  あたらしい and 新聞 → しんぶん, rather than asking for one reading of 新.
- Use common on’yomi/kun’yomi examples; do not require every dictionary/name reading.
- Show a checked word, reading, and meaning in learning feedback for both modes.
- Do not infer per-kanji reading alignment automatically from a whole-word reading.
- Later explore recalling before showing choices, kana input, and an explicit
  "I don't know" option to distinguish uncertainty from a selected wrong answer.

Implemented as **Word readings** alongside **Kanji meanings** at `/quiz`.
The word pool is a starter set, not a separately classified N4 vocabulary list.
Whole-word readings and glosses were checked against JMdict through
`kanjiapi.dev/v1/words/{character}`. Words with alternate dictionary readings were
avoided in the initial quiz pool; valid readings must never be used as distractors.
Each meaning question also has a checked example word for its results screen.

Mode selection, cancellation, repeat attempts, kana options, 0/10 and 10/10
boundaries, feedback accuracy, and 320px layout were verified. Correct readings,
furigana, and meanings are not displayed as feedback before the final results;
the question's four possible readings are shown as choices.

Prioritize feedback and targeted retries next, then evaluate pool expansion or images.

## 3. Attempts and missed-item practice — complete (first version)

- Save attempts and their question/answer snapshots separately from Anki reviews.
- Record the task type so multiple-choice recognition is distinguishable from
  Anki recall and future reading exercises.
- Show attempt history and offer a retry of missed questions.
- Include retries after a gap and across sessions; immediate recognition of a
  just-shown answer should not be treated as established recall.
- If cancelled attempts are retained later, distinguish them from completed tests.

Completed attempts now auto-save through Go into SQLite's `quiz_attempts` table.
Each row retains its mode, full/retry kind, timestamps, source-run ID when
available, and the exact prompt/choice/answer/example snapshot. Go validates
completion and recomputes the saved score. Stable attempt IDs make saving safe
to retry without duplicates; changed contents cannot overwrite an existing ID.

The setup view lists the latest twenty completed attempts for the selected mode.
Reviewing one restores its stored result without saving it again. **Retry missed**
uses only the missed snapshots, with shuffled order/options and a new run ID.
Short retries have their own score and label, preserving the original result.
Cancelled runs are not saved, and quiz attempts never enter Anki difficulty scores.

Verified database reopening, duplicate/conflicting saves, mode separation,
history bounds/order, reading and meaning retries, single-item retries, failed
saves, lost-response recovery, cancellation, and mobile history/review layout.

### Cross-attempt practice queue — complete

- Replay all saved completed attempts into separate meaning/reading queues,
  including historical attempts outside the latest-twenty history view.
- A wrong answer adds an item or resets it to **0/2**. Two consecutive correct
  answers in separate saved attempts clear it for now; a later mistake re-adds it.
- Full quizzes and focused retries both count. Unasked questions, cancelled runs,
  unsaved results, and reviewing history do not change progress.
- Back-to-back attempts are allowed. This is a readiness signal, not permanent mastery.
- Display candidates and streaks. Select up to ten least-recently-practised items,
  then shuffle questions and options; each item appears once per run.
- Queue sessions use the existing `retry` kind without a source-attempt ID.
  Each question retains its latest saved snapshot. No queue table is needed.
- Replay by precise completion timestamp, then attempt ID for exact ties.
  Candidate selection ties use question ID for deterministic ordering.

Automatic delayed scheduling and reminders remain a possible later refinement.

Verified historical replay, precise timestamp/tie ordering, resets/re-entry,
duplicate-save idempotence, mode separation, and oldest-first selection. Browser
checks with temporary SQLite covered queue sessions in both modes, cancellation,
0→1→reset→1→clear progress, refresh/re-entry, reload persistence, and 320px layout.

## 4. Level selection and locally stored question data — complete (N5/N4 first batches)

- Meaning practice now offers **24 estimated-N5** and **24 estimated-N4** kanji.
  Pools are separate, not cumulative. The original twelve N4 IDs are preserved.
- The initial data migration retained twelve unclassified starter reading questions.
  These are superseded by the levelled reading pools described below.
- One canonical `backend/internal/quiz/data/questions.json` stores checked questions, example words,
  version, check date, and source references. Go embeds/validates it at startup
  and serves selected pools through `/quiz/questions?mode=meaning&level=N5`.
- The frontend loads the selected pool with cancellation, loading/error/retry
  states, then selects ten distinct questions and shuffles them/options locally.
- New meaning question snapshots carry their estimated level. Historic snapshots
  need no migration. History and practice queues remain mode-based; meaning queues
  can mix levels and label the new snapshots appropriately.
- Sources and the question-review process are documented in [quiz-data.md](quiz-data.md).

Checked all 60 example snapshots against JMdict through kanjiapi.dev, inspected
KANJIDIC meanings and distractors, and verified level membership against JLPT
Sensei. Offline tests cover pool invariants, selection, snapshots, and historical
queue compatibility. Browser checks covered both levels, stale loading responses,
retries, saved review, cross-level queue clearing, reading regression, load-error
recovery, and 320px layout.

Later batches can broaden N5/N4 coverage and introduce N3–N1 after checking their
questions. These pools do not claim complete level coverage.

### Levelled vocabulary readings — complete

- Add **24 estimated-N5** and **24 estimated-N4** reading questions, using
  vocabulary classifications rather than inferring levels from component kanji.
- Both modes now have level selection, with independent remembered selections
  while the page is open. Reading defaults to N5; meanings retain N4 as default.
- Six starter words have verified classifications and keep their original IDs.
  The other six remain usable from saved snapshots/history/queues without being
  retroactively assigned a vocabulary level.
- Both modes' new question snapshots retain level metadata. Results distinguish
  single-level and mixed-level sessions; legacy reading results use a generic label.
- Reading queues continue across both levels and older saved items. The existing
  two-consecutive-correct rule and separation from meaning progress remain intact.

All 48 reading word/reading pairs and glosses were checked in JMdict, with level
membership checked in JLPT Sensei's vocabulary lists. Accepted alternatives,
including uncommon readings of 甘い, 案内, 安心, and 安全, are excluded from
distractors. Offline regressions cover these alternatives and legacy-ID progress.
Browser checks covered both reading levels, stale loading responses, cancellation,
independent mode selections, saved review/retries, mixed-level queue clearing,
retired starter snapshots, loading errors/retry, and mobile/desktop layout.

## 5. Coverage-aware practice — complete (first version)

- Show attempted/available and unseen counts for the selected mode and level,
  with a progress bar, loading/error states, and manual refresh.
- Derive coverage from all saved completed attempts, including focused retries.
  Count each distinct current-pool question as attempted regardless of correctness;
  keep per-item attempt counts and latest completion time for balancing.
- Match stable IDs to the current catalog, so matching legacy snapshots count
  without rewriting history. Retired IDs do not inflate pool coverage.
- Select regular quizzes by unseen first, then fewest attempts, then oldest
  practice. Randomize exact ties, selected question order, and choices.
- Check fresh coverage when starting a full quiz, waiting for pending saves;
  cancelled preparation requests cannot start an unwanted late quiz.
- Re-load the coverage panel when returning to setup. Viewing history and
  cancelling/unsaved runs do not change coverage. Coverage is not mastery.

Verified 24-item coverage after three saved ten-question runs, isolation across
modes/levels, historical IDs, idempotent saves, more-than-twenty-attempt history,
database reopening, and count/recency/tie ordering. Browser checks covered direct
restarts while saving, cancellations, retry counts, saved review, load failures,
reload persistence, and 320px/desktop layouts.

### Later study refinements

- Optional "study these first" preview, distinct from testing recall.
- Fixed named sets for explicit study sessions alongside balanced short quizzes.
- Explore images as optional mnemonic aids, including their source/attribution;
  abstract and multi-meaning kanji may not suit a single literal image.

## 6. JLPT-style sentence practice — Exercise 4 complete; Anki generation later

- Added **JLPT Kanji exercise → Exercise 4**, using the twelve questions and
  source answer key supplied by the user. Six reading and six spelling tasks.
- Render highlighted/underlined targets inside their full sentences, preserving
  inflections. Full exercises retain original question order and shuffle choices.
- Save twelve-question attempts, task/target context, and exercise identity;
  review and retry from immutable snapshots. Show separate task scores.
- Coverage and the two-correct queue operate in sentence mode, independent of
  standalone meaning/reading progress. Twelve-item missed retries are supported.
- Migrate existing SQLite quiz rows transactionally without changing their JSON,
  scores, IDs, or parent references. Preserve the ten-question regular-quiz rules.
- Verified the source key, migration/rollback, separate task streaks, complete
  exercise flow, 0/12 and 12/12, cancellation, saved review, queue clearing,
  regular-mode regression, and mobile/desktop layout.

Additional exercises can be supplied as separate blocks in one text document or
paste. See [jlpt-kanji-exercises.md](jlpt-kanji-exercises.md) for the input format.

### Later Anki sentence generation

- Inspect `JP Sentences` fields and annotations before selecting a parser.
- Start with a small checked set: written word → reading, then reading → kanji.
- Preserve inflections and context, remove revealing furigana, and verify that
  exactly one option is correct in context.
- Evaluate question quality before generating across the whole deck.
