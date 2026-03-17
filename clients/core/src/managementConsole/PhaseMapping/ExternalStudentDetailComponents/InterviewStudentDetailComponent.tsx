import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import { safeFederatedLazyStudentDetail } from '../utils/safeFederatedLazy'

export const InterviewStudentDetailComponent =
  safeFederatedLazyStudentDetail<CoursePhaseStudentIdentifierProps>(
    () => import('interview_component/provide'),
    () => <></>,
  )
