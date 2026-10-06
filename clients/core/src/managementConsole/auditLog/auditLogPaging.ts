import type { AuditCursor } from './interfaces/auditLog'

export const AUDIT_PAGE_SIZE = 50

// The log is keyset-paginated: a page only knows the cursor of the page after
// it, so walking back needs the cursors already visited. The stack holds the
// cursor of every page beyond the newest one; an empty stack is the newest page.

export const currentCursor = (stack: AuditCursor[]): AuditCursor | undefined =>
  stack[stack.length - 1]

export const hasNewerPage = (stack: AuditCursor[]): boolean => stack.length > 0

export const pushOlderPage = (stack: AuditCursor[], next: AuditCursor | null): AuditCursor[] =>
  next ? [...stack, next] : stack

export const popNewerPage = (stack: AuditCursor[]): AuditCursor[] => stack.slice(0, -1)
