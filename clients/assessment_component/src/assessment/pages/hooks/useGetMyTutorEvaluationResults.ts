import { useQuery } from '@tanstack/react-query'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import type { TutorEvaluationResults } from '../../interfaces/evaluationResults'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import { useGetCoursePhaseConfig } from './useGetCoursePhaseConfig'

/**
 * The anonymized tutor evaluation results about the signed-in staff member. Only asked for once
 * the tutor evaluation is enabled and results are released; null when the caller tutors no team.
 */
export const useGetMyTutorEvaluationResults = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { isStudentOfCourse } = useCourseStore()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()

  const enabled =
    !!phaseId &&
    !isStudentOfCourse(courseId ?? '') &&
    (coursePhaseConfig?.tutorEvaluationEnabled ?? false) &&
    (coursePhaseConfig?.resultsReleased ?? false)

  const { data, ...queryInfo } = useQuery<TutorEvaluationResults | null>({
    queryKey: assessmentKeys.results.myTutorEvaluation(phaseId),
    queryFn: () => assessmentApi.evaluations.myTutorResults(phaseId ?? ''),
    enabled,
  })

  return { ...queryInfo, enabled, data: enabled ? (data ?? null) : null }
}
