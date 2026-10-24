import type { Review } from './types'

export async function getReviews(signal: AbortSignal): Promise<Review[]> {
  const response = await fetch('/api/reviews', { signal })

  if (!response.ok) {
    throw new Error(`Could not load reviews (HTTP ${response.status}). Check that the Go backend is running.`)
  }

  return response.json()
}
