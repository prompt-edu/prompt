import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useCourseAuditLog } from '@core/network/hooks/useAuditLog'
import { AuditLogView } from '@managementConsole/auditLog/AuditLogView'
import type { AuditLogFilters } from '@managementConsole/auditLog/interfaces/auditLog'

export const CourseAuditLogPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const [filters, setFilters] = useState<AuditLogFilters>({})
  const query = useCourseAuditLog(courseId, filters)

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
