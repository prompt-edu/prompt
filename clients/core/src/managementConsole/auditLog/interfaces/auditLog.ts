export interface AuditEntry {
  id: string
  createdAt: string
  actorID?: string
  actorName: string
  actorEmail: string
  actorRoles: string[]
  actorRole: string
  action: string
  actionKey: string
  outcome: string
  entityType?: string
  entityID?: string
  entityName?: string
  courseID?: string
  coursePhaseID?: string
  sourceService: string
  httpMethod?: string
  httpPath?: string
  httpStatus?: number
  metadata?: Record<string, unknown>
}

export interface AuditCursor {
  createdAt: string
  id: string
}

export interface AuditLogPage {
  entries: AuditEntry[]
  nextCursor: AuditCursor | null
}

export interface AuditLogFilters {
  actorRole?: string
  outcome?: string
  sourceService?: string
  entityType?: string
  actionKey?: string
  coursePhaseID?: string
  search?: string
  from?: string
  to?: string
}
