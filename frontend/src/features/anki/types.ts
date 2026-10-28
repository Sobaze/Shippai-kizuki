export type AnkiDeckList = {
  decks: string[]
}

export type VocabularyImportSummary = {
  deck: string
  cardCount: number
  reviewCount: number
  unscoredReviewCount: number
  missingReadingCount: number
  importedAt: string | null
}
