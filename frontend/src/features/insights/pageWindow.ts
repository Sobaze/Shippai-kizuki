export type PageWindow = {
  page: number
  totalPages: number
  totalItems: number
  startIndex: number
  endIndex: number
  from: number
  to: number
}

export function getPageWindow(rawPage: string | null, totalItems: number, pageSize: number): PageWindow {
  const parsed = rawPage && /^[1-9]\d*$/.test(rawPage) ? Number(rawPage) : 1
  const requested = Number.isSafeInteger(parsed) ? parsed : 1
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const page = Math.min(requested, totalPages)
  const startIndex = (page - 1) * pageSize
  const endIndex = Math.min(startIndex + pageSize, totalItems)
  return { page, totalPages, totalItems, startIndex, endIndex, from: totalItems === 0 ? 0 : startIndex + 1, to: endIndex }
}

export function getPageNumbers(page: number, totalPages: number): Array<number | null> {
  if (totalPages <= 7) return Array.from({ length: totalPages }, (_, index) => index + 1)

  const numbers = new Set([1, totalPages])
  const start = Math.max(2, Math.min(page - 1, totalPages - 3))
  const end = Math.min(totalPages - 1, Math.max(page + 1, 4))
  for (let value = start; value <= end; value++) numbers.add(value)

  const result: Array<number | null> = []
  let previous = 0
  for (const value of [...numbers].sort((a, b) => a - b)) {
    if (previous && value - previous === 2) result.push(previous + 1)
    else if (previous && value - previous > 2) result.push(null)
    result.push(value)
    previous = value
  }
  return result
}
