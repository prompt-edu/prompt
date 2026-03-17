import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const DevOpsStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('devops_challenge_component/provide'),
    () => <></>,
  )
