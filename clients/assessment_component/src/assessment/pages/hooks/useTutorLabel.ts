import { useGetCoursePhaseConfig } from './useGetCoursePhaseConfig'

export const DEFAULT_TUTOR_LABEL = 'Tutor'

export const useTutorLabel = () =>
  useGetCoursePhaseConfig().data?.tutorDisplayName || DEFAULT_TUTOR_LABEL
