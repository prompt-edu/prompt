import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentKeys } from '../../network/cache'

import { getSchemaHasAssessmentData } from '../../network/queries/getPhaseHasAssessmentData'

export const useSchemaHasAssessmentData = (schemaID: string | undefined) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery({
    queryKey: assessmentKeys.assessmentSchemas.hasAssessmentData(schemaID, phaseId),
    queryFn: () => getSchemaHasAssessmentData(schemaID!, phaseId!),
    enabled: Boolean(schemaID && phaseId),
    refetchOnWindowFocus: true,
  })
}
