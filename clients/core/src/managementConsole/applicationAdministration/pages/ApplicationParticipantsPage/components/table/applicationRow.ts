import { ApplicationParticipation } from '@core/managementConsole/applicationAdministration/interfaces/applicationParticipation'
import type { Student } from '@tumaet/prompt-shared-state'
import { PassStatus } from '@tumaet/prompt-shared-state'

export interface ApplicationRow {
  id: string
  coursePhaseID: string
  courseParticipationID: string
  passStatus: PassStatus
  restrictedData: Record<string, any>
  student: Student
  score: number | null
  email: string
  studyProgram: string
  studyDegree: string
  gender: string
  [key: string]: unknown
}

export function buildApplicationRows(
  participations: ApplicationParticipation[] | undefined,
  additionalScores?: { key: string }[],
): ApplicationRow[] {
  if (!participations) return []

  return participations.map((app) => ({
    id: app.courseParticipationID,
    coursePhaseID: app.coursePhaseID,
    courseParticipationID: app.courseParticipationID,
    passStatus: app.passStatus,
    restrictedData: app.restrictedData,
    student: app.student,
    score: app.score,

    email: app.student.email ?? '',
    studyProgram: app.student.studyProgram ?? '',
    studyDegree: app.student.studyDegree ?? '',
    gender: app.student.gender ?? '',

    ...Object.fromEntries(
      (additionalScores ?? []).map((s) => [s.key, app.restrictedData?.[s.key] ?? null]),
    ),
  }))
}
