import { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const AssessmentStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('assessment_component/provide'),
    () => <>server unavailable</>,
  )
