import { useEffect, useState } from 'react'
import { prefetchKanjiDetails } from './api'
import { KanjiDetail } from './KanjiDetail'
import { KanjiEvidenceTable } from './KanjiEvidenceTable'
import { recommendationLabels } from './labels'
import type { KanjiStats } from './types'

export function KanjiGrid({ entries }: { entries: KanjiStats[] }) {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)
  const selectedKanji = selectedIndex === null ? undefined : entries[selectedIndex]

  useEffect(() => {
    if (selectedIndex === null) return
    for (const index of [selectedIndex - 1, selectedIndex + 1]) {
      if (entries[index]) prefetchKanjiDetails(entries[index].kanji)
    }
  }, [entries, selectedIndex])

  return (
    <>
      <div className="kanji-grid">
        {entries.map((entry, index) => (
          <article className="kanji-card" key={entry.kanji} aria-label={`Kanji ${entry.kanji}`}>
            <div className="kanji-card-header">
              <span className="kanji-character" lang="ja">{entry.kanji}</span>
              <div className="kanji-score">
                {entry.difficulty === null ? (
                  <strong className="kanji-unscored">Not scored yet</strong>
                ) : (
                  <strong>{entry.difficulty}<small> / 100</small></strong>
                )}
                <span>Difficulty</span>
              </div>
            </div>

            <p className={`kanji-recommendation kanji-${entry.recommendation}`}>
              {recommendationLabels[entry.recommendation]}
            </p>
            <p className="kanji-evidence-count">
              {entry.wordCount} vocabulary {entry.wordCount === 1 ? 'item' : 'items'}
              {' · '}{entry.qualifiedWordCount} qualifying
            </p>
            <p className="kanji-context">
              {entry.qualifiedWordCount === 0
                ? 'More reviews are needed before any word can contribute a score.'
                : entry.qualifiedWordCount === 1
                  ? 'The score comes from one word. Other vocabulary contexts are needed to investigate a wider pattern.'
                  : 'A pattern across multiple words. Open the evidence to see what contributes.'}
            </p>

            <button
              type="button"
              className="kanji-open"
              id={`kanji-open-${entry.kanji}`}
              onClick={() => setSelectedIndex(index)}
              onPointerEnter={() => prefetchKanjiDetails(entry.kanji)}
              onFocus={() => prefetchKanjiDetails(entry.kanji)}
              aria-label={`Open meanings and readings for ${entry.kanji}`}
            >
              Meanings &amp; readings →
            </button>

            <details className="kanji-evidence">
              <summary>Inspect contributing words</summary>
              <KanjiEvidenceTable kanji={entry.kanji} evidence={entry.evidence} />
            </details>
          </article>
        ))}
      </div>
      {selectedKanji !== undefined && selectedIndex !== null && (
        <KanjiDetail
          entry={selectedKanji}
          position={selectedIndex + 1}
          total={entries.length}
          onPrevious={() => setSelectedIndex((index) => index === null ? null : Math.max(0, index - 1))}
          onNext={() => setSelectedIndex((index) => index === null ? null : Math.min(entries.length - 1, index + 1))}
          onClose={() => {
            const buttonId = `kanji-open-${selectedKanji.kanji}`
            setSelectedIndex(null)
            requestAnimationFrame(() => document.getElementById(buttonId)?.focus())
          }}
        />
      )}
    </>
  )
}
