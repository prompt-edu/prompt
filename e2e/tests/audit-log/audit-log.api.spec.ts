import { test, expect } from '../../src/fixtures/api'
import { SEEDED_COURSES } from '../../src/data/constants'

const COURSE_ID = SEEDED_COURSES.iPraktikum.id

test.describe('core API: audit log', () => {
  test('an unparseable coursePhaseID is rejected instead of ignored', async ({ apiAs }) => {
    const api = await apiAs('admin')

    // Silently dropping the filter would answer a typo with the full unfiltered
    // log and a 200.
    const bad = await api.get('/api/audit-log?coursePhaseID=not-a-uuid')
    expect(bad.status()).toBe(400)

    const scoped = await api.get(
      `/api/courses/${COURSE_ID}/audit-log?coursePhaseID=not-a-uuid`,
    )
    expect(scoped.status()).toBe(400)
  })

  test('an oversized limit is clamped to the documented maximum', async ({ apiAs }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/audit-log?limit=1000')
    expect(res.ok()).toBeTruthy()

    const page = (await res.json()) as { entries: unknown[] }
    expect(page.entries.length).toBeLessThanOrEqual(200)
  })

  test('the ingest endpoint refuses unauthenticated events', async ({ apiAs }) => {
    const api = await apiAs('admin')
    // A user token is not an ingest credential: only the per-service shared
    // secret is.
    const res = await api.post('/api/audit', {
      data: { action: 'Forged entry', outcome: 'success' },
    })
    expect(res.status()).toBe(401)
  })
})
