# JLPT Kanji exercise

The supplied **JLPT Kanji exercises** are under **Quiz → JLPT practice → Kanji exercises**.
The supplied bank offers **Exercises 1–11**, adapted from the passages, choices, and answer keys supplied
by the user from JapaneseTest4You. The original reference for Exercise 4 is:

<https://japanesetest4you.com/japanese-language-proficiency-test-jlpt-n2-kanji-exercise-4/>

The supplied publication date is March 14, 2015. This is attributed third-party,
JLPT-inspired practice; the date does not establish an official 2015 past paper.
The website returned HTTP 403 during implementation, so the supplied text is the
source used here. Source passage wording and answer keys are retained, including
awkward phrasing in the supplied material; added feedback explains each target.
Whitespace/blank-line artifacts in the choices are removed.

## Available sets and behaviour

| Exercise | Questions | Readings | Spellings |
| --- | ---: | ---: | ---: |
| 1 | 12 | 6 | 6 |
| 2 | 12 | 6 | 6 |
| 3 | 12 | 6 | 6 |
| 4 | 12 | 6 | 6 |
| 5 | 12 | 6 | 6 |
| 6 | 12 | 5 | 7 |
| 7 | 10 | 7 | 3 |
| 8 | 10 | 5 | 5 |
| 9 | 10 | 4 | 6 |
| 10 | 10 | 10 | 0 |
| 11 | 10 | 7 | 3 |
| **Total** | **122** | **68** | **54** |

- Choose a named set in the selector. The supplied sets retain numerical order
  under their group; original pilot sets are listed first. Reviewing a saved set
  restores its exercise selection.
- Each set keeps its actual length and original question order, with shuffled choices.
- Reading tasks ask for a kana reading; spelling tasks ask for the written form.
- Every target is a separate question. Shared passages are shown in full each
  time, highlighting and underlining only the expression currently requested.
- Withhold feedback until the selected exercise is complete, whether 10 or 12 questions.
- Results show the total score and separate reading/spelling scores, followed by
  the correct sentence, answer, target reading, and meaning.
- Preserve inflections: 迫られた → せまられた and 積み → つみ.
- Retrying misses can contain 1–12 questions. Queue sessions select at most ten
  waiting items, as before. Each target has its own reading/spelling task ID;
  a reading success does not clear a different spelling target.
- Full restarts repeat the selected exercise, even after its coverage is complete.
  Unseen-first selection remains the rule for regular meaning/reading quizzes.
- Completed exercises and retries save immutable sentence/target snapshots.
  Cancelling does not save a partial run or change coverage/streaks.

Coverage and the default queue focus on the selected exercise. **Include other
exercises and older saved items** shows the whole sentence-practice queue.
Progress in one set does not clear another set's targets. Source kana choices,
including the source's じしん/ぢしん spelling distinction, retain the supplied key.

The exercise's N2 label applies to the set. It is not a claim that every target
word has an independently verified N2 vocabulary classification. Sentence practice
has its own history/coverage/queue, separate from standalone meanings/readings.

## Adding later exercises

Paste several exercises consecutively or provide a `.txt` / `.md` document.
The current paste format is sufficient. Keep these parts together for each set:

```text
Exercise 12
Source URL: https://...

Sentence ... target ...
1. target
choice 1
choice 2
choice 3
choice 4

[remaining questions]

Answer key:
1. correct option (target => answer)
...

---
Exercise 13
Source URL: ...
...
```

Clear exercise headings, complete sentences/choices, and complete answer keys
are more important than special formatting. We can review and add one exercise
at a time from the document. This is a content-input convention, not an automatic
upload/import feature. New catalog entries appear in the selector through the
backend's exercise-list endpoint.

## Implementation

`backend/internal/quiz/data/exercises.json` retains the original Exercise 4.
`exercises.supplied.json` stores the ten newly supplied sets with shared passages
stored once. Each question names its passage by zero-based index and its target.
If the target occurs more than once, an explicit one-based `occurrence` chooses
which instance to highlight (Exercise 8's いそがしく uses its first occurrence).

`backend/scripts/build-exercises.mjs` expands these into
`exercises.additional.generated.json`. This is an offline content build, not a
web scraper. It checks source numbering, four unique choices, answer positions,
target/feedback consistency, and exact target occurrence before writing output.
The original Exercise 4 file and stable IDs are preserved.

Go embeds both exercise catalogs and merges them with the regular quiz pools.
Pools have an `exerciseId` and `title`. Questions have `task`, `exerciseId`, and a
`context` containing `before`/`after`; the target stays in `prompt`. Splitting the
sentence around the target avoids unreliable automatic word alignment and allows
the inflected target to be preserved exactly.

Routes:

- `GET /quiz/exercises?level=N2` lists IDs/titles, actual lengths, and task counts
  without returning question/answer content.
- `GET /quiz/questions?mode=sentence&level=N2&exercise=exercise-4`
- `GET /quiz/coverage?mode=sentence&level=N2&exercise=exercise-4`
- Existing save/history/practice routes accept `mode=sentence`.

Go validates sentence task/target relationships, complete full-exercise length,
and answer choices. SQLite now accepts sentence mode and totals up to twelve.
For existing databases, startup copies old quiz rows unchanged into the expanded
schema inside a transaction, including parent-attempt references and original
JSON snapshots. Failed migrations roll back. Anki review/dictionary tables are
unaffected. Restart Go after changing the embedded exercise data.

Exercise length validation uses the registered set's actual length. No new
schema migration is required to add these 10/12-question sets. Embedded exercise
lengths are decoded once for validation, rather than re-reading every passage for
each answer.

From `backend/` after editing the supplied source data:

```bash
node scripts/build-exercises.mjs
node --test scripts/build-exercises.test.mjs
go test ./...
```

Offline checks verify the supplied answer-key positions, all twelve targets,
separate task progress, twelve-item retries, old-row preservation/idempotence,
and rollback on migration failure. Browser checks cover cancellation, full
0/12 and 12/12 results, task feedback, queue clearing, history/restart, and
mobile/desktop layouts.

The expanded-bank checks independently transcribe all eleven source answer keys,
check all 122 targets and full shared-passage reconstruction, and verify totals
and task distributions. Browser checks complete every set, verify 10-question
and 12-question saving, reading-only results, selector ordering/stale responses,
Exercise 4 history/restart, independent queues, retries, and list-load recovery.

## Later Anki sentence integration

Use this format as the first working example. Before generating similar tasks
from an Anki sentence deck, inspect its actual fields/annotations, identify the
target and full inflected reading, remove revealing furigana, and verify one
correct offered answer in the sentence. No sentence-deck parser or generator is
introduced by this first exercise.
