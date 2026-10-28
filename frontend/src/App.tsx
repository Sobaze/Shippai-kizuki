import { useState } from 'react'
import { ReviewsPage } from './features/reviews/ReviewsPage'
import { VocabularyPage } from './features/vocabulary/VocabularyPage'
import { KanjiPage } from './features/kanji/KanjiPage'
import { AnkiConnection } from './features/anki/AnkiConnection'
import './App.css'

function App() {
  const [dataVersion, setDataVersion] = useState(0)
  return (
    <main className="app">
      <header className="app-header">
        <p className="eyebrow">失敗気づき · Mistake → Insight → Learning → Improvement</p>
        <h1>Shippai Kizuki</h1>
        <p>A closer look at your Japanese learning, one review at a time.</p>
      </header>
      <AnkiConnection onImported={() => setDataVersion((version) => version + 1)} />
      <KanjiPage key={`kanji-${dataVersion}`} />
      <VocabularyPage key={`vocabulary-${dataVersion}`} />
      <details className="review-history">
        <summary>View individual review history</summary>
        <ReviewsPage key={`reviews-${dataVersion}`} />
      </details>
    </main>
  )
}

export default App
