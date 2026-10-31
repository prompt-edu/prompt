import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, TEAM_ALLOCATION_API } from '../../src/env'
import {
  FULL_COURSE_PHASES,
  TEAM_ALLOCATION_FOREIGN_PHASE_ID,
  TEAM_ALLOCATION_TUTOR_SCOPE,
} from '../../src/data/constants'

// The team allocation server is reached on the browser origin through the e2e
// nginx proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401
// for missing tokens and 403 for valid tokens lacking the required role. All
// checks are side-effect-free (reads, plus writes that auth rejects or that
// find nothing to change) on the graph and the tutor scope Team Allocation
// phases, so this file cannot interfere with the journey specs (which own their
// own standalone phases).
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
    expect(res.status()).toBe(403)
  })

  test('rejects a student creating a team', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.post(phaseUrl(PHASE_ID, 'team'), { data: { teamNames: ['nope'] } })
    expect(res.status()).toBe(403)
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
    expect(res.status()).toBe(403)
  })
})

// course-editor tutors one team of the tutor scope phase. Allocations are the
// lecturer's decision: the tutor reads their own team but writes nothing.
test.describe('team allocation API tutor access', () => {
  const { phaseId, ownTeamId, ownTeamParticipationIds, unallocatedParticipationId } =
    TEAM_ALLOCATION_TUTOR_SCOPE
  const [ownMember] = ownTeamParticipationIds

  test('lets a tutor read only their own team allocations', async ({ apiAs }) => {
    const api = await apiAs('course-editor')
    const res = await api.get(phaseUrl(phaseId, 'allocation'))
    expect(res.status()).toBe(200)

    const allocations = (await res.json()) as { courseParticipationID: string }[]
    expect(allocations.map((a) => a.courseParticipationID).sort()).toEqual(
      [...ownTeamParticipationIds].sort(),
    )
  })

  test('rejects a tutor changing an allocation of their own team', async ({ apiAs }) => {
    const api = await apiAs('course-editor')

    const assign = await api.put(phaseUrl(phaseId, `allocation/${unallocatedParticipationId}`), {
      data: { teamID: ownTeamId },
    })
    expect(assign.status()).toBe(403)

    const reassign = await api.put(phaseUrl(phaseId, `allocation/${ownMember}`), {
      data: { teamID: ownTeamId },
    })
    expect(reassign.status()).toBe(403)

    const remove = await api.delete(phaseUrl(phaseId, `allocation/${ownMember}`))
    expect(remove.status()).toBe(403)

    const unchanged = await api.get(phaseUrl(phaseId, `allocation/${ownMember}`))
    expect(unchanged.status()).toBe(200)
    expect(await unchanged.json()).toEqual({ teamAllocation: ownTeamId })
  })

  test('admits a lecturer to the allocation write routes', async ({ apiAs }) => {
    const api = await apiAs('course-lecturer')
    // The participant has no allocation, so the 404 proves the request passed
    // auth and reached the handler without changing the fixture.
    const res = await api.delete(phaseUrl(phaseId, `allocation/${unallocatedParticipationId}`))
    expect(res.status()).toBe(404)
  })
})
