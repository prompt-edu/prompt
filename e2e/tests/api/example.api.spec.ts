import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, EXAMPLE_API } from '../../src/env'
import { EXAMPLE_FOREIGN_PHASE_ID, EXAMPLE_PHASE_ID } from '../../src/data/constants'

// The example server is reached on the browser origin through the e2e nginx
// proxy (same path prefix as prod Traefik). Its /info read requires admin or a
// course-scoped lecturer; prompt-sdk auth answers 401 both for missing tokens
// and for valid tokens lacking the required role.
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${EXAMPLE_API}/course_phase/${phaseId}/${path}`

test.describe('example API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(EXAMPLE_PHASE_ID, 'info'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student on the lecturer-only info endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(EXAMPLE_PHASE_ID, 'info'))
    expect(res.status()).toBe(401)
  })

  test('accepts a course lecturer', async ({ apiAs }) => {
    const api = await apiAs('course-lecturer')
    const res = await api.get(phaseUrl(EXAMPLE_PHASE_ID, 'info'))
    expect(res.status()).toBe(200)
  })

  test('rejects a course lecturer on a phase of a course they do not lead', async ({
    apiAs,
  }) => {
    const api = await apiAs('course-lecturer')
    const res = await api.get(phaseUrl(EXAMPLE_FOREIGN_PHASE_ID, 'info'))
    expect(res.status()).toBe(401)
  })
})
