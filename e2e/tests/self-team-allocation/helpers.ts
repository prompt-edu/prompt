import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { BASE_URL, SELF_TEAM_ALLOCATION_API } from '../../src/env'
import { SELF_TEAM_ALLOCATION_PHASE_ID } from '../../src/data/constants'

export interface Team {
  id: string
  name: string
  members: { id: string; firstName: string; lastName: string }[]
}

export function teamsUrl(phaseId = SELF_TEAM_ALLOCATION_PHASE_ID): string {
  return `${BASE_URL}${SELF_TEAM_ALLOCATION_API}/course_phase/${phaseId}/team`
}

export async function getTeams(
  api: APIRequestContext,
  phaseId = SELF_TEAM_ALLOCATION_PHASE_ID,
): Promise<Team[]> {
  const res = await api.get(teamsUrl(phaseId))
  if (!res.ok()) {
    throw new Error(`GET teams failed: ${res.status()} ${await res.text()}`)
  }
  return (await res.json()) as Team[]
}

// Removes a test-created team (any state), so specs stay independent and
// re-runs in UI watch mode start clean. Admin passes the lecturer-only delete.
export async function deleteTeamByName(name: string, phaseId = SELF_TEAM_ALLOCATION_PHASE_ID) {
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
