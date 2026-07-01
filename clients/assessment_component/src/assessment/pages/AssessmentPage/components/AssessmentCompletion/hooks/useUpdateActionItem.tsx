import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { UpdateActionItemRequest } from '../../../../../interfaces/actionItem'
import { updateActionItem } from '../../../../../network/mutations/updateActionItem'

export const useUpdateActionItem = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (actionItem: UpdateActionItemRequest) => {
      return updateActionItem(phaseId ?? '', actionItem)
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
