# JLPT Kanji exercise

The third practice type is named **JLPT Kanji exercise**. Its first set is
**Exercise 4**, adapted from the twelve questions and answer key supplied by the
user from JapaneseTest4You:

<https://japanesetest4you.com/japanese-language-proficiency-test-jlpt-n2-kanji-exercise-4/>

The supplied publication date is March 14, 2015. This is attributed third-party,
JLPT-inspired practice; the date does not establish an official 2015 past paper.
The website returned HTTP 403 during implementation, so the supplied text is the
source used here. Source sentence wording and choices are retained, including
awkward wording in question 8; the added feedback explains the target expression.

## Exercise 4

- Twelve fixed questions, in their original order, with shuffled answer choices.
- Six **reading** tasks: choose the kana reading of a kanji expression.
- Six **spelling** tasks: choose the written form that fits a kana expression.
- Show the complete sentence, with the target highlighted and underlined.
- Withhold feedback until all twelve questions are answered.
- Results show the total score and separate reading/spelling scores, followed by
  the correct sentence, answer, target reading, and meaning.
- Preserve inflections: 迫られた → せまられた and 積み → つみ.
- Retrying misses can contain 1–12 questions. Queue sessions select at most ten
  waiting items, as before. Each target has its own reading/spelling task ID;
  a reading success does not clear a different spelling target.
- Full restarts repeat the complete exercise, even after coverage reaches 12/12.
  Unseen-first selection remains the rule for regular meaning/reading quizzes.
- Completed exercises and retries save immutable sentence/target snapshots.
  Cancelling does not save a partial run or change coverage/streaks.

The exercise's N2 label applies to the set. It is not a claim that every target
word has an independently verified N2 vocabulary classification. Sentence practice
has its own history/coverage/queue, separate from standalone meanings/readings.

## Adding later exercises

Paste several exercises consecutively or provide a `.txt` / `.md` document.
The current paste format is sufficient. Keep these parts together for each set:

```text
Exercise 5
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
Exercise 6
Source URL: ...
...
```

Clear exercise headings, complete sentences/choices, and complete answer keys
are more important than special formatting. We can review and add one exercise
at a time from the document. This is a content-input convention, not an automatic
upload/import feature. An exercise selector can be introduced when another set
is added; the current setup opens Exercise 4 directly.

## Implementation

`backend/quiz_exercises.json` stores this manually supplied exercise separately
from generated N2 vocabulary. It is embedded and merged into the catalog at startup.
Pools have an `exerciseId` and `title`. Questions have `task`, `exerciseId`, and a
`context` containing `before`/`after`; the target stays in `prompt`. Splitting the
sentence around the target avoids unreliable automatic word alignment and allows
the inflected target to be preserved exactly.

Routes:

- `GET /quiz/questions?mode=sentence&level=N2&exercise=exercise-4`
- `GET /quiz/coverage?mode=sentence&level=N2&exercise=exercise-4`
- Existing save/history/practice routes accept `mode=sentence`.

Go validates sentence task/target relationships, complete full-exercise length,
and answer choices. SQLite now accepts sentence mode and totals up to twelve.
For existing databases, startup copies old quiz rows unchanged into the expanded
schema inside a transaction, including parent-attempt references and original
JSON snapshots. Failed migrations roll back. Anki review/dictionary tables are
unaffected. Restart Go after changing the embedded exercise data.

Offline checks verify the supplied answer-key positions, all twelve targets,
separate task progress, twelve-item retries, old-row preservation/idempotence,
and rollback on migration failure. Browser checks cover cancellation, full
0/12 and 12/12 results, task feedback, queue clearing, history/restart, and
mobile/desktop layouts.

## Later Anki sentence integration

Use this format as the first working example. Before generating similar tasks
from an Anki sentence deck, inspect its actual fields/annotations, identify the
target and full inflected reading, remove revealing furigana, and verify one
correct offered answer in the sentence. No sentence-deck parser or generator is
introduced by this first exercise.
