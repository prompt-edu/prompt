import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import type { FeedbackItem } from '../../../../../interfaces/feedbackItem'
import { getMyFeedbackItems } from '../../../../../network/queries/getMyFeedbackItems'

export const useGetMyFeedbackItems = (options: { enabled: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<FeedbackItem[]>({
    queryKey: ['my-feedback-items', phaseId],
    queryFn: () => getMyFeedbackItems(phaseId ?? ''),
    enabled: options.enabled,
  })

  const feedbackItems = useMemo(() => data || [], [data])

  return {
    feedbackItems,
    ...queryInfo,
  }
}
