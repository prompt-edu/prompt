import type {
  AuditCursor,
  AuditLogFilters,
  AuditLogPage,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const buildAuditLogQuery = (
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
): string => {
  const params = new URLSearchParams()
  const set = (key: string, value?: string) => {
    if (value) params.set(key, value)
  }
  set('actorRole', filters.actorRole)
  set('outcome', filters.outcome)
  set('sourceService', filters.sourceService)
  set('entityType', filters.entityType)
  set('actionKey', filters.actionKey)
  set('coursePhaseID', filters.coursePhaseID)
  set('search', filters.search)
  set('from', filters.from)
  set('to', filters.to)
  // Sent explicitly so the page size the table renders and the page size the
  // server returns cannot drift apart.
  params.set('limit', String(limit))
  if (cursor) {
    params.set('cursorTs', cursor.createdAt)
    params.set('cursorId', cursor.id)
  }
  return params.toString()
}

export const getCourseAuditLog = async (
  courseId: string,
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
): Promise<AuditLogPage> => {
  const query = buildAuditLogQuery(filters, limit, cursor)
  return (await axiosInstance.get(`/api/courses/${courseId}/audit-log?${query}`)).data
}

export const getGlobalAuditLog = async (
  filters: AuditLogFilters,
  limit: number,
  cursor?: AuditCursor,
): Promise<AuditLogPage> => {
  const query = buildAuditLogQuery(filters, limit, cursor)
  return (await axiosInstance.get(`/api/audit-log?${query}`)).data
}
