export const SORT_OPTIONS = [
  { id: 'firstName', label: 'First Name' },
  { id: 'lastName', label: 'Last Name' },
  { id: 'acceptanceStatus', label: 'Acceptance Status' },
  { id: 'interviewDate', label: 'Interview Date' },
  { id: 'interviewScore', label: 'Interview Score' },
] as const

export type SortOption = (typeof SORT_OPTIONS)[number]['id']

export const DEFAULT_SORT: SortOption = 'interviewDate'

// Same query param and `<id>:<asc|desc>` format as PromptTableURL in @tumaet/prompt-ui-components.
export const SORTING_QUERY_PARAM = 'sorting'

const isSortOption = (value: string): value is SortOption =>
  SORT_OPTIONS.some((option) => option.id === value)

/**
 * Reads the sort option from a `sorting` query param value. The overview only sorts ascending,
 * so anything that is not a known ascending option falls back to the default.
 */
export const parseSortParam = (value: string | null): SortOption => {
  const [id, order = 'asc'] = (value ?? '').split(',')[0].split(':')
  return order === 'asc' && isSortOption(id) ? id : DEFAULT_SORT
}

/** Serializes the sort option for the `sorting` query param, or `null` for the default. */
export const serializeSortParam = (sortBy: SortOption): string | null =>
  sortBy === DEFAULT_SORT ? null : `${sortBy}:asc`
