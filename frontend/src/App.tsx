import { ReviewsPage } from './features/reviews/ReviewsPage'
import { VocabularyPage } from './features/vocabulary/VocabularyPage'
import './App.css'

function App() {
  return (
    <main className="app">
      <header className="app-header">
        <p className="eyebrow">失敗気づき · Mistake → Insight → Learning → Improvement</p>
        <h1>Shippai Kizuki</h1>
        <p>A closer look at your Japanese learning, one review at a time.</p>
      </header>
      <VocabularyPage />
      <details className="review-history">
        <summary>View individual review history</summary>
        <ReviewsPage />
      </details>
    </main>
  )
}

export default App
