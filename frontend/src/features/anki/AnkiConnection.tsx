import { useEffect, useRef, useState } from 'react'
import { getAnkiDecks } from './api'
import { VocabularyImport } from './VocabularyImport'
import './anki.css'

type ConnectionState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; decks: string[]; checkedAt: Date }
  | { status: 'error'; message: string }

export function AnkiConnection({ onImported }: { onImported: () => void }) {
  const [state, setState] = useState<ConnectionState>({ status: 'idle' })
  const requestRef = useRef<AbortController | null>(null)

  useEffect(() => () => requestRef.current?.abort(), [])

  async function connect() {
    requestRef.current?.abort()
    const controller = new AbortController()
    requestRef.current = controller
    setState({ status: 'loading' })

    try {
      const result = await getAnkiDecks(controller.signal)
      if (!controller.signal.aborted) {
        setState({ status: 'success', decks: result.decks, checkedAt: new Date() })
      }
    } catch (error: unknown) {
      if (!controller.signal.aborted) {
        setState({
          status: 'error',
          message: error instanceof Error ? error.message : 'Could not connect to Anki.',
        })
      }
    }
  }

  return (
    <section className="anki-connection" aria-labelledby="anki-heading">
      <div className="anki-header">
        <div>
          <p className="anki-eyebrow">Local connection · AnkiConnect</p>
          <h2 id="anki-heading">Your Anki decks</h2>
        </div>
        <button type="button" onClick={connect} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Connecting…'
            : state.status === 'success' ? 'Refresh decks'
              : state.status === 'error' ? 'Retry connection' : 'Connect to Anki'}
        </button>
      </div>

      <p className="anki-note">
        Open Anki Desktop with AnkiConnect enabled. Sync your phone and desktop first
        to make the latest collection available here.
      </p>

      {state.status === 'idle' && <p className="anki-note">Connect to read the deck names from your active desktop profile.</p>}
      {state.status === 'loading' && <p role="status">Asking Anki for your decks…</p>}
      {state.status === 'error' && <p className="anki-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        <>
          <p className="anki-status" role="status">
            <span>Connected on last check</span>
            {' · '}{state.decks.length} {state.decks.length === 1 ? 'deck' : 'decks'}
            {' · '}{state.checkedAt.toLocaleTimeString()}
          </p>
          {state.decks.length === 0 ? (
            <p>Anki responded, but returned no decks.</p>
          ) : (
            <ul className="anki-decks" aria-label="Decks in Anki Desktop">
              {state.decks.map((deck) => <li key={deck}>{deck}</li>)}
            </ul>
          )}
          <p className="anki-note">Refresh after syncing Anki to update this list.</p>
        </>
      )}

      <VocabularyImport onImported={onImported} />
    </section>
  )
}
