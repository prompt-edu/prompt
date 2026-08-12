import type {
  AuditCursor,
  AuditLogFilters,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { useInfiniteQuery } from '@tanstack/react-query'
import { getCourseAuditLog, getGlobalAuditLog } from '../queries/getAuditLog'

export const useCourseAuditLog = (courseId: string | undefined, filters: AuditLogFilters) => {
  return useInfiniteQuery({
    queryKey: ['auditLog', courseId, filters],
    queryFn: ({ pageParam }) => getCourseAuditLog(courseId!, filters, pageParam),
    initialPageParam: undefined as AuditCursor | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    enabled: !!courseId,
  })
}

export const useGlobalAuditLog = (filters: AuditLogFilters) => {
  return useInfiniteQuery({
    queryKey: ['auditLog', 'global', filters],
    queryFn: ({ pageParam }) => getGlobalAuditLog(filters, pageParam),
    initialPageParam: undefined as AuditCursor | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
  })
}
