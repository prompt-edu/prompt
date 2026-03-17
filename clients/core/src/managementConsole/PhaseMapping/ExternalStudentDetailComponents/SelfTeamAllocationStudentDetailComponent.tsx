import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const SelfTeamAllocationStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('self_team_allocation_component/provide'),
    () => <></>,
  )
