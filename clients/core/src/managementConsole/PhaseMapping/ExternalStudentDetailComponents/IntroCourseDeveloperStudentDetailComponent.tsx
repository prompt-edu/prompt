import { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const IntroCourseDeveloperStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('intro_course_developer_component/provide'),
    () => <></>,
  )
