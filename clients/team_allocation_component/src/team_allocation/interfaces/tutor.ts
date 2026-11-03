// One person to import as a tutor. The university login identifies the tutor: an import sets
// exactly the listed teams for that login.
export interface Tutor {
  courseParticipationID: string
  firstName: string
  lastName: string
  universityLogin: string
  teamIDs: string[]
}

// Someone the import dialog offers as a tutor, from any source (students or course staff).
export interface TutorCandidate {
  id: string
  firstName: string
  lastName: string
  universityLogin?: string
}
