import { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const InterviewStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('interview_component/provide'),
    () => <></>,
  )
