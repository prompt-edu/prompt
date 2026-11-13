import type { Team } from '@tumaet/prompt-shared-state'

export interface TutorRow {
  // Unique per row: a tutor who leads several teams has one row per team.
  id: string
  tutorId: string
  firstName: string
  lastName: string
  teamName: string
}

export const getTutorRows = (teams: Team[]): TutorRow[] =>
  teams.flatMap((team) =>
    team.tutors
      .filter((tutor) => tutor.id)
      .map((tutor) => ({
        id: `${team.id}:${tutor.id}`,
        tutorId: tutor.id as string,
        firstName: tutor.firstName,
        lastName: tutor.lastName,
        teamName: team.name,
      })),
  )

// Each tutor once, however many teams they lead, so a tutor is counted once in the diagrams.
export const getUniqueTutorIds = (rows: TutorRow[]): string[] => [
  ...new Set(rows.map((row) => row.tutorId)),
]

export const getTeamNamesOfTutor = (teams: Team[], tutorId: string): string[] =>
  teams.filter((team) => team.tutors.some((tutor) => tutor.id === tutorId)).map((team) => team.name)
