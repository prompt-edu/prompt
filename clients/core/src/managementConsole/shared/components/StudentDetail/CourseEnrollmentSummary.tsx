import { PassStatus } from '@tumaet/prompt-shared-state'
import { CourseEnrollment } from '../../interfaces/StudentEnrollment'

interface CourseEnrollmentSummaryProps {
  enrollments: CourseEnrollment[]
}

function countEnrollmentPhaseStatuses(enrollment: CourseEnrollment) {
  const hasFailed = enrollment.coursePhases.some((phase) => phase.passStatus === PassStatus.FAILED)
  const hasNotAssessed = enrollment.coursePhases.some(
    (phase) => phase.passStatus === PassStatus.NOT_ASSESSED,
  )
  return { hasFailed, hasNotAssessed }
}

export function CourseEnrollmentSummary({ enrollments }: CourseEnrollmentSummaryProps) {
  const inProgress = enrollments.filter((enrollment) => {
    const { hasFailed, hasNotAssessed } = countEnrollmentPhaseStatuses(enrollment)
    return !hasFailed && hasNotAssessed
  }).length

  const completed = enrollments.filter((enrollment) => {
    if (enrollment.coursePhases.length === 0) return false
    const { hasFailed, hasNotAssessed } = countEnrollmentPhaseStatuses(enrollment)
    return !hasFailed && !hasNotAssessed
  }).length

  const failedOrCanceled = enrollments.filter((enrollment) =>
    enrollment.coursePhases.some((phase) => phase.passStatus === PassStatus.FAILED),
  ).length

  const parts: string[] = []
  if (inProgress > 0) parts.push(`In Progress: ${inProgress}`)
  if (completed > 0) parts.push(`Completed: ${completed}`)
  if (failedOrCanceled > 0) parts.push(`Failed/Canceled: ${failedOrCanceled}`)

  if (parts.length === 0) return null

  return (
    <div className='w-full text-center -mt-2 mb-1 text-sm'>
      <p>{parts.join(' · ')}</p>
    </div>
  )
}
