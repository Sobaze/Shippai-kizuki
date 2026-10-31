import { useState } from 'react'
import { Link } from 'react-router-dom'
import { AnkiConnection } from '../anki/AnkiConnection'
import { ReviewsPage } from '../reviews/ReviewsPage'
import { ScoringGuide } from './ScoringGuide'
import './home.css'

export function HomePage() {
  const [dataVersion, setDataVersion] = useState(0)
  const [historyOpened, setHistoryOpened] = useState(false)

  return (
    <>
      <header className="app-header">
        <p className="eyebrow">Mistake → Insight → Learning → Improvement</p>
        <h1>Shippai Kizuki</h1>
        <p>Turn your Japanese review history into a clearer picture of what to practice.</p>
        <Link className="home-insights-link" to="/insights/kanji">Explore your insights →</Link>
      </header>

      <AnkiConnection onImported={() => setDataVersion((version) => version + 1)} />
      <ScoringGuide />

      <details className="review-history" onToggle={(event) => {
        if (event.currentTarget.open) setHistoryOpened(true)
      }}>
        <summary>View recent individual reviews</summary>
        {historyOpened && <ReviewsPage key={dataVersion} />}
      </details>
    </>
  )
}
