import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { FeedbackItem } from '../../../../../interfaces/feedbackItem'
import { assessmentApi } from '../../../../../network/api'
import { assessmentKeys } from '../../../../../network/cache'

export const useGetFeedbackItemsForStudent = (courseParticipationID: string, enabled = true) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<FeedbackItem[]>({
    queryKey: assessmentKeys.feedbackItems.ofStudent(phaseId, courseParticipationID),
    queryFn: () => assessmentApi.feedbackItems.ofStudent(phaseId ?? '', courseParticipationID),
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
