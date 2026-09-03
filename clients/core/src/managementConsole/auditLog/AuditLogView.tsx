import {
  Button,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
  PromptTable,
} from '@tumaet/prompt-ui-components'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useMemo } from 'react'
import { AUDIT_PAGE_SIZE } from './auditLogPaging'
import { AuditLogFilterBar } from './components/AuditLogFilterBar'
import { getAuditLogColumns } from './components/auditLogColumns'
import type { AuditLogFilters, AuditLogPage } from './interfaces/auditLog'

interface AuditLogViewProps {
  title: string
  page?: AuditLogPage
  isDisabled: boolean
  isLoading: boolean
  isError: boolean
  isFetching: boolean
  hasNewer: boolean
  hasOlder: boolean
  onNewer: () => void
  onOlder: () => void
  filters: AuditLogFilters
  onFiltersChange: (filters: AuditLogFilters) => void
}

export const AuditLogView = ({
  title,
  page,
  isDisabled,
  isLoading,
  isError,
  isFetching,
  hasNewer,
  hasOlder,
  onNewer,
  onOlder,
  filters,
  onFiltersChange,
}: AuditLogViewProps) => {
  const entries = page?.entries ?? []
  const columns = useMemo(() => getAuditLogColumns(), [])

  if (isDisabled) {
    return (
      <div className='space-y-6'>
        <ManagementPageHeader>{title}</ManagementPageHeader>
        <p className='text-muted-foreground'>Audit logging is turned off for this deployment.</p>
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>{title}</ManagementPageHeader>

      <AuditLogFilterBar filters={filters} onFiltersChange={onFiltersChange} />

      {isError ? (
        <ErrorPage message='Failed to load the audit log.' />
      ) : isLoading ? (
        <LoadingPage />
      ) : (
        <>
          <PromptTable data={entries} columns={columns} pageSize={AUDIT_PAGE_SIZE} />
          {(hasNewer || hasOlder) && (
            <div className='flex items-center justify-center gap-2'>
              <Button variant='outline' onClick={onNewer} disabled={!hasNewer || isFetching}>
                <ChevronLeft className='mr-1 h-4 w-4' />
                Newer entries
              </Button>
              <Button variant='outline' onClick={onOlder} disabled={!hasOlder || isFetching}>
                Older entries
                <ChevronRight className='ml-1 h-4 w-4' />
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
