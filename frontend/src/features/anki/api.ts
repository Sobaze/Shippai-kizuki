import type { AnkiDeckList, VocabularyImportSummary } from './types'

export async function getAnkiDecks(signal: AbortSignal): Promise<AnkiDeckList> {
  const response = await fetch('/api/anki/decks', { signal })

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Could not connect to Anki (HTTP ${response.status}).`)
  }

  return response.json()
}

export async function getVocabularyImport(signal: AbortSignal): Promise<VocabularyImportSummary> {
  const response = await fetch('/api/anki/import-status', { signal })
  if (!response.ok) {
    throw new Error((await response.text()).trim() || 'Could not load saved import status.')
  }
  return response.json()
}

export async function importVocabulary(signal: AbortSignal): Promise<VocabularyImportSummary> {
  const response = await fetch('/api/anki/import-vocabulary', { method: 'POST', signal })
  if (!response.ok) {
    throw new Error((await response.text()).trim() || 'Could not import JP Vocabulary.')
  }
  return response.json()
}
