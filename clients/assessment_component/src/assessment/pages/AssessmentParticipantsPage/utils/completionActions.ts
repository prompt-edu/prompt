import type {
  AssessmentCompletion,
  AssessmentCompletionSkipReason,
  BatchMarkAssessmentCompletionsResult,
  BatchUnmarkAssessmentCompletionsResult,
  SkippedAssessmentCompletion,
} from '../../../interfaces/assessmentCompletion'

export interface CompletionActionSummary {
  title: string
  description?: string
  variant: 'default' | 'destructive'
}

const SKIP_REASON_TEXT: Record<AssessmentCompletionSkipReason, [string, string]> = {
  remaining_assessments: [
    'still has unassessed competencies',
    'still have unassessed competencies',
  ],
  no_completion: ['has no grade suggestion yet', 'have no grade suggestion yet'],
  already_completed: ['was already final', 'were already final'],
  not_completed: ['was not final', 'were not final'],
}

export const pluralizeAssessments = (count: number): string =>
  count === 1 ? '1 assessment' : `${count} assessments`

const completionsByParticipation = (completions: AssessmentCompletion[]) =>
  new Map(completions.map((completion) => [completion.courseParticipationID, completion]))

// Only a saved, non-final assessment can become final; the server rechecks the remaining ones
export const canMarkAnyAsCompleted = (
  courseParticipationIDs: string[],
  completions: AssessmentCompletion[],
): boolean => {
  const byParticipation = completionsByParticipation(completions)
  return courseParticipationIDs.some((id) => byParticipation.get(id)?.completed === false)
}

export const canUnmarkAny = (
  courseParticipationIDs: string[],
  completions: AssessmentCompletion[],
): boolean => {
  const byParticipation = completionsByParticipation(completions)
  return courseParticipationIDs.some((id) => byParticipation.get(id)?.completed === true)
}

const describeSkipped = (skipped: SkippedAssessmentCompletion[]): string | undefined => {
  if (skipped.length === 0) {
    return undefined
  }

  const counts = new Map<AssessmentCompletionSkipReason, number>()
  for (const { reason } of skipped) {
    counts.set(reason, (counts.get(reason) ?? 0) + 1)
  }
  const reasons = [...counts].map(
    ([reason, count]) => `${count} ${SKIP_REASON_TEXT[reason][count === 1 ? 0 : 1]}`,
  )
  return `${skipped.length} skipped: ${reasons.join(', ')}`
}

export const summarizeMarkResult = (
  result: BatchMarkAssessmentCompletionsResult,
): CompletionActionSummary => ({
  title:
    result.marked.length > 0
      ? `Marked ${pluralizeAssessments(result.marked.length)} as final`
      : 'No assessments marked as final',
  description: describeSkipped(result.skipped),
  variant: result.marked.length > 0 ? 'default' : 'destructive',
})

export const summarizeUnmarkResult = (
  result: BatchUnmarkAssessmentCompletionsResult,
): CompletionActionSummary => ({
  title:
    result.unmarked.length > 0
      ? `Reopened ${pluralizeAssessments(result.unmarked.length)} for editing`
      : 'No assessments reopened',
  description: describeSkipped(result.skipped),
  variant: result.unmarked.length > 0 ? 'default' : 'destructive',
})
