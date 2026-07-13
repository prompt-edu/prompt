import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { BASE_URL, TEAM_ALLOCATION_API } from '../../src/env'

export interface Team {
  id: string
  name: string
}

// One entry of a published (TEASE-shaped) allocation set: a team and the
// course participations assigned to it.
export interface TeamAllocation {
  projectId: string
  students: string[]
}

// The allocation router's read shape: one course participation and the team it
// was allocated to.
export interface ParticipationAllocation {
  courseParticipationID: string
  teamAllocation: string
}

const coursePhaseBase = (phaseId: string) =>
  `${BASE_URL}${TEAM_ALLOCATION_API}/course_phase/${phaseId}`

export function teamsUrl(phaseId: string): string {
  return `${coursePhaseBase(phaseId)}/team`
}

export function allocationUrl(phaseId: string, courseParticipationId?: string): string {
  const base = `${coursePhaseBase(phaseId)}/allocation`
  return courseParticipationId ? `${base}/${courseParticipationId}` : base
}

function teaseSaveUrl(phaseId: string): string {
  return `${BASE_URL}${TEAM_ALLOCATION_API}/tease/course_phase/${phaseId}/save`
}

function timeframeUrl(phaseId: string): string {
  return `${coursePhaseBase(phaseId)}/survey/timeframe`
}

// Opens the survey (start in the past, deadline in the future) so the student
// survey remote renders its form instead of the "not configured" error page.
// Runs as admin (the endpoint is staff-only).
export async function openSurvey(phaseId: string): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    const now = Date.now()
    const res = await admin.put(timeframeUrl(phaseId), {
      data: {
        timeframeSet: true,
        surveyStart: new Date(now - 24 * 60 * 60 * 1000).toISOString(),
        surveyDeadline: new Date(now + 24 * 60 * 60 * 1000).toISOString(),
      },
    })
    if (!res.ok()) {
      throw new Error(`set survey timeframe failed: ${res.status()} ${await res.text()}`)
    }
  } finally {
    await admin.dispose()
  }
}

export async function getTeams(api: APIRequestContext, phaseId: string): Promise<Team[]> {
  const res = await api.get(teamsUrl(phaseId))
  if (!res.ok()) {
    throw new Error(`GET teams failed: ${res.status()} ${await res.text()}`)
  }
  const body = (await res.json()) as { teams: Team[] }
  return body.teams ?? []
}

export async function createTeam(
  api: APIRequestContext,
  phaseId: string,
  name: string,
): Promise<Team> {
  const res = await api.post(teamsUrl(phaseId), { data: { teamNames: [name] } })
  if (res.status() !== 201) {
    throw new Error(`create team failed: ${res.status()} ${await res.text()}`)
  }
  const team = (await getTeams(api, phaseId)).find((t) => t.name === name)
  if (!team) throw new Error(`created team "${name}" not found`)
  return team
}

// The external TEASE tool computes the allocation and posts it back through the
// save endpoint; here the test plays that role, publishing the workspace and
// replacing the phase's allocation set. Runs as admin (the endpoint is staff-only).
export async function publishAllocation(
  phaseId: string,
  allocations: TeamAllocation[],
): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    const res = await admin.post(teaseSaveUrl(phaseId), {
      data: { constraints: {}, lockedStudents: {}, allocations },
    })
    if (!res.ok()) {
      throw new Error(`publish allocation failed: ${res.status()} ${await res.text()}`)
    }
  } finally {
    await admin.dispose()
  }
}

export async function clearAllocations(phaseId: string): Promise<void> {
  await publishAllocation(phaseId, [])
}

export async function getAllocations(
  api: APIRequestContext,
  phaseId: string,
): Promise<ParticipationAllocation[]> {
  const res = await api.get(allocationUrl(phaseId))
  if (!res.ok()) {
    throw new Error(`GET allocations failed: ${res.status()} ${await res.text()}`)
  }
  return (await res.json()) as ParticipationAllocation[]
}

// GET /allocation/:courseParticipationID returns the allocated team's UUID as a
// bare JSON string (not an object), or 404 when the student has no allocation.
export async function getAllocatedTeamId(
  api: APIRequestContext,
  phaseId: string,
  courseParticipationId: string,
): Promise<string | null> {
  const res = await api.get(allocationUrl(phaseId, courseParticipationId))
  if (res.status() === 404) return null
  if (!res.ok()) {
    throw new Error(`GET allocation failed: ${res.status()} ${await res.text()}`)
  }
  return (await res.json()) as string
}

// Removes a test-created team (clearing its allocations first so the delete
// isn't blocked by the allocation FK), so specs stay independent and UI-watch
// re-runs start clean. Admin passes the staff-only delete.
export async function deleteTeamByName(name: string, phaseId: string): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    const team = (await getTeams(admin, phaseId)).find((t) => t.name === name)
    if (team) {
      await admin.delete(`${teamsUrl(phaseId)}/${team.id}`)
    }
  } finally {
    await admin.dispose()
  }
}
