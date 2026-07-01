import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import { getCoursePhaseParticipations } from '../../network/queries/getCoursePhaseParticipations'

export const useGetCoursePhaseParticipations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<AssessmentParticipationWithStudent[]>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })
}
