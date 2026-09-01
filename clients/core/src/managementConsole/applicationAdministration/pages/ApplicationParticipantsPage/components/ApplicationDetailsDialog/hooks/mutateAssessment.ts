import type { ApplicationAssessment } from '@core/managementConsole/applicationAdministration/interfaces/applicationAssessment'
import { coreCache } from '@core/network/cache'
import { postApplicationAssessment } from '@core/network/mutations/postApplicationAssessment'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const useModifyAssessment = (courseParticipationID: string) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (applicationAssessment: ApplicationAssessment) => {
      return postApplicationAssessment(
        phaseId ?? 'undefined',
        courseParticipationID,
        applicationAssessment,
      )
    },
    onSuccess: () => {
      coreCache.applicationAssessmentSaved(queryClient)
    },
    onError: () => {
      toast({
        title: 'Failed to Store Assessment',
        description: 'Please try again later!',
        variant: 'destructive',
      })
    },
  })
}
