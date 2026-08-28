import type {
  AuditCursor,
  AuditLogFilters,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getCourseAuditLog, getGlobalAuditLog } from '../queries/getAuditLog'

export const useCourseAuditLog = (
  courseId: string | undefined,
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
) => {
  return useQuery({
    queryKey: ['auditLog', courseId, filters, limit, cursor],
    queryFn: () => getCourseAuditLog(courseId!, filters, limit, cursor),
    enabled: !!courseId,
    // Keep the current page on screen while the next one loads, so paging does
    // not unmount the table.
    placeholderData: keepPreviousData,
  })
}

export const useGlobalAuditLog = (
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
) => {
  return useQuery({
    queryKey: ['auditLog', 'global', filters, limit, cursor],
    queryFn: () => getGlobalAuditLog(filters, limit, cursor),
    placeholderData: keepPreviousData,
  })
}
