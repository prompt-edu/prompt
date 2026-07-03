import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import type { FeedbackItem } from '../../../../../interfaces/feedbackItem'
import { getMyFeedbackItems } from '../../../../../network/queries/getMyFeedbackItems'

export const useGetMyFeedbackItems = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<FeedbackItem[]>({
    queryKey: ['my-feedback-items', phaseId],
    queryFn: () => getMyFeedbackItems(phaseId ?? ''),
  })

  const feedbackItems = useMemo(() => data || [], [data])

  const positiveFeedbackItems = useMemo(
    () => feedbackItems.filter((item) => item.feedbackType === 'positive'),
    [feedbackItems],
  )

  const negativeFeedbackItems = useMemo(
    () => feedbackItems.filter((item) => item.feedbackType === 'negative'),
    [feedbackItems],
  )

  return {
    feedbackItems,
    positiveFeedbackItems,
    negativeFeedbackItems,
    ...queryInfo,
  }
}
