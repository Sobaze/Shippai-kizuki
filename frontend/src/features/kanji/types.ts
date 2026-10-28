export type KanjiRecommendation =
  | 'limited_review_evidence'
  | 'limited_cross_word_evidence'
  | 'suggested_practice'
  | 'no_extra_practice'

export type KanjiEvidence = {
  vocabulary: string
  reading?: string
  difficulty: number
  reviewCount: number
  qualifies: boolean
}

export type KanjiStats = {
  kanji: string
  difficulty: number | null
  wordCount: number
  qualifiedWordCount: number
  recommendation: KanjiRecommendation
  evidence: KanjiEvidence[]
}

export type KanjiDetails = {
  kanji: string
  meanings: string[]
  onReadings: string[]
  kunReadings: string[]
  nameReadings: string[]
  strokeCount: number
}
