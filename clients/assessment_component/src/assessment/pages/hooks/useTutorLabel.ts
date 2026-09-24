import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import {
  type CoursePhaseConfigQueryOptions,
  useGetCoursePhaseConfig,
} from './useGetCoursePhaseConfig'

export const DEFAULT_TUTOR_LABEL = 'Tutor'

export const getTutorLabel = (config?: CoursePhaseConfig) =>
  config?.tutorDisplayName || DEFAULT_TUTOR_LABEL

export const useTutorLabel = (options?: CoursePhaseConfigQueryOptions) =>
  getTutorLabel(useGetCoursePhaseConfig(options).data)
