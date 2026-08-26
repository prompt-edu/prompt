import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ActionItem } from '../../../interfaces/actionItem'
import { assessmentKeys } from '../../../network/cache'
import { getMyActionItems } from '../../../network/queries/getMyActionItems'

export const useGetMyActionItems = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<ActionItem[]>({
    queryKey: assessmentKeys.actionItems.mine(phaseId),
    queryFn: () => getMyActionItems(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
