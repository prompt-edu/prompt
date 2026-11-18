/**
 * Whether the tutor evaluation deadline has passed, as the server decides it before serving tutors
 * their results. An unset deadline arrives as Go's zero time (year 1) and counts as open.
 */
export const isTutorEvaluationDeadlinePassed = (
  deadline: Date | string | undefined,
  now: Date = new Date(),
): boolean => {
  if (!deadline) return false
  const deadlineDate = new Date(deadline)
  return deadlineDate.getUTCFullYear() > 1 && deadlineDate < now
}
