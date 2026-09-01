import type {
  AuditCursor,
  AuditLogFilters,
  AuditLogPage,
  AuditLogStatus,
} from '@core/managementConsole/auditLog/interfaces/auditLog'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/audit-log`

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

export const auditLog = {
  inCourse: (
    courseID: string,
    filters: AuditLogFilters,
    limit: number,
    cursor?: AuditCursor,
  ): Promise<AuditLogPage> =>
    coreRequest.get(
      `${API_PREFIX}/courses/${courseID}/audit-log?${buildAuditLogQuery(filters, limit, cursor)}`,
    ),

  global: (filters: AuditLogFilters, limit: number, cursor?: AuditCursor): Promise<AuditLogPage> =>
    coreRequest.get(`${path}?${buildAuditLogQuery(filters, limit, cursor)}`),

  status: (): Promise<AuditLogStatus> => coreRequest.get(`${path}/status`),
}
