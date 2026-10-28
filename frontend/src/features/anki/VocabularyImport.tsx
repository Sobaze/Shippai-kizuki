import { useEffect, useRef, useState } from 'react'
import { getVocabularyImport, importVocabulary } from './api'
import type { VocabularyImportSummary } from './types'

type ImportState =
  | { status: 'loading' }
  | { status: 'ready'; summary: VocabularyImportSummary }
  | { status: 'importing' }
  | { status: 'error'; message: string }

export function VocabularyImport({ onImported }: { onImported: () => void }) {
  const [state, setState] = useState<ImportState>({ status: 'loading' })
  const requestRef = useRef<AbortController | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    requestRef.current = controller
    getVocabularyImport(controller.signal)
      .then((summary) => {
        if (!controller.signal.aborted) setState({ status: 'ready', summary })
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setState({ status: 'error', message: error instanceof Error ? error.message : 'Could not load import status.' })
        }
      })
    return () => requestRef.current?.abort()
  }, [])

  async function runImport() {
    requestRef.current?.abort()
    const controller = new AbortController()
    requestRef.current = controller
    setState({ status: 'importing' })
    try {
      const summary = await importVocabulary(controller.signal)
      if (!controller.signal.aborted) {
        setState({ status: 'ready', summary })
        onImported()
      }
    } catch (error: unknown) {
      if (!controller.signal.aborted) {
        setState({ status: 'error', message: error instanceof Error ? error.message : 'Import failed.' })
      }
    }
  }

  return (
    <section className="anki-import" aria-labelledby="anki-import-heading">
      <div className="anki-header">
        <div>
          <h3 id="anki-import-heading">Use JP Vocabulary for your insights</h3>
          <p className="anki-note">Read cards and answer history from Anki, then save a local snapshot in SQLite.</p>
        </div>
        <button type="button" onClick={runImport} disabled={state.status === 'loading' || state.status === 'importing'}>
          {state.status === 'importing' ? 'Importing…' : state.status === 'error' ? 'Retry import' : 'Import JP Vocabulary'}
        </button>
      </div>

      {state.status === 'loading' && <p role="status">Checking saved import…</p>}
      {state.status === 'importing' && <p role="status">Reading cards and review history. The previous saved snapshot remains available until the import succeeds.</p>}
      {state.status === 'error' && <p className="anki-import-error" role="alert">{state.message}</p>}
      {state.status === 'ready' && (
        state.summary.importedAt === null ? (
          <p className="anki-note">No JP Vocabulary data imported yet. Import it to populate vocabulary and kanji scores.</p>
        ) : (
          <div className="anki-import-summary" role="status">
            <p><strong>{state.summary.cardCount.toLocaleString()} cards</strong> · {state.summary.reviewCount.toLocaleString()} scored answer events saved</p>
            <p className="anki-note">Last imported: {new Date(state.summary.importedAt).toLocaleString()}</p>
            {state.summary.missingReadingCount > 0 && <p className="anki-note">{state.summary.missingReadingCount} cards have no confidently extracted reading.</p>}
            {state.summary.unscoredReviewCount > 0 && <p className="anki-note">{state.summary.unscoredReviewCount} non-answer records retained but excluded from scoring.</p>}
          </div>
        )
      )}
      <p className="anki-note">Re-import after syncing your phone and desktop. It replaces this deck’s saved snapshot without duplicating its history. Anki itself is read-only.</p>
    </section>
  )
}
