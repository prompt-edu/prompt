import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { CreateActionItemRequest } from '../../../../../interfaces/actionItem'
import { createActionItem } from '../../../../../network/mutations/createActionItem'

export const useCreateActionItem = (setError: (error: string | undefined) => void) => {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  return useMutation({
    mutationFn: (actionItem: CreateActionItemRequest) => {
      return createActionItem(phaseId ?? '', actionItem)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actionItems', phaseId] })
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
