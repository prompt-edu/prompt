import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { StudentAssessmentResults } from '../../../interfaces/assessmentResults'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetMyAssessmentResults = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<StudentAssessmentResults>({
    queryKey: assessmentKeys.results.myAssessment(phaseId),
    queryFn: () => assessmentApi.assessments.myResults(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
