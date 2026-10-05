interface CourseWithPhaseTypes {
  id: string
  coursePhases: { id: string; coursePhaseType: string }[]
}

export const getApplicationParticipantPath = (
  courses: CourseWithPhaseTypes[],
  courseId: string | undefined,
  courseParticipationID: string | undefined,
): string | undefined => {
  const applicationPhaseId = courses
    .find((c) => c.id === courseId)
    ?.coursePhases.find((p) => p.coursePhaseType === 'Application')?.id
  return courseId && applicationPhaseId && courseParticipationID
    ? `/management/course/${courseId}/${applicationPhaseId}/participants/${courseParticipationID}`
    : undefined
}
