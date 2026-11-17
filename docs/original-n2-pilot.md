# Original N2 practice pilot

Open **Quiz → JLPT practice** and choose **Kanji exercises** or
**Reading comprehension**. Original sets are grouped first in the selector;
the eleven user-supplied kanji exercises remain under “Supplied exercise sets”.

| Set | Focus | Questions |
| --- | --- | ---: |
| Original Kanji 1 — Work and decisions | Three readings + three spellings | 6 |
| Original Kanji 2 — Results and cooperation | Three readings + three spellings | 6 |
| Original Reading 1 — Rethinking meetings | Main point and cause/reason | 2 |
| Original Reading 2 — Beyond the numbers | Writer's position and implied purpose | 2 |
| Original Reading 3 — Course information | Eligibility, deadlines, and conditions | 2 |

These are original, untimed, JLPT-inspired sets targeting N2 practice. They are
a small pilot for feedback on difficulty and presentation, not official exam
questions or a full mock exam. The notice describes a fictional course.

## Content choices

All twelve kanji targets are in the existing estimated-N2 vocabulary source:
対策, 能率, 見直す, 方針, 応対, 工夫, 比較的, 実績, 妥当, 調整, 活躍, 応援.
Their source word/reading pairs are checked offline against `n2.audit.json`.
Sentences and explanations are newly written for the app. Spelling choices can
include real homophones, such as 調整 / 調製, where the sentence selects the
appropriate meaning.

Reading questions and options are in Japanese. The passages use connected
argument, contrast, and practical conditions rather than isolated vocabulary.
Each reading set has two questions sharing the same complete passage, with four
distinct options and one selected answer. The six comprehension keys, before
choice shuffling, are `[2,3]`, `[1,4]`, and `[2,4]`.

Every original question includes an English explanation. Comprehension feedback
also quotes supporting Japanese text verbatim from the passage and offers a
“Read passage again” disclosure. Explanations/evidence are not rendered during
the active questions. Reading sets use a long-form layout and a single-column
results list; word-level practice retains its larger kanji-choice presentation.

## Progress and review

- Complete the whole selected set: six or two questions, respectively.
- Full restarts preserve authored question order and shuffle choices.
- Saved snapshots include passages, evidence, explanations, and chosen answers.
- Review restores the source practice category and set as the restart target.
- Coverage and queues focus on the selected set; the all-exercise toggle includes
  other named sets. The two-consecutive-correct rule still works per question ID.
- Cancelling does not save a partial result. Retries have their own score.

## Implementation and checks

`backend/internal/quiz/data/original-n2.json` is the authored source, embedded at
startup with the other catalogs. Pools have `category: "kanji"` or
`"comprehension"` and `original: true`. The exercise list supports
`GET /quiz/exercises?level=N2&category=comprehension` (or `kanji`) and includes
category, origin, size, and task-count metadata.

Comprehension snapshots use `task: "comprehension"`, a `passage` with `title` and
`text`, a question in `prompt`, `choices`/`answer`, and `explanation`/`evidence`.
They do not use word-target context or word examples; the existing API's word
example object is empty for this task. Validation requires evidence to occur in
the passage. New fields are omitted from old word snapshots when unused.

Named practice continues to use the existing stored `sentence` mode, so these
small sets fit the current schema and need no new database migration. Category
separation is explicit in the exercise list and UI.

From `backend/`, run `go test ./...` and
`node --test scripts/original-n2.test.mjs`; from `frontend/`, run `npm test`.
Checks cover N2 target membership, keys, complete passages/evidence, input
validation, actual set lengths, immutable feedback snapshots, grading/retries,
and category filtering. Browser checks cover all five sets at 320px/1440px,
feedback timing, saved category restoration, independent queues, clearing,
cancellation, and old Exercise 4 compatibility.
