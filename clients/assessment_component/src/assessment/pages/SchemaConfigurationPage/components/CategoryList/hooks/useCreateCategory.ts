import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CreateCategoryRequest } from '../../../../../interfaces/category'
import { assessmentCache } from '../../../../../network/cache'
import { createCategory } from '../../../../../network/mutations/createCategory'

export const useCreateCategory = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (category: CreateCategoryRequest) => createCategory(phaseId ?? '', category),
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
