import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { deleteCategory } from '../../../../../network/mutations/deleteCategory'

export const useDeleteCategory = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (categoryID: string) => deleteCategory(phaseId ?? '', categoryID),
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
        setError('An unexpected error occurred. Please try again.')
      }
    },
  })
}
