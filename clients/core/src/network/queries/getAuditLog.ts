import type {
  AuditCursor,
  AuditLogFilters,
  AuditLogPage,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { axiosInstance } from '@tumaet/prompt-shared-state'

const buildQuery = (filters: AuditLogFilters, cursor?: AuditCursor): string => {
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
  if (cursor) {
    params.set('cursorTs', cursor.createdAt)
    params.set('cursorId', cursor.id)
  }
  return params.toString()
}

export const getCourseAuditLog = async (
  courseId: string,
  filters: AuditLogFilters,
  cursor?: AuditCursor,
): Promise<AuditLogPage> => {
  const query = buildQuery(filters, cursor)
  return (await axiosInstance.get(`/api/courses/${courseId}/audit-log?${query}`)).data
}

export const getGlobalAuditLog = async (
  filters: AuditLogFilters,
  cursor?: AuditCursor,
): Promise<AuditLogPage> => {
  const query = buildQuery(filters, cursor)
  return (await axiosInstance.get(`/api/audit-log?${query}`)).data
}
