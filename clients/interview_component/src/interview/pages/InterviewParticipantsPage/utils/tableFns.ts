import { format } from 'date-fns'

export interface NumericRangeFilterValue {
  min?: string
  max?: string
  noScore?: boolean
}

export const NOT_SCHEDULED = 'Not scheduled'

const compareNullsLast = <T>(a: T | null, b: T | null, compare: (a: T, b: T) => number): number => {
  if (a === null) return b === null ? 0 : 1
  if (b === null) return -1
  return compare(a, b)
}

export const compareNullableNumbers = (a: number | null, b: number | null): number =>
  compareNullsLast(a, b, (x, y) => x - y)

export const compareNullableStrings = (a: string | null, b: string | null): number =>
  compareNullsLast(a, b, (x, y) => x.localeCompare(y))

export const matchesNumericRange = (
  value: number | null,
  filterValue: NumericRangeFilterValue | undefined,
): boolean => {
  if (filterValue?.noScore) return value === null
  if (value === null) return false
  const min = filterValue?.min ? Number(filterValue.min) : undefined
  const max = filterValue?.max ? Number(filterValue.max) : undefined
  if (min !== undefined && (Number.isNaN(min) || value < min)) return false
  if (max !== undefined && (Number.isNaN(max) || value > max)) return false
  return true
}

export const matchesSelect = (label: string, filterValue: unknown): boolean =>
  !Array.isArray(filterValue) || filterValue.length === 0 || filterValue.includes(label)

export const getInterviewDayLabel = (startTime: number | null): string =>
  startTime === null ? NOT_SCHEDULED : format(startTime, 'PPP')

export const getInterviewDayOptions = (startTimes: (number | null)[]): string[] => {
  const scheduled = startTimes.filter((time): time is number => time !== null).sort((a, b) => a - b)
  const days = [...new Set(scheduled.map(getInterviewDayLabel))]
  return scheduled.length < startTimes.length ? [...days, NOT_SCHEDULED] : days
}

export const isUrl = (value: string): boolean => /^https?:\/\//.test(value)
