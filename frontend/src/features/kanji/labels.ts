import type { KanjiRecommendation } from './types'

export const recommendationLabels: Record<KanjiRecommendation, string> = {
  limited_review_evidence: 'Limited review evidence',
  limited_cross_word_evidence: 'Limited cross-word evidence',
  suggested_practice: 'Suggested practice',
  no_extra_practice: 'No extra practice suggested',
}
