import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { CoursePhaseParticipationsWithResolution } from '@tumaet/prompt-shared-state'
import { getCoursePhaseParticipations } from '@/network/queries/getCoursePhaseParticipations'

export const useGetCoursePhaseParticipants = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
    enabled: !!phaseId,
  })
}
