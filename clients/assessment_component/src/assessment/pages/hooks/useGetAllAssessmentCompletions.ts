import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentCompletion } from '../../interfaces/assessmentCompletion'
import { assessmentKeys } from '../../network/cache'
import { getAllAssessmentCompletionsInPhase } from '../../network/queries/getAllAssessmentCompletionsInPhase'

const EMPTY_ASSESSMENT_COMPLETIONS: AssessmentCompletion[] = []

export const useGetAllAssessmentCompletions = (options?: { enabled?: boolean }) => {
  const { phaseId, coursePhaseID } = useParams<{ phaseId: string; coursePhaseID: string }>()
  const id = phaseId || coursePhaseID
  const enabled = (options?.enabled ?? true) && !!id

  const { data, ...queryInfo } = useQuery<AssessmentCompletion[]>({
    queryKey: assessmentKeys.assessmentCompletions(id),
    queryFn: () => getAllAssessmentCompletionsInPhase(id ?? ''),
    enabled,
  })

  return {
    ...queryInfo,
    data: enabled ? (data ?? EMPTY_ASSESSMENT_COMPLETIONS) : EMPTY_ASSESSMENT_COMPLETIONS,
  }
}
