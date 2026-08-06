import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Assessment } from '../../interfaces/assessment'
import { getAllAssessmentsInPhase } from '../../network/queries/getAllAssessmentsInPhase'

const EMPTY_ASSESSMENTS: Assessment[] = []

export const useGetAllAssessments = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = options?.enabled ?? true

  const query = useQuery<Assessment[]>({
    queryKey: ['assessments', phaseId],
    queryFn: () => getAllAssessmentsInPhase(phaseId ?? ''),
    enabled,
  })

  // A disabled query stays pending forever, which would hang the loading gates above it
  if (!enabled) {
    return { ...query, data: EMPTY_ASSESSMENTS, isPending: false, isError: false }
  }

  return query
}
