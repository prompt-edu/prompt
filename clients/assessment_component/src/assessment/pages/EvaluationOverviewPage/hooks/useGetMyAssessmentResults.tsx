import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { StudentAssessmentResults } from '../../../interfaces/assessmentResults'
import { getMyAssessmentResults } from '../../../network/queries/getMyAssessmentResults'

export const useGetMyAssessmentResults = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<StudentAssessmentResults>({
    queryKey: ['myAssessmentResults', phaseId],
    queryFn: () => getMyAssessmentResults(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
