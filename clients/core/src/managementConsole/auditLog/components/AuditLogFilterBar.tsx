import {
  DatePickerWithRange,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { endOfDay, startOfDay } from 'date-fns'
import { useEffect, useRef, useState } from 'react'
import type { DateRange } from 'react-day-picker'
import type { AuditLogFilters } from '../interfaces/auditLog'

interface AuditLogFilterBarProps {
  filters: AuditLogFilters
  onFiltersChange: (filters: AuditLogFilters) => void
}

const ALL = 'all'

const useDebouncedValue = <T,>(value: T, delayMs = 400): T => {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(id)
  }, [value, delayMs])
  return debounced
}

const OUTCOMES = [
  { value: 'success', label: 'Success' },
  { value: 'denied', label: 'Denied' },
  { value: 'error', label: 'Error' },
]

const ROLES = [
  { value: 'PROMPT_Admin', label: 'Prompt Admin' },
  { value: 'PROMPT_Lecturer', label: 'Prompt Lecturer' },
  { value: 'Lecturer', label: 'Course Lecturer' },
  { value: 'Editor', label: 'Course Editor' },
  { value: 'Student', label: 'Student' },
]

export const AuditLogFilterBar = ({ filters, onFiltersChange }: AuditLogFilterBarProps) => {
  const update = (partial: Partial<AuditLogFilters>) => onFiltersChange({ ...filters, ...partial })

  // Debounce the free-text search so a burst of keystrokes does not fire a
  // request (and a full-table ILIKE scan) per character. The debounced value is
  // pushed into the filters, which are part of the query key.
  const [searchInput, setSearchInput] = useState(filters.search ?? '')
  const debouncedSearch = useDebouncedValue(searchInput, 400)
  const filtersRef = useRef(filters)
  filtersRef.current = filters
  useEffect(() => {
    const next = debouncedSearch || undefined
    if (next !== filtersRef.current.search) {
      onFiltersChange({ ...filtersRef.current, search: next })
    }
  }, [debouncedSearch, onFiltersChange])

  const dateRange: DateRange | undefined = filters.from
    ? { from: new Date(filters.from), to: filters.to ? new Date(filters.to) : undefined }
    : undefined

  // Send whole-day bounds: startOfDay(from)..endOfDay(to). Sending the raw
  // midnight values would exclude the selected end day and pull in the evening
  // before the start day once converted to UTC.
  const onDateChange = (range: DateRange | undefined) =>
    update({
      from: range?.from ? startOfDay(range.from).toISOString() : undefined,
      to: range?.to ? endOfDay(range.to).toISOString() : undefined,
    })

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <Input
        placeholder='Search actor, action, entity…'
        value={searchInput}
        onChange={(e) => setSearchInput(e.target.value)}
        className='w-64'
      />

      <Select
        value={filters.outcome ?? ALL}
        onValueChange={(value) => update({ outcome: value === ALL ? undefined : value })}
      >
        <SelectTrigger className='w-40'>
          <SelectValue placeholder='Outcome' />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All outcomes</SelectItem>
          {OUTCOMES.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={filters.actorRole ?? ALL}
        onValueChange={(value) => update({ actorRole: value === ALL ? undefined : value })}
      >
        <SelectTrigger className='w-44'>
          <SelectValue placeholder='Role' />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All roles</SelectItem>
          {ROLES.map((r) => (
            <SelectItem key={r.value} value={r.value}>
              {r.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Input
        placeholder='Source (e.g. core)'
        value={filters.sourceService ?? ''}
        onChange={(e) => update({ sourceService: e.target.value || undefined })}
        className='w-40'
      />

      <DatePickerWithRange date={dateRange} setDate={onDateChange} />
    </div>
  )
}
