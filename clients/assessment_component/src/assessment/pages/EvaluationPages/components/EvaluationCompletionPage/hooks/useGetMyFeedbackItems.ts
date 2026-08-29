import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { FeedbackItem } from '../../../../../interfaces/feedbackItem'
import { assessmentApi } from '../../../../../network/api'
import { assessmentKeys } from '../../../../../network/cache'

export const useGetMyFeedbackItems = (options: { enabled: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<FeedbackItem[]>({
    queryKey: assessmentKeys.feedbackItems.mine(phaseId),
    queryFn: () => assessmentApi.feedbackItems.listMine(phaseId ?? ''),
    enabled: options.enabled,
  })

  const feedbackItems = useMemo(() => data || [], [data])

  return {
    feedbackItems,
    ...queryInfo,
  }
}
