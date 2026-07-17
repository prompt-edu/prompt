import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, TEAM_ALLOCATION_API } from '../../src/env'
import { FULL_COURSE_PHASES, TEAM_ALLOCATION_FOREIGN_PHASE_ID } from '../../src/data/constants'

// The team allocation server is reached on the browser origin through the e2e
// nginx proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401
// both for missing tokens and for valid tokens lacking the required role. All
// checks are side-effect-free (reads, plus a batch that auth rejects before it
// runs) on the graph Team Allocation phase, so this file cannot interfere with
// the journey specs (which own their own standalone phases).
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${TEAM_ALLOCATION_API}/course_phase/${phaseId}/${path}`

const PHASE_ID = FULL_COURSE_PHASES.teamAllocation.id

test.describe('team allocation API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(PHASE_ID, 'team'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student on a lecturer-only endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'skill'))
    expect(res.status()).toBe(401)
  })

  test('rejects a student creating a team', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.post(phaseUrl(PHASE_ID, 'team'), { data: { teamNames: ['nope'] } })
    expect(res.status()).toBe(401)
  })

  test('accepts an enrolled student on the team list', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'team'))
    expect(res.status()).toBe(200)
  })

  test('accepts an enrolled student on the allocation list', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'allocation'))
    expect(res.status()).toBe(200)
  })

  test('rejects a student on a phase of a course they are not enrolled in', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(TEAM_ALLOCATION_FOREIGN_PHASE_ID, 'allocation'))
    expect(res.status()).toBe(401)
  })
})
