import { useEffect, useState } from 'react'
import { getKanjiStats, prefetchKanjiDetails } from './api'
import { KanjiDetail } from './KanjiDetail'
import { KanjiEvidenceTable } from './KanjiEvidenceTable'
import { recommendationLabels } from './labels'
import type { KanjiStats } from './types'
import './kanji.css'

type KanjiState =
  | { status: 'loading' }
  | { status: 'success'; kanji: KanjiStats[] }
  | { status: 'error'; message: string }

export function KanjiPage() {
  const [state, setState] = useState<KanjiState>({ status: 'loading' })
  const [reloadCount, setReloadCount] = useState(0)
  const [selectedKanji, setSelectedKanji] = useState<KanjiStats | null>(null)
  const [visibleCount, setVisibleCount] = useState(24)

  useEffect(() => {
    const controller = new AbortController()

    getKanjiStats(controller.signal)
      .then((kanji) => {
        if (!controller.signal.aborted) {
          setState({ status: 'success', kanji })
        }
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load kanji insights.',
          })
        }
      })

    return () => controller.abort()
  }, [reloadCount])

  function refreshKanji() {
    setState({ status: 'loading' })
    setReloadCount((count) => count + 1)
  }

  return (
    <section className="kanji-insights" aria-labelledby="kanji-heading">
      <div className="kanji-header">
        <div>
          <p className="kanji-eyebrow">Patterns across vocabulary</p>
          <h2 id="kanji-heading">Kanji worth a closer look</h2>
        </div>
        <button type="button" onClick={refreshKanji} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Loading…' : 'Refresh kanji'}
        </button>
      </div>

      <p className="kanji-note">
        Difficulty in words containing a kanji is a study clue, not proof that the kanji
        caused a failed review. Evidence comes from your saved vocabulary answer history.
      </p>

      <details className="kanji-rules">
        <summary>How do scores and evidence work?</summary>
        <p>
          Each distinct word with at least 3 reviews contributes its vocabulary difficulty
          equally. The kanji score is their rounded average. Early evidence stays visible
          but does not affect the score.
        </p>
        <p>
          Suggested practice requires at least 2 qualifying words and a score of 50 or more.
          With no qualifying words, there is no score yet. These are experimental rules;
          reviews are not weighted by age.
        </p>
      </details>

      {state.status === 'loading' && <p role="status">Looking for kanji patterns…</p>}
      {state.status === 'error' && <p className="kanji-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        state.kanji.length === 0 ? (
          <p>No kanji evidence yet. Import JP Vocabulary above to analyze its answer history.</p>
        ) : (
          <>
            <p className="kanji-note">Showing {Math.min(visibleCount, state.kanji.length)} of {state.kanji.length.toLocaleString()} kanji</p>
            <div className="kanji-grid">
              {state.kanji.slice(0, visibleCount).map((entry) => (
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
                    onClick={() => setSelectedKanji(entry)}
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
            {visibleCount < state.kanji.length && (
              <button type="button" className="kanji-more" onClick={() => setVisibleCount((count) => count + 24)}>Show more kanji</button>
            )}
          </>
        )
      )}
      {selectedKanji !== null && (
        <KanjiDetail
          key={selectedKanji.kanji}
          entry={selectedKanji}
          onClose={() => {
            const buttonId = `kanji-open-${selectedKanji.kanji}`
            setSelectedKanji(null)
            requestAnimationFrame(() => document.getElementById(buttonId)?.focus())
          }}
        />
      )}
    </section>
  )
}
