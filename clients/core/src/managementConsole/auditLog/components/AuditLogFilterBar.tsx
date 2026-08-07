import {
  DatePickerWithRange,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { DateRange } from 'react-day-picker'
import type { AuditLogFilters } from '../interfaces/auditLog'

interface AuditLogFilterBarProps {
  filters: AuditLogFilters
  onFiltersChange: (filters: AuditLogFilters) => void
}

const ALL = 'all'

const OUTCOMES = [
  { value: 'success', label: 'Success' },
  { value: 'denied', label: 'Denied' },
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

  const dateRange: DateRange | undefined = filters.from
    ? { from: new Date(filters.from), to: filters.to ? new Date(filters.to) : undefined }
    : undefined

  const onDateChange = (range: DateRange | undefined) =>
    update({
      from: range?.from ? range.from.toISOString() : undefined,
      to: range?.to ? range.to.toISOString() : undefined,
    })

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <Input
        placeholder='Search actor, action, entity…'
        value={filters.search ?? ''}
        onChange={(e) => update({ search: e.target.value || undefined })}
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
