import { matchesPracticeFilter, practiceFilterOptions } from './practiceFilter'
import type { PracticeFilter } from './practiceFilter'

type PracticeFiltersProps = {
  label: string
  value: PracticeFilter
  recommendations: string[]
  onChange: (filter: PracticeFilter) => void
}

export function PracticeFilters({ label, value, recommendations, onChange }: PracticeFiltersProps) {
  return (
    <div className="practice-filters" role="group" aria-label={label}>
      {practiceFilterOptions.map((option) => (
        <button
          key={option.value}
          type="button"
          aria-pressed={value === option.value}
          onClick={() => onChange(option.value)}
        >
          {option.label}
          <span>{recommendations.filter((item) => matchesPracticeFilter(item, option.value)).length.toLocaleString()}</span>
        </button>
      ))}
    </div>
  )
}
