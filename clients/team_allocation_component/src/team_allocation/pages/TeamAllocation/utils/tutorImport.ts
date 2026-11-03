import type { Student } from '@tumaet/prompt-shared-state'
import type { CourseStaffMember } from '../../../interfaces/courseStaff'
import type { Tutor, TutorCandidate } from '../../../interfaces/tutor'

export const studentToTutorCandidate = (student: Student): TutorCandidate => ({
  id: student.id ?? '',
  firstName: student.firstName,
  lastName: student.lastName,
  universityLogin: student.universityLogin,
})

// A course editor has no student record, so they are imported under their Keycloak user id.
export const staffMemberToTutorCandidate = (member: CourseStaffMember): TutorCandidate => ({
  id: member.keycloakUserID,
  firstName: member.firstName,
  lastName: member.lastName,
  universityLogin: member.universityLogin,
})

// Only someone with a university login can be scoped to their teams, so only they are selectable.
export const isSelectableTutor = (candidate: TutorCandidate): boolean =>
  !!candidate.id && !!candidate.universityLogin?.trim()

export const buildTutorImport = (
  candidates: TutorCandidate[],
  selectedIds: string[],
  teamAssignments: Record<string, string[]>,
): Tutor[] =>
  candidates
    .filter((candidate) => selectedIds.includes(candidate.id) && isSelectableTutor(candidate))
    .map((candidate) => ({
      courseParticipationID: candidate.id,
      firstName: candidate.firstName,
      lastName: candidate.lastName,
      universityLogin: candidate.universityLogin ?? '',
      teamIDs: teamAssignments[candidate.id] ?? [],
    }))
