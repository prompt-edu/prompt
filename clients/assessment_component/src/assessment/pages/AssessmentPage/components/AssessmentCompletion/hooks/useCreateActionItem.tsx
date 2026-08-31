import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CreateActionItemRequest } from '../../../../../interfaces/actionItem'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useCreateActionItem = (setError: (error: string | undefined) => void) => {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  return useMutation({
    mutationFn: (actionItem: CreateActionItemRequest) => {
      return assessmentApi.actionItems.create(phaseId ?? '', actionItem)
    },
    onSuccess: () => {
      assessmentCache.actionItemsChanged(queryClient, phaseId)
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
