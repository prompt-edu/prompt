import type {
  AuditCursor,
  AuditLogFilters,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { coreKeys } from '@core/network/cache'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getCourseAuditLog, getGlobalAuditLog } from '../queries/getAuditLog'

export const useCourseAuditLog = (
  courseId: string | undefined,
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
  enabled = true,
) => {
  return useQuery({
    queryKey: coreKeys.auditLog.inCourse(courseId, filters, limit, cursor),
    queryFn: () => getCourseAuditLog(courseId!, filters, limit, cursor),
    enabled: enabled && !!courseId,
    // Keep the current page on screen while the next one loads, so paging does
    // not unmount the table.
    placeholderData: keepPreviousData,
  })
}

export const useGlobalAuditLog = (
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
  enabled = true,
) => {
  return useQuery({
    queryKey: coreKeys.auditLog.global(filters, limit, cursor),
    queryFn: () => getGlobalAuditLog(filters, limit, cursor),
    enabled,
    placeholderData: keepPreviousData,
  })
}
