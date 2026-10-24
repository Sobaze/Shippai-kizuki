import type { VocabularyStats } from './types'

export async function getVocabularyStats(signal: AbortSignal): Promise<VocabularyStats[]> {
  const response = await fetch('/api/vocabulary-stats', { signal })

  if (!response.ok) {
    throw new Error(`Could not load vocabulary scores (HTTP ${response.status}). Check that the Go backend is running.`)
  }

  return response.json()
}
