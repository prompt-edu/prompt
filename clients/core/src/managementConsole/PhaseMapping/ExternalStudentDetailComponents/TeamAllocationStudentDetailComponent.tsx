import { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const TeamAllocationStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('team_allocation_component/provide'),
    () => <></>,
  )
