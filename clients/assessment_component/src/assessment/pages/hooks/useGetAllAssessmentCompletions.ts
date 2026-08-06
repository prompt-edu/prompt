import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { AssessmentCompletion } from '../../interfaces/assessmentCompletion'
import { getAllAssessmentCompletionsInPhase } from '../../network/queries/getAllAssessmentCompletionsInPhase'

const EMPTY_ASSESSMENT_COMPLETIONS: AssessmentCompletion[] = []

export const useGetAllAssessmentCompletions = (options?: { enabled?: boolean }) => {
  const { phaseId, coursePhaseID } = useParams<{ phaseId: string; coursePhaseID: string }>()
  const id = phaseId || coursePhaseID
  const enabled = (options?.enabled ?? true) && !!id

  const query = useQuery<AssessmentCompletion[]>({
    queryKey: ['assessmentCompletions', id],
    queryFn: () => getAllAssessmentCompletionsInPhase(id ?? ''),
    enabled,
  })

  // A disabled query stays pending forever, which would hang the loading gates above it
  if (!enabled) {
    return { ...query, data: EMPTY_ASSESSMENT_COMPLETIONS, isPending: false, isError: false }
  }

  return query
}
