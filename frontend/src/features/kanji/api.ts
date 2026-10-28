import type { KanjiDetails, KanjiStats } from './types'

// SQLite survives backend restarts; this small cache avoids another round trip
// and loading render when reopening an entry in the current browser session.
const dictionaryEntries = new Map<string, KanjiDetails>()
const dictionaryRequests = new Map<string, Promise<KanjiDetails>>()

export async function getKanjiStats(signal: AbortSignal): Promise<KanjiStats[]> {
  const response = await fetch('/api/kanji-stats', { signal })

  if (!response.ok) {
    throw new Error(`Could not load kanji insights (HTTP ${response.status}). Check that the Go backend is running with the latest code.`)
  }

  return response.json()
}

export function getCachedKanjiDetails(character: string): KanjiDetails | undefined {
  return dictionaryEntries.get(character)
}

export function getKanjiDetails(character: string): Promise<KanjiDetails> {
  const cached = dictionaryEntries.get(character)
  if (cached) return Promise.resolve(cached)

  const pending = dictionaryRequests.get(character)
  if (pending) return pending

  // Do not tie a shared dictionary request to one popup's lifetime. Completing
  // it warms both caches even if the popup closes or StrictMode reruns its effect.
  const request = fetchKanjiDetails(character)
    .then((details) => {
      dictionaryEntries.set(character, details)
      return details
    })
    .finally(() => dictionaryRequests.delete(character))

  dictionaryRequests.set(character, request)
  return request
}

export function prefetchKanjiDetails(character: string): void {
  void getKanjiDetails(character).catch(() => {
    // A failed preload must not block opening. The popup can retry and show errors.
  })
}

async function fetchKanjiDetails(character: string): Promise<KanjiDetails> {
  const response = await fetch(`/api/kanji/${encodeURIComponent(character)}`)

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Could not load the dictionary entry (HTTP ${response.status}).`)
  }

  return response.json()
}
