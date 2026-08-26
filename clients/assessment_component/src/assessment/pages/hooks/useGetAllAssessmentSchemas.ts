import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentSchema } from '../../interfaces/assessmentSchema'
import { assessmentKeys } from '../../network/cache'
import { getAllAssessmentSchemas } from '../../network/queries/getAllAssessmentSchemas'

export const useGetAllAssessmentSchemas = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<AssessmentSchema[]>({
    queryKey: assessmentKeys.assessmentSchemas.inPhase(phaseId),
    queryFn: () => getAllAssessmentSchemas(phaseId ?? ''),
  })
}
