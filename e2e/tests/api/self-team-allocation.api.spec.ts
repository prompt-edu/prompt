import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, SELF_TEAM_ALLOCATION_API } from '../../src/env'
import {
  SELF_TEAM_ALLOCATION_FOREIGN_PHASE_ID,
  SELF_TEAM_ALLOCATION_PHASE_ID,
} from '../../src/data/constants'

// The phase server is reached on the browser origin through the e2e nginx
// proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401 both
// for missing tokens and for valid tokens lacking the required role.
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${SELF_TEAM_ALLOCATION_API}/course_phase/${phaseId}/${path}`

test.describe('self team allocation API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(SELF_TEAM_ALLOCATION_PHASE_ID, 'team'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a token without a course-scoped role', async ({ apiAs }) => {
    // `lecturer` has PROMPT_Lecturer only; /team requires admin, course
    // lecturer, or course student.
    const api = await apiAs('lecturer')
    const res = await api.get(phaseUrl(SELF_TEAM_ALLOCATION_PHASE_ID, 'team'))
    expect(res.status()).toBe(401)
  })

  test('rejects a student on a lecturer-only endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(SELF_TEAM_ALLOCATION_PHASE_ID, 'team/tutors'))
    expect(res.status()).toBe(401)
  })

  test('accepts an enrolled student (is_student round-trip via core)', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(SELF_TEAM_ALLOCATION_PHASE_ID, 'team'))
    expect(res.status()).toBe(200)
  })

  test('rejects a student on a phase of a course they are not enrolled in', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(SELF_TEAM_ALLOCATION_FOREIGN_PHASE_ID, 'team'))
    expect(res.status()).toBe(401)
  })
})
