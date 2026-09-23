import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { useGetCoursePhaseConfig } from './useGetCoursePhaseConfig'

export const DEFAULT_TUTOR_LABEL = 'Tutor'

export const getTutorLabel = (config?: CoursePhaseConfig) =>
  config?.tutorDisplayName || DEFAULT_TUTOR_LABEL

export const useTutorLabel = () => getTutorLabel(useGetCoursePhaseConfig().data)
