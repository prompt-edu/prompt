type Id = string | undefined

export const interviewKeys = {
  slots: (phaseId: Id) => ['interviewSlotsWithAssignments', phaseId] as const,
  reviews: (phaseId: Id) => ['interviewReviews', phaseId] as const,
  myAssignment: (phaseId: Id) => ['myInterviewAssignment', phaseId] as const,
  // Owned by @tumaet/prompt-shared-state: react-query is a Module Federation singleton, so these
  // two entries are shared with core and must keep their literals
  participants: (phaseId: Id) => ['participants', phaseId] as const,
  coursePhase: (phaseId: Id) => ['course_phase', phaseId] as const,
}
