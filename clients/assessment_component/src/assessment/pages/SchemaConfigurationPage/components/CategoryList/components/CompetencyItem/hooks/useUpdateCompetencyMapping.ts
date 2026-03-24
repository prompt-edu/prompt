import { useParams } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'

import { Competency } from '../../../../../../../interfaces/competency'
import { createCompetencyMapping } from '../../../../../../../network/mutations/createCompetencyMapping'
import { deleteCompetencyMapping } from '../../../../../../../network/mutations/deleteCompetencyMapping'

export const useUpdateCompetencyMapping = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      competency,
      newMappedCompetencyId,
      action,
      currentMapping,
    }: {
      competency: Competency
      newMappedCompetencyId: string
      action: 'add' | 'remove' | 'update'
      evaluationType?: 'self' | 'peer'
      currentMapping?: string
    }) => {
      if (action === 'update') {
        // First remove existing mapping if it exists
        if (currentMapping) {
          const removeMapping = {
            fromCompetencyId: currentMapping,
            toCompetencyId: competency.id,
          }
          await deleteCompetencyMapping(phaseId ?? '', removeMapping)
        }

        // Then add the new mapping
        const addMapping = {
          fromCompetencyId: newMappedCompetencyId,
          toCompetencyId: competency.id,
        }
        return createCompetencyMapping(phaseId ?? '', addMapping)
      } else {
        const mapping = {
          fromCompetencyId: newMappedCompetencyId,
          toCompetencyId: competency.id,
        }

        if (action === 'add') {
          return createCompetencyMapping(phaseId ?? '', mapping)
        } else {
          return deleteCompetencyMapping(phaseId ?? '', mapping)
        }
      }
    },
    onSuccess: () => {
      // Invalidate all category-related queries for the current phase
      queryClient.invalidateQueries({ queryKey: ['categories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['selfEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['peerEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['tutorEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['assessments'] })
      setError(undefined)
    },
    onError: (error: any) => {
      if (error?.response?.data?.error) {
        const serverError = error.response.data?.error
        setError(serverError)
      } else {
        setError('An unexpected error occurred while updating competency mapping.')
      }
    },
  })
}
