import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ActionItem } from '../../interfaces/actionItem'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

const EMPTY_ACTION_ITEMS: ActionItem[] = []

export const useGetActionItemsForStudent = (enabled = true) => {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()

  const { data, ...queryInfo } = useQuery<ActionItem[]>({
    queryKey: assessmentKeys.actionItems.ofParticipant(phaseId, courseParticipationID),
    queryFn: () =>
      assessmentApi.actionItems.ofParticipant(phaseId ?? '', courseParticipationID ?? ''),
    enabled: enabled && !!phaseId && !!courseParticipationID,
  })

  return { ...queryInfo, actionItems: data ?? EMPTY_ACTION_ITEMS }
}
