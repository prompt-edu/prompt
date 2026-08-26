import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { assessmentKeys } from '../../../network/cache'

import { getFeedbackItemsForTutorInPhase } from '../../../network/queries/getFeedbackItemsForTutorInPhase'

export const useGetFeedbackItemsForTutorInPhase = (
  tutorParticipationID: string,
  options?: { enabled?: boolean },
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<FeedbackItem[]>({
    queryKey: assessmentKeys.feedbackItems.ofTutor(phaseId, tutorParticipationID),
    queryFn: () => getFeedbackItemsForTutorInPhase(phaseId ?? '', tutorParticipationID),
    enabled: options?.enabled && !!phaseId && !!tutorParticipationID,
  })
}
