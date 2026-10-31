export const practiceFilterOptions = [
  { value: 'all', label: 'All' },
  { value: 'practice', label: 'Suggested practice' },
  { value: 'limited', label: 'Limited evidence' },
  { value: 'no-extra', label: 'No extra practice suggested' },
] as const

export type PracticeFilter = typeof practiceFilterOptions[number]['value']

export function parsePracticeFilter(value: string | null): PracticeFilter {
  return practiceFilterOptions.find((option) => option.value === value)?.value ?? 'all'
}

export function matchesPracticeFilter(recommendation: string, filter: PracticeFilter): boolean {
  switch (filter) {
    case 'all': return true
    case 'practice': return recommendation === 'suggested_practice'
    case 'limited': return recommendation.startsWith('limited_')
    case 'no-extra': return recommendation === 'no_extra_practice'
  }
}
