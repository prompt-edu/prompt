import { useState } from 'react'
import { currentCursor, hasNewerPage, popNewerPage, pushOlderPage } from './auditLogPaging'
import type { AuditCursor, AuditLogFilters } from './interfaces/auditLog'

// Owns the filter and keyset-cursor state shared by the admin and course audit
// log pages. Changing a filter drops the cursor stack in the same update, so a
// new filter always starts from the newest page.
export const useAuditLogBrowser = () => {
  const [filters, setFilters] = useState<AuditLogFilters>({})
  const [cursorStack, setCursorStack] = useState<AuditCursor[]>([])

  const navigation = (nextCursor?: AuditCursor | null) => ({
    hasNewer: hasNewerPage(cursorStack),
    hasOlder: !!nextCursor,
    onNewer: () => setCursorStack(popNewerPage),
    onOlder: () => setCursorStack((stack) => pushOlderPage(stack, nextCursor ?? null)),
  })

  const onFiltersChange = (next: AuditLogFilters) => {
    setFilters(next)
    setCursorStack([])
  }

  return { filters, cursor: currentCursor(cursorStack), onFiltersChange, navigation }
}
