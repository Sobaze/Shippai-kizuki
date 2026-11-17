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
      <header className="app-header home-hero paper-surface">
        <div className="home-hero-copy">
          <p className="eyebrow"><span lang="ja">学びの路地</span> · A little practice, every day</p>
          <h1>Shippai Kizuki</h1>
          <p className="home-japanese-title" lang="ja"><span>今夜も、</span><span>ひとつずつ。</span></p>
          <p>Turn your Japanese review history into a clearer picture of what to practice.</p>
          <div className="home-hero-actions">
            <Link className="home-practice-link" to="/quiz">Start your practice <span aria-hidden="true">→</span></Link>
            <Link className="home-insights-link" to="/insights/kanji">Explore your insights →</Link>
          </div>
          <p className="home-learning-loop">Mistake → Insight → Learning → Improvement</p>
        </div>
        <figure className="home-alley" aria-hidden="true">
          <img src="/izakaya-alley.svg" alt="" width="760" height="720" />
          <figcaption lang="ja">小さな積み重ね。</figcaption>
        </figure>
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
