export type StudentNameUpdateRequest = {
  studentNamesPerID: { [courseParticipationID: string]: StudentName } // key: UUID string, value: full name
}

export type StudentName = {
  firstName: string
  lastName: string
}
