# Checked quiz data

## Current pools

`backend/quiz_data.json` holds the original N5/N4 pools; the N2 expansion is in
`backend/quiz_n2.generated.json`. Contextual exercises are in `backend/quiz_exercises.json`.
Go embeds and merges these files with `//go:embed`,
validates the combined catalog at startup, and serves one selected pool with its source/version
metadata through `GET /quiz/questions?mode=meaning&level=N5`.

| Mode | API level | Questions | Contents |
| --- | --- | ---: | --- |
| Kanji meanings | `N2` | 374 | Full N2-specific source list; see [N2 data notes](n2-practice.md) |
| Word readings | `N2` | 1,640 | Independently classified, exact dictionary matches; see [N2 data notes](n2-practice.md) |
| JLPT Kanji exercise | `N2`, `exercise-4` | 12 | Fixed sentence reading/spelling tasks; see [exercise notes](jlpt-kanji-exercises.md) |
| Kanji meanings | `N5` | 24 | 日 月 年 人 国 山 川 水 火 木 金 土 大 小 上 下 中 外 東 西 南 北 書 読 |
| Kanji meanings | `N4` | 24 | 海 町 店 道 空 広 新 強 近 朝 社 建 会 同 自 開 明 動 体 思 家 安 売 住 |
| Word readings | `N5` | 24 | 新しい 新聞 勉強 近い 朝ご飯 浴びる 危ない 上げる 赤い 明るい 開ける 秋 甘い 洗う 歩く 遊ぶ 暖かい 番号 便利 帽子 病院 地下鉄 地図 電車 |
| Word readings | `N4` | 24 | 最近 案内 安心 安全 浅い 集まる 集める 謝る 番組 場所 美術館 貿易 部長 文学 文化 文法 地理 注意 注射 駐車場 動物園 道具 選ぶ 原因 |

Meaning levels are **estimated kanji classifications**, not classifications of
the example words. Reading levels are **estimated vocabulary classifications**,
checked separately. For example, 明 is in the N4 meaning pool while 明るい is in
the N5 reading pool. These are checked subsets, not complete or official JLPT
syllabuses. For either mode, choosing N4 does not mix in N5 questions.

N2 is the default in both modes. Its 374 kanji are the **N2-specific additions**,
not the cumulative prerequisite list for the exam. Its vocabulary pool excludes
kana-only/unsupported forms, duplicate prompts, and unmatched readings.
N2 sources, licensing, validation evidence, and the development-only rebuild
process are documented in [n2-practice.md](n2-practice.md).

## Sources and review

Data format version: **1**. Reading expansion checked: **2026-11-03**;
meaning batch checked: **2026-11-02**.

- Estimated membership: JLPT Sensei's [N5](https://jlptsensei.com/jlpt-n5-kanji-list/)
  and [N4](https://jlptsensei.com/jlpt-n4-kanji-list/) kanji study lists.
- Reading vocabulary membership: [N5 vocabulary](https://jlptsensei.com/jlpt-n5-vocabulary-list/)
  and [N4 vocabulary](https://jlptsensei.com/jlpt-n4-vocabulary-list/), not the
  kanji lists. All current N5 words are on page 1 except 新聞 on
  [page 6](https://jlptsensei.com/jlpt-n5-vocabulary-list/page/6/); all N4 words are
  on page 1 except 最近 on [page 4](https://jlptsensei.com/jlpt-n4-vocabulary-list/page/4/).
- Core kanji meanings: [kanjiapi.dev](https://kanjiapi.dev/), backed by
  [KANJIDIC2](https://www.edrdg.org/wiki/KANJIDIC_Project.html).
- Exact written-word/reading pairs and example glosses: JMdict through
  `https://kanjiapi.dev/v1/words/{character}`. Match the full written form and
  pronunciation, not just a component kanji's reading.
- Dictionary data attribution: [JMdict / EDRDG](https://www.edrdg.org/wiki/JMdict-EDICT_Dictionary_Project.html),
  [CC BY-SA 4.0](https://www.edrdg.org/edrdg/licence.html).
- The [official JLPT FAQ](https://www.jlpt.jp/e/faq/#anchor28) explains why current
  test-content lists are not published.

The original 48 meaning examples and 48 reading word snapshots were matched against dictionary entries. Meaning answers
and all offered distractors were inspected against the target kanji's meanings;
reading distractors were checked against dictionary variants so valid readings
are not marked wrong. Prompts ask for **a meaning**, since a kanji may have other
meanings. Questions/options are authored practice material, not copied JLPT tests.

Valid alternatives such as 甘い→うまい, 案内→あない, 安心→あんじん,
安全→あんせん, and 開ける→ひらける are not offered as wrong answers.
The quiz checks the one correct offered choice, not a claim that a word has only
one dictionary pronunciation.

## Stable IDs and historical starter questions

The original twelve N4 meaning IDs are preserved. Six starter reading targets
have verified classifications and retain their original IDs/options:

| Level | Word | Stable ID |
| --- | --- | --- |
| N5 | 新しい | `reading-new` |
| N5 | 新聞 | `reading-newspaper` |
| N5 | 勉強 | `reading-study` |
| N5 | 近い | `reading-near` |
| N5 | 朝ご飯 | `reading-breakfast` |
| N4 | 最近 | `reading-recently` |

海外, 海水, 強さ, 道路, 近道, and 早朝 were not assigned N5/N4 levels from these
vocabulary lists. They no longer appear in new full quizzes, but their saved
snapshots can still be reviewed, retried, and practised from the queue.
The `starter` pool endpoint is replaced by `reading&level=N5` / `N4`.

Only new snapshots receive level metadata; existing saved snapshots are not
rewritten. Queues span levels, keyed by question ID within the mode, so a correct
answer to a reclassified target advances its existing streak. Retired starter
items keep working from their snapshots. Results without level metadata use
"Kanji meanings" or "Word readings" rather than inventing a historic level.

## Coverage follows the current catalog

`GET /quiz/coverage?mode=reading&level=N5` counts saved attempts whose mode and
question IDs match the current pool. Level metadata in older snapshots is not
required. The response contains distinct attempted/unseen totals and, for every
current question, its attempt count and latest practice time (`null` if unseen).
All completed history contributes, including retries and attempts outside the
latest-twenty history view. Correctness does not affect exposure counts.

Adding a new ID makes a new unseen item; retaining an ID preserves exposure and
queue progress. Retiring an item removes it from current-pool coverage without
deleting its saved history or queue snapshot. Regular quiz selection prefers
unseen items, then fewer attempts, then older practice; exact ties are randomized.
Coverage shows exposure, not mastery or complete JLPT coverage.

## Adding a reviewed batch

1. Choose a target in the intended estimated-level list. Use the kanji list for
   meanings and the vocabulary list for readings. For a meaning question, check
   dictionary meanings and use a short, accurate core gloss as the answer.
2. Pick an example containing that kanji. Verify the exact written form, kana
   reading, and gloss in JMdict. The example word's difficulty need not equal the
   component kanji's estimated level.
3. Author three distinct distractors. Reject synonyms or other valid meanings of
   the target kanji. For reading questions, check alternative dictionary readings
   before using a kana string as a wrong answer.
4. Give a new item a stable, unique ID. Keep an existing ID for the same learning
   target; changing IDs would split its queue progress.
5. Update the check date and provenance when changing content. The version is the
   data format version; changing it requires corresponding loader support. Run
   `go test ./...` in `backend/` and `npm test` in `frontend/`.
6. Restart `go run .` to rebuild the embedded JSON and manually check a quiz in
   the affected pool. Historical review must still show the original snapshot.

Go checks supported pools, minimum size, unique IDs/prompts, one-kanji meaning
prompts, complete examples, kana readings, and four distinct choices containing
the answer. These structural checks do not replace semantic review of answers
and distractors. No external dictionary request is needed during quizzes/tests.
