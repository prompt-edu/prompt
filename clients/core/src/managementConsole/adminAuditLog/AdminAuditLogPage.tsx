import { useState } from 'react'
import { useGlobalAuditLog } from '@core/network/hooks/useAuditLog'
import { AuditLogView } from '@managementConsole/auditLog/AuditLogView'
import type { AuditLogFilters } from '@managementConsole/auditLog/interfaces/auditLog'

export const AdminAuditLogPage = () => {
  const [filters, setFilters] = useState<AuditLogFilters>({})
  const query = useGlobalAuditLog(filters)

  return (
    <AuditLogView
      title='Audit Log'
      pages={query.data?.pages}
      isLoading={query.isLoading}
      isError={query.isError}
      hasNextPage={!!query.hasNextPage}
      isFetchingNextPage={query.isFetchingNextPage}
      onLoadMore={() => query.fetchNextPage()}
      filters={filters}
      onFiltersChange={setFilters}
    />
  )
}
