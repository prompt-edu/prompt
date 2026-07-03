import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { AssessmentCompletion } from '../../interfaces/assessmentCompletion'
import { getAllAssessmentCompletionsInPhase } from '../../network/queries/getAllAssessmentCompletionsInPhase'

export const useGetAllAssessmentCompletions = () => {
  const { phaseId, coursePhaseID } = useParams<{ phaseId: string; coursePhaseID: string }>()
  const id = phaseId || coursePhaseID

  return useQuery<AssessmentCompletion[]>({
    queryKey: ['assessmentCompletions', id],
    queryFn: () => getAllAssessmentCompletionsInPhase(id ?? ''),
    enabled: !!id,
  })
}
