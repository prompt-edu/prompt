import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { FeedbackItem } from '../../../interfaces/feedbackItem'

import { getFeedbackItemsForTutorInPhase } from '../../../network/queries/getFeedbackItemsForTutorInPhase'

export const useGetFeedbackItemsForTutorInPhase = (
  tutorParticipationID: string,
  options?: { enabled?: boolean },
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<FeedbackItem[]>({
    queryKey: ['tutor-feedback-items', phaseId, tutorParticipationID],
    queryFn: () => getFeedbackItemsForTutorInPhase(phaseId ?? '', tutorParticipationID),
    enabled: options?.enabled && !!phaseId && !!tutorParticipationID,
  })
}
