import type { AuditLogFilters } from '@core/managementConsole/auditLog/interfaces/auditLog'
import { describe, expect, it } from 'vitest'
import { buildAuditLogQuery } from './auditLog'

describe('buildAuditLogQuery', () => {
  it('sends only the filters that are set, plus the page size', () => {
    const params = new URLSearchParams(buildAuditLogQuery({ outcome: 'denied' }, 50))
    expect(params.get('outcome')).toBe('denied')
    expect(params.get('limit')).toBe('50')
    expect(params.has('search')).toBe(false)
    expect(params.has('actorRole')).toBe(false)
    expect(params.has('cursorTs')).toBe(false)
  })

  it('drops empty filter values rather than sending them', () => {
    const filters: AuditLogFilters = { search: '', sourceService: undefined, actorRole: 'Lecturer' }
    const params = new URLSearchParams(buildAuditLogQuery(filters, 20))
    expect(params.has('search')).toBe(false)
    expect(params.has('sourceService')).toBe(false)
    expect(params.get('actorRole')).toBe('Lecturer')
  })

  it('sends both cursor halves together, since the server rejects one alone', () => {
    const params = new URLSearchParams(
      buildAuditLogQuery({}, 50, {
        createdAt: '2026-08-28T12:23:28.027804Z',
        id: '07f99ae1-ff39-47bb-8e6d-3fdff9eb4933',
      }),
    )
    expect(params.get('cursorTs')).toBe('2026-08-28T12:23:28.027804Z')
    expect(params.get('cursorId')).toBe('07f99ae1-ff39-47bb-8e6d-3fdff9eb4933')
  })
})
