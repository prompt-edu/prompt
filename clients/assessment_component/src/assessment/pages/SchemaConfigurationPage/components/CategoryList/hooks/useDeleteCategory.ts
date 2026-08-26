import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentCache } from '../../../../../network/cache'

import { deleteCategory } from '../../../../../network/mutations/deleteCategory'

export const useDeleteCategory = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (categoryID: string) => deleteCategory(phaseId ?? '', categoryID),
    onSuccess: () => {
      assessmentCache.schemaChanged(queryClient, phaseId)
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
