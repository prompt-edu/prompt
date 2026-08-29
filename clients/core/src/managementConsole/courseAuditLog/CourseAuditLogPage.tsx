import { useCourseAuditLog } from '@core/network/hooks/useAuditLog'
import { useAuditLogStatus } from '@core/network/hooks/useAuditLogStatus'
import { AuditLogView } from '@managementConsole/auditLog/AuditLogView'
import { AUDIT_PAGE_SIZE } from '@managementConsole/auditLog/auditLogPaging'
import { useAuditLogBrowser } from '@managementConsole/auditLog/useAuditLogBrowser'
import { useParams } from 'react-router-dom'

export const CourseAuditLogPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { data: status, isPending: isStatusPending } = useAuditLogStatus()
  const auditLogEnabled = status?.enabled ?? false
  const { filters, cursor, onFiltersChange, navigation } = useAuditLogBrowser()
  const query = useCourseAuditLog(courseId, filters, AUDIT_PAGE_SIZE, cursor, auditLogEnabled)

  return (
    <AuditLogView
      title='Audit Log'
      page={query.data}
      isDisabled={!isStatusPending && !auditLogEnabled}
      isLoading={isStatusPending || query.isLoading}
      isError={query.isError}
      isFetching={query.isFetching}
      filters={filters}
      onFiltersChange={onFiltersChange}
      {...navigation(query.data?.nextCursor)}
    />
  )
}
