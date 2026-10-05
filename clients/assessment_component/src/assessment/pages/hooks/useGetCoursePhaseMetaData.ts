import { useQuery } from '@tanstack/react-query'
import { type CoursePhaseWithMetaData, getCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import { assessmentKeys } from '../../network/cache'

export const useGetCoursePhaseMetaData = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CoursePhaseWithMetaData>({
    queryKey: assessmentKeys.coursePhase(phaseId),
    queryFn: () => getCoursePhase(phaseId ?? ''),
    enabled: !!phaseId,
  })
}
