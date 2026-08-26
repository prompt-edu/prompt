import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ActionItem } from '../../../interfaces/actionItem'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetMyActionItems = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<ActionItem[]>({
    queryKey: assessmentKeys.actionItems.mine(phaseId),
    queryFn: () => assessmentApi.actionItems.listMine(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
