interface NamedParticipation {
  courseParticipationID: string
  student: { id?: string; firstName?: string; lastName?: string }
}

export const participantName = (
  segment: string,
  participations: NamedParticipation[],
): string | undefined => {
  const participation = participations.find(
    (p) => p.courseParticipationID === segment || p.student.id === segment,
  )
  const name = [participation?.student.firstName, participation?.student.lastName]
    .filter(Boolean)
    .join(' ')
  return name || undefined
}
