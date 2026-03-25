import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { AssessmentSchema } from '../../interfaces/assessmentSchema'
import { getAllAssessmentSchemas } from '../../network/queries/getAllAssessmentSchemas'

export const useGetAllAssessmentSchemas = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<AssessmentSchema[]>({
    queryKey: ['assessmentSchemas', phaseId],
    queryFn: () => getAllAssessmentSchemas(phaseId ?? ''),
  })
}
