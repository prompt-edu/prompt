import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Assessment } from '../../interfaces/assessment'
import { getAllAssessmentsInPhase } from '../../network/queries/getAllAssessmentsInPhase'

export const useGetAllAssessments = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<Assessment[]>({
    queryKey: ['assessments', phaseId],
    queryFn: () => getAllAssessmentsInPhase(phaseId ?? ''),
  })
}
