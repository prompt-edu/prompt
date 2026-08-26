import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { FeedbackItem } from '../../../../../interfaces/feedbackItem'
import { assessmentKeys } from '../../../../../network/cache'
import { getFeedbackItemsForStudent } from '../../../../../network/queries/getFeedbackItemsForStudent'

export const useGetFeedbackItemsForStudent = (courseParticipationID: string, enabled = true) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<FeedbackItem[]>({
    queryKey: assessmentKeys.feedbackItems.ofStudent(phaseId, courseParticipationID),
    queryFn: () => getFeedbackItemsForStudent(phaseId ?? '', courseParticipationID),
    enabled: enabled && !!phaseId && !!courseParticipationID,
  })

  const feedbackItems = useMemo(
    () => data?.filter((item) => item.feedbackText !== '') || [],
    [data],
  )

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
