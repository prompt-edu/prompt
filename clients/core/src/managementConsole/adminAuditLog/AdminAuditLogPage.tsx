import { useGlobalAuditLog } from '@core/network/hooks/useAuditLog'
import { useAuditLogStatus } from '@core/network/hooks/useAuditLogStatus'
import { AuditLogView } from '@managementConsole/auditLog/AuditLogView'
import { AUDIT_PAGE_SIZE } from '@managementConsole/auditLog/auditLogPaging'
import { useAuditLogBrowser } from '@managementConsole/auditLog/useAuditLogBrowser'

export const AdminAuditLogPage = () => {
  const { data: status, isPending: isStatusPending, isError: isStatusError } = useAuditLogStatus()
  const auditLogEnabled = status?.enabled ?? false
  const { filters, cursor, onFiltersChange, navigation } = useAuditLogBrowser()
  const query = useGlobalAuditLog(filters, AUDIT_PAGE_SIZE, cursor, auditLogEnabled)

  return (
    <AuditLogView
      title='Audit Log'
      page={query.data}
      isDisabled={status?.enabled === false}
      isLoading={isStatusPending || query.isLoading}
      isError={isStatusError || query.isError}
      isFetching={query.isFetching}
      filters={filters}
      onFiltersChange={onFiltersChange}
      {...navigation(query.data?.nextCursor)}
    />
  )
}
