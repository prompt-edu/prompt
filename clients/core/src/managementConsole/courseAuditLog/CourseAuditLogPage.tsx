import { useCourseAuditLog } from '@core/network/hooks/useAuditLog'
import { AuditLogView } from '@managementConsole/auditLog/AuditLogView'
import { AUDIT_PAGE_SIZE } from '@managementConsole/auditLog/auditLogPaging'
import { useAuditLogBrowser } from '@managementConsole/auditLog/useAuditLogBrowser'
import { useParams } from 'react-router-dom'

export const CourseAuditLogPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { filters, cursor, onFiltersChange, navigation } = useAuditLogBrowser()
  const query = useCourseAuditLog(courseId, filters, AUDIT_PAGE_SIZE, cursor)

  return (
    <AuditLogView
      title='Audit Log'
      page={query.data}
      isLoading={query.isLoading}
      isError={query.isError}
      isFetching={query.isFetching}
      filters={filters}
      onFiltersChange={onFiltersChange}
      {...navigation(query.data?.nextCursor)}
    />
  )
}
