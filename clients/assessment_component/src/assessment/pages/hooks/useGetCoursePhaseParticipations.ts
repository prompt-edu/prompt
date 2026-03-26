import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { getCoursePhaseParticipations } from '../../network/queries/getCoursePhaseParticipations'
import { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'

export const useGetCoursePhaseParticipations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<AssessmentParticipationWithStudent[]>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })
}
