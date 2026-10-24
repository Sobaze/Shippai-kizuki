export type PracticeRecommendation =
  | 'limited_evidence'
  | 'suggested_practice'
  | 'no_extra_practice'

export type VocabularyStats = {
  vocabulary: string
  reviewCount: number
  ratingCounts: {
    again: number
    hard: number
    good: number
    easy: number
  }
  difficulty: number
  recommendation: PracticeRecommendation
}
