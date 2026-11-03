# N2 exam-focused practice

N2 is selected by default in both quiz modes. The existing N4/N5 pools, saved
attempts, and progress remain available.

| Mode | N2 pool | Full quizzes for a first pass |
| --- | ---: | ---: |
| Kanji meanings | 374 distinct kanji | 38 |
| Word readings | 1,640 distinct written words | 164 |

Regular quizzes remain ten questions long. Unseen-first selection, saved
coverage, exact result snapshots, and the two-correct practice queue work with
these larger pools. Covering this content once in 30 days means roughly 13 new
kanji meanings and 55 new word readings daily, plus targeted retries. Coverage
records exposure; it does not establish retained recall or an exam score.

The kanji pool covers **all 374 N2-specific entries in JLPT Sensei's list** across
four pages. The source describes roughly 1,000 cumulative kanji for N2, including
N3/N4/N5 prerequisites. This pool is the N2-specific addition, not that cumulative
set and not an official syllabus.

## Sources, content, and validation

- Kanji classification: <https://jlptsensei.com/jlpt-n2-kanji-list/> and pages 2–4.
- Vocabulary classification: Stephen Kraus's `yomitan-jlpt-vocab`, revision
  `b062d4e38c4bdd0950ae1d4ec55f04b176182e03` (dictionary revision `2025.08.01.0`),
  based on Jonathan Waller's Tanos JLPT estimates. Source:
  <https://github.com/stephenmk/yomitan-jlpt-vocab/tree/b062d4e38c4bdd0950ae1d4ec55f04b176182e03>.
- Kanji glosses and exact example/reading pairs: KANJIDIC2 and JMdict through
  <https://kanjiapi.dev/>. Dictionary attribution: EDRDG / Jim Breen and the
  KANJIDIC/JMdict contributors, <https://www.edrdg.org/edrdg/licence.html>.
- Checked build date: **2026-11-03**. Classification estimates are independent:
  a kanji's level does not assign a vocabulary level to every word containing it.

The vocabulary source has **1,856 records**. The build retains only standalone
kanji-containing forms with supported kana readings, excludes existing reading
targets, and matches the **exact written form and pronunciation** in JMdict.
It produces 1,640 unique written prompts after excluding 212 ineligible records,
collapsing three additional readings of already-included prompts, and deferring
one unmatched pair (`嚙る / かじる`). It does not claim all N2 vocabulary is present.
Kana-only words are excluded because their written form would reveal the answer.

Meaning questions use dictionary glosses, with reviewed core-sense overrides
where dictionary ordering would foreground an obscure sense. Examples are
dictionary-matched, favour common entries, and include reviewed overrides such
as `仏教` for 仏 and `卒業` for 卒. Example words need not themselves be N2 vocabulary.
All meaning-choice sets are checked against the target's full dictionary glosses
and conservative synonym families; the full meaning pool was inspected during
the build. This is recognition practice, not a reproduction of JLPT questions.

Reading distractors use nearby kana changes (voicing, small kana, vowel changes,
and omitted long vowels), falling back to nearby readings in the vocabulary
pool. Every offered alternative is checked against **all dictionary readings of
the same written form**, including uncommon variants and equivalent じ/ぢ or
ず/づ spellings. Only one offered choice is accepted. Exact duplicates and
sound-equivalent alternatives are excluded. Choices remain stable in saved
snapshots and are shuffled when a quiz/retry begins.

## Bundled data and reproducible checks

- `backend/quiz_n2.generated.json`: generated offline question catalog, embedded
  by Go and merged with the original `quiz_data.json` at startup.
- `backend/quiz_n2.audit.json`: source membership, dictionary gloss/example
  evidence, accepted alternative readings, and exclusions. Used in offline
  data checks; it is not sent to the browser or embedded in the application.
- `backend/scripts/build-n2.mjs`: development-only data build, with a pinned
  vocabulary revision and a cached source fetcher. It generates only the N2
  catalog and audit artifacts; it does not change the original question pools.
- `backend/scripts/build-n2.test.mjs`: checks the complete generated content,
  membership, meaning-choice conflicts, and reading alternatives without network.

From `backend/`:

```bash
# Only when intentionally rebuilding/reviewing source data; requires internet.
node scripts/build-n2.mjs

# Offline data and application checks.
node --test scripts/build-n2.test.mjs
go test ./...
```

Source responses are cached outside the workspace in the OS temporary folder's
`opencode/shippai-n2-source-cache` directory. `N2_BUILD_CACHE` can override this
path. Rebuilding uses cached responses; remove that cache explicitly if a new
source review is intended. Review generated changes before adopting a rebuild.
Restart Go to load changed embedded data. Normal app startup and quizzes are
fully local and never run this script or request external dictionaries.

Kanji IDs use their Unicode code points; word IDs use stable written-form hashes.
Existing IDs are not renumbered when the source order changes. Old results remain
immutable; both new coverage and queue progress continue to use stable IDs.

## Focused queues and larger lists

The queue initially shows the **selected level**. Matching current-pool IDs also
include legacy snapshots without level metadata. Select **Include other levels
and older saved items** to return to the mode's entire saved queue. The two-correct
rule still applies across all saved attempts; filtering only changes the displayed
practice candidates. Long lists initially show 24 items and expand 24 at a time.

## Data attribution and licence

The generated N2 catalog and audit data are adaptations distributed under
**Creative Commons Attribution-ShareAlike 4.0 International**:
<https://creativecommons.org/licenses/by-sa/4.0/>.

Vocabulary source copyright: **© Stephen Kraus 2021–2025**, using Jonathan
Waller's JLPT resources (<https://www.tanos.co.uk/jlpt/>, originally CC BY).
Retain the source attribution and licence when redistributing these data.
Upstream licence and disclaimer:
<https://github.com/stephenmk/yomitan-jlpt-vocab/blob/b062d4e38c4bdd0950ae1d4ec55f04b176182e03/LICENSE.txt>.

Dictionary data: EDRDG, KANJIDIC2 and JMdict contributors, under CC BY-SA 4.0.
Adaptations here include filtering, selecting glosses/examples, core-sense edits,
stable question IDs, generating distractors, and adding verification evidence.
Source data and these adaptations are provided as-is without warranties or
endorsement by the source authors or the JLPT organisers.
