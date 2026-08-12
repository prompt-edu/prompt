import type { AssessmentCompletion } from '../../../../../interfaces/assessmentCompletion'
import type { AssessmentParticipationWithStudent } from '../../../../../interfaces/assessmentParticipationWithStudent'
import { isValidGrade } from '../../../../utils/gradeConfig'
import type { StudentGradeEntry } from '../interfaces/campusGradeExport'

/**
 * Joins the phase participants with their final grade.
 *
 * Only completions marked as completed count, mirroring both the participants
 * table and the server's own `GetAllGrades` query. Grades outside the grade
 * vocabulary are treated as "no grade" rather than exported: the assessment
 * database still holds legacy `6.0` rows that mean exactly that.
 */
export const buildStudentGradeEntries = (
  participations: AssessmentParticipationWithStudent[],
  assessmentCompletions: AssessmentCompletion[],
): StudentGradeEntry[] => {
  const completionsByParticipation = new Map<string, AssessmentCompletion>(
    assessmentCompletions
      .filter((completion) => completion.completed)
      .map((completion) => [completion.courseParticipationID, completion]),
  )

  return participations.map((participation) => {
    const completion = completionsByParticipation.get(participation.courseParticipationID)
    const hasGrade = completion !== undefined && isValidGrade(completion.gradeSuggestion)

    return {
      courseParticipationID: participation.courseParticipationID,
      firstName: participation.student.firstName,
      lastName: participation.student.lastName,
      matriculationNumber: participation.student.matriculationNumber,
      grade: hasGrade ? completion.gradeSuggestion : null,
      completedAt: hasGrade ? completion.completedAt : null,
      passStatus: participation.passStatus,
    }
  })
}
