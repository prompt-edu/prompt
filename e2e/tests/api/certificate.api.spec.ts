import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, CERTIFICATE_API } from '../../src/env'
import { CERTIFICATE_FOREIGN_PHASE_ID, CERTIFICATE_PHASES } from '../../src/data/constants'

// The certificate server is reached on the browser origin through the e2e
// nginx proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401
// both for missing tokens and for valid tokens lacking the required role.
// All checks are side-effect-free (GET only) on the graph-tail certificate
// phase, which the journeys never touch.
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${CERTIFICATE_API}/course_phase/${phaseId}/${path}`

const PHASE_ID = CERTIFICATE_PHASES.graphTail

test.describe('certificate API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(PHASE_ID, 'config'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student on the template download endpoint', async ({ apiAs }) => {
    // config/template is staff-only; students hold the Student role only.
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'config/template'))
    expect(res.status()).toBe(401)
  })

  test('rejects a student on the participants endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'participants'))
    expect(res.status()).toBe(401)
  })

  test('rejects a student on the preview endpoint', async ({ apiAs }) => {
    // Preview is limited to admins and lecturers.
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'certificate/preview'))
    expect(res.status()).toBe(401)
  })

  test('accepts a course editor on the participants endpoint', async ({ apiAs }) => {
    const api = await apiAs('course-editor')
    const res = await api.get(phaseUrl(PHASE_ID, 'participants'))
    expect(res.status()).toBe(200)
  })

  test('accepts an enrolled student on the status endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'certificate/status'))
    expect(res.status()).toBe(200)
  })

  test('rejects a student on a phase of a course they are not enrolled in', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(CERTIFICATE_FOREIGN_PHASE_ID, 'certificate/status'))
    expect(res.status()).toBe(401)
  })
})
