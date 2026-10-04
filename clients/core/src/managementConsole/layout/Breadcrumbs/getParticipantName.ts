interface NamedParticipation {
  courseParticipationID: string
  student?: { id?: string; firstName?: string; lastName?: string }
}

export const getParticipantName = (
  id: string,
  participations: NamedParticipation[],
): string | undefined => {
  const student = participations.find(
    (p) => p.courseParticipationID === id || p.student?.id === id,
  )?.student
  const name = [student?.firstName, student?.lastName].filter(Boolean).join(' ')
  return name || undefined
}
