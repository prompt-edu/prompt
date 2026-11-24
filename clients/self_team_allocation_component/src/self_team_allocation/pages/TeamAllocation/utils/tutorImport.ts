import type { Student } from '@tumaet/prompt-shared-state'
import type { CourseStaff, CourseStaffMember } from '../../../interfaces/courseStaff'
import type { Tutor, TutorAssignment, TutorCandidate } from '../../../interfaces/tutor'

// The teams each person already leads: by normalized login, and by the id the rows carry for a
// person the import cannot find by login.
export interface CurrentTutorTeams {
  byLogin: Record<string, string[]>
  byId: Record<string, string[]>
}

// Matches the server (tutorscope.NormalizeLogin), which identifies a tutor by this form of the login.
export const normalizeLogin = (login: string | undefined): string =>
  (login ?? '').trim().toLowerCase()

export const studentToTutorCandidate = (student: Student): TutorCandidate => ({
  id: student.id ?? '',
  firstName: student.firstName,
  lastName: student.lastName,
  universityLogin: student.universityLogin,
})

// Course staff have no student record, so they are imported under their Keycloak user id.
export const staffMemberToTutorCandidate = (member: CourseStaffMember): TutorCandidate => ({
  id: member.keycloakUserID,
  firstName: member.firstName,
  lastName: member.lastName,
  universityLogin: member.universityLogin,
})

// A person in both the lecturer and the editor group is offered once.
export const staffToTutorCandidates = (staff: CourseStaff): TutorCandidate[] =>
  [
    ...new Map(
      [...staff.lecturers, ...staff.editors].map((member) => [member.keycloakUserID, member]),
    ).values(),
  ].map(staffMemberToTutorCandidate)

// Only someone with a university login can be scoped to their teams, so only they are selectable.
export const isSelectableTutor = (candidate: TutorCandidate): boolean =>
  !!candidate.id && normalizeLogin(candidate.universityLogin) !== ''

// The dialog selects a person by their normalized login, like the import identifies them, so the
// same person offered under two ids is one selection.
export const tutorKey = (candidate: TutorCandidate): string =>
  normalizeLogin(candidate.universityLogin)

export const currentTutorTeams = (assignments: TutorAssignment[]): CurrentTutorTeams => {
  const current: CurrentTutorTeams = { byLogin: {}, byId: {} }
  const add = (teams: Record<string, string[]>, key: string, teamID: string) => {
    teams[key] = [...(teams[key] ?? []), teamID]
  }
  for (const assignment of assignments) {
    const login = normalizeLogin(assignment.universityLogin)
    if (login) add(current.byLogin, login, assignment.teamID)
    add(current.byId, assignment.id, assignment.teamID)
  }
  return current
}

// An import replaces the teams of the tutor with the candidate's login, or, when nobody has that
// login yet, the rows stored under the candidate's id. Those are the teams the person starts from.
export const currentTeamsOf = (current: CurrentTutorTeams, candidate: TutorCandidate): string[] =>
  current.byLogin[tutorKey(candidate)] ?? current.byId[candidate.id] ?? []

export const buildTutorImport = (
  candidates: TutorCandidate[],
  selectedKeys: string[],
  teamsOf: (candidate: TutorCandidate) => string[],
): Tutor[] => {
  const seen = new Set<string>()
  return candidates
    .filter((candidate) => {
      const key = tutorKey(candidate)
      if (!isSelectableTutor(candidate) || !selectedKeys.includes(key) || seen.has(key))
        return false
      seen.add(key)
      return true
    })
    .map((candidate) => ({
      courseParticipationID: candidate.id,
      firstName: candidate.firstName,
      lastName: candidate.lastName,
      universityLogin: candidate.universityLogin ?? '',
      teamIDs: teamsOf(candidate),
    }))
}
