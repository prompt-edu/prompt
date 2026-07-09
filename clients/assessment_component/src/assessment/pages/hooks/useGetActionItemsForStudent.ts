import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { ActionItem } from '../../interfaces/actionItem'
import { getAllActionItemsForStudentInPhase } from '../../network/queries/getAllActionItemsForStudentInPhase'

const EMPTY_ACTION_ITEMS: ActionItem[] = []

export const useGetActionItemsForStudent = (enabled = true) => {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()

  const { data, ...queryInfo } = useQuery<ActionItem[]>({
    queryKey: ['actionItems', phaseId, courseParticipationID],
    queryFn: () => getAllActionItemsForStudentInPhase(phaseId ?? '', courseParticipationID ?? ''),
    enabled: enabled && !!phaseId && !!courseParticipationID,
  })

  return { ...queryInfo, actionItems: data ?? EMPTY_ACTION_ITEMS }
}
