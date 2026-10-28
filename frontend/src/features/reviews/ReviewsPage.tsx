import { useEffect, useState } from 'react'
import { getReviews } from './api'
import type { Review, ReviewRating } from './types'
import './reviews.css'

type ReviewsState =
  | { status: 'loading' }
  | { status: 'success'; reviews: Review[] }
  | { status: 'error'; message: string }

const ratingLabels: Record<ReviewRating, string> = {
  again: 'Again',
  hard: 'Hard',
  good: 'Good',
  easy: 'Easy',
}

export function ReviewsPage() {
  const [state, setState] = useState<ReviewsState>({ status: 'loading' })
  const [reloadCount, setReloadCount] = useState(0)

  useEffect(() => {
    const controller = new AbortController()

    getReviews(controller.signal)
      .then((reviews) => {
        if (!controller.signal.aborted) {
          setState({ status: 'success', reviews })
        }
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load reviews.',
          })
        }
      })

    return () => controller.abort()
  }, [reloadCount])

  function reloadReviews() {
    setState({ status: 'loading' })
    setReloadCount((count) => count + 1)
  }

  return (
    <section className="reviews" aria-labelledby="reviews-heading">
      <div className="reviews-header">
        <div>
          <h2 id="reviews-heading">Vocabulary reviews</h2>
          <p>Individual Anki-style rating events used to calculate vocabulary scores.</p>
        </div>
        <button type="button" onClick={reloadReviews} disabled={state.status === 'loading'}>
          {state.status === 'loading' ? 'Loading…' : 'Refresh'}
        </button>
      </div>

      <p className="reviews-note">
        Saved Anki and manually recorded answers. This list shows the latest 100;
        difficulty scores use the complete saved history.
      </p>

      {state.status === 'loading' && <p role="status">Loading reviews from the backend…</p>}
      {state.status === 'error' && <p className="reviews-error" role="alert">{state.message}</p>}
      {state.status === 'success' && (
        state.reviews.length === 0 ? (
          <p>No reviews recorded yet.</p>
        ) : (
          <div className="reviews-table-container">
            <table className="reviews-table">
              <caption>Latest {state.reviews.length} answers · newest first</caption>
              <thead>
                <tr>
                  <th scope="col">Vocabulary</th>
                  <th scope="col">Rating</th>
                  <th scope="col">Source</th>
                  <th scope="col">Reviewed</th>
                </tr>
              </thead>
              <tbody>
                {state.reviews.map((review, index) => (
                  <tr key={index}>
                    <td lang="ja">{review.vocabulary}</td>
                    <td>
                      <span className={`rating rating-${review.rating}`}>
                        {ratingLabels[review.rating]}
                      </span>
                    </td>
                    <td>{review.source === 'anki' ? 'Anki' : 'Manual'}</td>
                    <td>{review.reviewedAt ? new Date(review.reviewedAt).toLocaleString() : '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}
    </section>
  )
}
