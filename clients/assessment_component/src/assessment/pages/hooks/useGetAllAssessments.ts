import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Assessment } from '../../interfaces/assessment'
import { assessmentKeys } from '../../network/cache'
import { getAllAssessmentsInPhase } from '../../network/queries/getAllAssessmentsInPhase'

const EMPTY_ASSESSMENTS: Assessment[] = []

export const useGetAllAssessments = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = options?.enabled ?? true

  const { data, ...queryInfo } = useQuery<Assessment[]>({
    queryKey: assessmentKeys.assessments.inPhase(phaseId),
    queryFn: () => getAllAssessmentsInPhase(phaseId ?? ''),
    enabled,
  })

  return { ...queryInfo, data: enabled ? (data ?? EMPTY_ASSESSMENTS) : EMPTY_ASSESSMENTS }
}
