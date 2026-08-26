import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetFeedbackItemsForTutorInPhase = (
  tutorParticipationID: string,
  options?: { enabled?: boolean },
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<FeedbackItem[]>({
    queryKey: assessmentKeys.feedbackItems.ofTutor(phaseId, tutorParticipationID),
    queryFn: () => assessmentApi.feedbackItems.ofTutor(phaseId ?? '', tutorParticipationID),
    enabled: options?.enabled && !!phaseId && !!tutorParticipationID,
  })
}
