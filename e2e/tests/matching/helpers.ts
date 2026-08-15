import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { Role } from '../../src/data/roles'

export function participationsPath(phaseId: string): string {
  return `/api/course_phases/${phaseId}/participations`
}

export interface SeedStudent {
  firstName: string
  lastName: string
  matriculationNumber: string
  courseParticipationId: string
}

export interface Participation {
  courseParticipationID: string
  passStatus: string
}

export async function apiAsRole(role: Role): Promise<APIRequestContext> {
  return apiContextFor(role)
}

export function buildReimportCsv(students: SeedStudent[]): string {
  const header = 'Students first name,Students last name,Students matriculation number'
  const rows = students.map((s) => `${s.firstName},${s.lastName},${s.matriculationNumber}`)
  return [header, ...rows].join('\n')
}

export async function getParticipations(
  api: APIRequestContext,
  phaseId: string,
): Promise<Participation[]> {
  const res = await api.get(participationsPath(phaseId))
  if (!res.ok()) {
    throw new Error(`GET participations failed: ${res.status()} ${await res.text()}`)
  }
  const body = (await res.json()) as { participations: Participation[] }
  return body.participations
}

export async function resetMatchingPhase(
  phaseId: string,
  courseParticipationIds: string[],
): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    const updates = courseParticipationIds.map((courseParticipationID) => ({
      courseParticipationID,
      coursePhaseID: phaseId,
      passStatus: 'not_assessed',
      restrictedData: {},
      studentReadableData: {},
    }))
    const res = await admin.put(participationsPath(phaseId), { data: updates })
    if (!res.ok()) {
      throw new Error(`reset participations failed: ${res.status()} ${await res.text()}`)
    }
  } finally {
    await admin.dispose()
  }
}
