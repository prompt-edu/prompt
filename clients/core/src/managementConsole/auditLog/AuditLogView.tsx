import {
  Button,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
  PromptTable,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { AuditLogFilterBar } from './components/AuditLogFilterBar'
import { getAuditLogColumns } from './components/auditLogColumns'
import type { AuditEntry, AuditLogFilters, AuditLogPage } from './interfaces/auditLog'

interface AuditLogViewProps {
  title: string
  pages?: AuditLogPage[]
  isLoading: boolean
  isError: boolean
  hasNextPage: boolean
  isFetchingNextPage: boolean
  onLoadMore: () => void
  filters: AuditLogFilters
  onFiltersChange: (filters: AuditLogFilters) => void
}

export const AuditLogView = ({
  title,
  pages,
  isLoading,
  isError,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
  filters,
  onFiltersChange,
}: AuditLogViewProps) => {
  const entries: AuditEntry[] = useMemo(() => pages?.flatMap((page) => page.entries) ?? [], [pages])
  const columns = useMemo(() => getAuditLogColumns(), [])

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
          {/*
            The audit log is paged server-side via "Load more"; keep every
            loaded row on a single PromptTable page (pageSize = row count) so its
            client-side pager does not split the accumulated rows and make the
            button appear to do nothing.
          */}
          <PromptTable data={entries} columns={columns} pageSize={Math.max(entries.length, 1)} />
          {hasNextPage && (
            <div className='flex justify-center'>
              <Button variant='outline' onClick={onLoadMore} disabled={isFetchingNextPage}>
                {isFetchingNextPage ? 'Loading…' : 'Load more'}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
