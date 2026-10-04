import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import {
  type CoursePhaseConfigQueryOptions,
  useGetCoursePhaseConfig,
} from './useGetCoursePhaseConfig'

export const DEFAULT_TUTOR_LABEL = 'Tutor'

export interface TutorLabel {
  // For titles, headers, and labels, e.g. "Tutor Overview" or "Coach Overview".
  title: string
  // For running text: "tutor" by default, the configured name exactly as entered otherwise.
  // Sentences using it stay singular, since the plural of a custom name (e.g. "PL") is unknown.
  text: string
}

export const getTutorLabel = (config?: CoursePhaseConfig): TutorLabel => {
  const displayName = config?.tutorDisplayName
  if (displayName) return { title: displayName, text: displayName }
  return { title: DEFAULT_TUTOR_LABEL, text: DEFAULT_TUTOR_LABEL.toLowerCase() }
}

export const useTutorLabel = (options?: CoursePhaseConfigQueryOptions) =>
  getTutorLabel(useGetCoursePhaseConfig(options).data)
