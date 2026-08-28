import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentSchema } from '../../interfaces/assessmentSchema'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

export const useGetAllAssessmentSchemas = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<AssessmentSchema[]>({
    queryKey: assessmentKeys.assessmentSchemas.inPhase(phaseId),
    queryFn: () => assessmentApi.schemas.list(phaseId ?? ''),
  })
}
