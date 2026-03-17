import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const MatchingStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('matching_component/provide'),
    () => <></>,
  )
