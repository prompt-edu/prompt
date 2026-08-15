import type { ApplicationParticipation } from '@core/managementConsole/applicationAdministration/interfaces/applicationParticipation'
import type { PassStatus, Student } from '@tumaet/prompt-shared-state'

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

export const EXPORTED_ANSWER_COLUMN_PREFIX = 'exportedAnswer:'

export function buildApplicationRows(
  participations: ApplicationParticipation[] | undefined,
  additionalScores?: { key: string }[],
  exportedAnswersByParticipation?: Map<string, Map<string, string>>,
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

    ...Object.fromEntries(
      Array.from(exportedAnswersByParticipation?.get(app.courseParticipationID) ?? []).map(
        ([questionID, answer]) => [`${EXPORTED_ANSWER_COLUMN_PREFIX}${questionID}`, answer],
      ),
    ),
  }))
}
