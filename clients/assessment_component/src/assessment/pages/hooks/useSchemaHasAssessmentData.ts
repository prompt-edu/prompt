import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

export const useSchemaHasAssessmentData = (schemaID: string | undefined) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery({
    queryKey: assessmentKeys.assessmentSchemas.hasAssessmentData(schemaID, phaseId),
    queryFn: () => assessmentApi.schemas.hasAssessmentData(phaseId!, schemaID!),
    enabled: Boolean(schemaID && phaseId),
    refetchOnWindowFocus: true,
  })
}
