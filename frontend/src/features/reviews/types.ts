export type ReviewRating = 'again' | 'hard' | 'good' | 'easy'

export type Review = {
  vocabulary: string
  rating: ReviewRating
}
