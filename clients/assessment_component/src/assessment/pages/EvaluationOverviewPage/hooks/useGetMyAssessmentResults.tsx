import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { StudentAssessmentResults } from '../../../interfaces/assessmentResults'
import { assessmentKeys } from '../../../network/cache'
import { getMyAssessmentResults } from '../../../network/queries/getMyAssessmentResults'

export const useGetMyAssessmentResults = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<StudentAssessmentResults>({
    queryKey: assessmentKeys.results.myAssessment(phaseId),
    queryFn: () => getMyAssessmentResults(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
