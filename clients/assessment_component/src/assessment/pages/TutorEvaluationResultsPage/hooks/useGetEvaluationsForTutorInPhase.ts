import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Evaluation } from '../../../interfaces/evaluation'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetEvaluationsForTutorInPhase = (
  tutorParticipationID: string,
  options?: { enabled?: boolean },
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<Evaluation[]>({
    queryKey: assessmentKeys.evaluations.ofTutor(phaseId, tutorParticipationID),
    queryFn: () => assessmentApi.evaluations.ofTutor(phaseId ?? '', tutorParticipationID),
    enabled: options?.enabled && !!phaseId && !!tutorParticipationID,
  })
}
