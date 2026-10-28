import { useEffect, useRef, useState } from 'react'
import { getCachedKanjiDetails, getKanjiDetails } from './api'
import { KanjiEvidenceTable } from './KanjiEvidenceTable'
import { recommendationLabels } from './labels'
import type { KanjiDetails, KanjiStats } from './types'
import './kanji-detail.css'

type KanjiDetailProps = {
  entry: KanjiStats
  onClose: () => void
}

type DictionaryState =
  | { status: 'loading' }
  | { status: 'success'; details: KanjiDetails }
  | { status: 'error'; message: string }

function ReadingGroup({ title, readings }: { title: string; readings: string[] }) {
  return (
    <section className="kanji-reading-group">
      <h3>{title}</h3>
      {readings.length === 0 ? (
        <p>No readings listed.</p>
      ) : (
        <ul className="kanji-readings" lang="ja">
          {readings.map((reading) => <li key={reading}>{reading}</li>)}
        </ul>
      )}
    </section>
  )
}

export function KanjiDetail({ entry, onClose }: KanjiDetailProps) {
  const [state, setState] = useState<DictionaryState>(() => {
    const cached = getCachedKanjiDetails(entry.kanji)
    return cached ? { status: 'success', details: cached } : { status: 'loading' }
  })
  const [retryCount, setRetryCount] = useState(0)
  const dialogRef = useRef<HTMLDialogElement>(null)
  const headingRef = useRef<HTMLHeadingElement>(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return

    const previousOverflow = document.body.style.overflow
    dialog.showModal()
    headingRef.current?.focus()
    document.body.style.overflow = 'hidden'

    return () => {
      dialog.close()
      document.body.style.overflow = previousOverflow
    }
  }, [])

  useEffect(() => {
    let active = true

    getKanjiDetails(entry.kanji)
      .then((details) => {
        if (active) {
          setState({ status: 'success', details })
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load the dictionary entry.',
          })
        }
      })

    return () => { active = false }
  }, [entry.kanji, retryCount])

  function retryLookup() {
    setState({ status: 'loading' })
    setRetryCount((count) => count + 1)
  }

  return (
    <dialog
      className="kanji-detail"
      ref={dialogRef}
      aria-labelledby="kanji-detail-heading"
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
      onClick={(event) => {
        if (event.target !== event.currentTarget) return
        const bounds = event.currentTarget.getBoundingClientRect()
        if (event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom) {
          onClose()
        }
      }}
    >
      <header className="kanji-modal-toolbar">
        <h2 id="kanji-detail-heading" ref={headingRef} tabIndex={-1}>Study <span lang="ja">{entry.kanji}</span></h2>
        <button type="button" className="kanji-close" onClick={onClose} aria-label="Close kanji details">Close ×</button>
      </header>

      <div className="kanji-modal-content">
        <div className="kanji-study-layout">
          <div className="kanji-study-character">
            <span className="kanji-study-glyph" lang="ja">{entry.kanji}</span>
            {state.status === 'success' && (
              <div className="kanji-meanings">
                <h3>Meaning</h3>
                <p>{state.details.meanings.length > 0 ? state.details.meanings.join('; ') : 'No English meanings listed.'}</p>
                <small>{state.details.strokeCount} strokes</small>
              </div>
            )}
            <div className="kanji-score">
              {entry.difficulty === null ? (
                <strong className="kanji-unscored">Not scored yet</strong>
              ) : (
                <strong>{entry.difficulty}<small> / 100</small></strong>
              )}
              <span>Your recorded difficulty</span>
            </div>
          </div>

          <div className="kanji-reading-stack">
            {state.status === 'loading' && <p role="status">Looking up meanings and readings…</p>}
            {state.status === 'error' && (
              <div>
                <p className="kanji-error" role="alert">{state.message}</p>
                <button type="button" className="kanji-retry" onClick={retryLookup}>Retry dictionary lookup</button>
              </div>
            )}
            {state.status === 'success' && (
              <>
                <ReadingGroup title="On’yomi · 音読み" readings={state.details.onReadings} />
                <ReadingGroup title="Kun’yomi · 訓読み" readings={state.details.kunReadings} />
                {state.details.nameReadings.length > 0 && (
                  <details className="kanji-name-readings">
                    <summary>Name readings · 名乗り</summary>
                    <ul className="kanji-readings" lang="ja">
                      {state.details.nameReadings.map((reading) => <li key={reading}>{reading}</li>)}
                    </ul>
                  </details>
                )}
                <details className="kanji-notation">
                  <summary>Reading notation help</summary>
                  <p>A dot marks where the kana ending (okurigana) begins, such as か.く for 書く. A hyphen indicates a prefix or suffix form.</p>
                  <p>These are dictionary-listed readings. A word’s actual reading depends on that word.</p>
                </details>
              </>
            )}
          </div>
        </div>

        <section className="kanji-evidence kanji-detail-evidence" aria-labelledby="kanji-detail-evidence-heading">
          <div className="kanji-practice-heading">
            <h3 id="kanji-detail-evidence-heading">Practice with your vocabulary</h3>
            <span className={`kanji-recommendation kanji-${entry.recommendation}`}>
              {recommendationLabels[entry.recommendation]}
            </span>
          </div>
          <p className="kanji-context">Try recalling each word’s reading, then reveal its furigana. Readings come from imported card fields; uncertain readings are left unavailable.</p>
          <p className="kanji-context">{entry.qualifiedWordCount} of {entry.wordCount} vocabulary items qualify for scoring.</p>
          <KanjiEvidenceTable kanji={entry.kanji} evidence={entry.evidence} practiceReadings />
        </section>

        <footer className="kanji-attribution">
          Kanji dictionary data: <a href="https://www.edrdg.org/wiki/KANJIDIC_Project.html" target="_blank" rel="noreferrer">KANJIDIC2 / EDRDG</a>
          {' '}(<a href="https://www.edrdg.org/edrdg/licence.html" target="_blank" rel="noreferrer">CC BY-SA 4.0</a>),
          {' '}via <a href={`https://kanjiapi.dev/v1/kanji/${encodeURIComponent(entry.kanji)}`} target="_blank" rel="noreferrer">kanjiapi.dev</a>.
        </footer>
      </div>
    </dialog>
  )
}
