import { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const DevOpsStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('devops_challenge_component/provide'),
    () => <></>,
  )
