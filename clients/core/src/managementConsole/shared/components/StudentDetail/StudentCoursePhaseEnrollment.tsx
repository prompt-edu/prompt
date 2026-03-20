import { PhaseStudentDetailMapping } from '@core/managementConsole/PhaseMapping/PhaseStudentDetailMapping'
import { Suspense } from 'react'
import { ProgressIndicator } from './PhaseProgressIndicator'
import { LinkHeading } from './LinkHeading'
import { Tooltip, TooltipContent, TooltipTrigger } from '@tumaet/prompt-ui-components'
import { PassStatus } from '@tumaet/prompt-shared-state'
import {
  CourseEnrollment,
  CoursePhaseEnrollment,
} from '@core/managementConsole/shared/interfaces/StudentEnrollment'

export function parsePostgresTimestamp(ts: string): Date {
  return new Date(ts.replace(' ', 'T') + 'Z')
}

export function formatDateTime(ts: string | null): string {
  if (!ts) return ''

  const d = parsePostgresTimestamp(ts)
  return d.toLocaleString('us-US', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function StudentCoursePhaseEnrollment({
  coursePhaseEnrollment,
  courseEnrollment,
  showLine,
  current,
  studentId,
  courseId,
}: {
  coursePhaseEnrollment: CoursePhaseEnrollment
  courseEnrollment: CourseEnrollment
  showLine?: boolean
  current: boolean
  studentId: string
  courseId: string
}) {
  const PhaseDetail = PhaseStudentDetailMapping[coursePhaseEnrollment.coursePhaseType.name]
  return (
    <div className='flex gap-2' key={coursePhaseEnrollment.coursePhaseId}>
      <div className='relative flex flex-col'>
        {!showLine && (
          <div className='absolute top-0 bottom-0 w-px bg-gray-300 left-1/2 -translate-x-1/2' />
        )}
        <div className='relative z-10'>
          <Tooltip>
            <TooltipTrigger>
              <ProgressIndicator
                passStatus={current ? 'CURRENT' : coursePhaseEnrollment.passStatus}
              />
            </TooltipTrigger>
            <TooltipContent>
              <div className='text-sm text-muted-foreground'>
                {coursePhaseEnrollment.passStatus !== PassStatus.NOT_ASSESSED ? (
                  <>
                    <span className='font-semibold'>{coursePhaseEnrollment.passStatus}</span> on{' '}
                    {formatDateTime(coursePhaseEnrollment.lastModified)}
                  </>
                ) : (
                  <span>phase participation ongoing</span>
                )}
              </div>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      <div className='ml-2 mb-[0.85rem] group w-full pr-2'>
        <div>
          <LinkHeading
            targetURL={`/management/course/${courseId}/${coursePhaseEnrollment.coursePhaseId}`}
          >
            <div className='font-semibold text-lg'>{coursePhaseEnrollment.name}</div>
          </LinkHeading>
        </div>
        {PhaseDetail && (
          <div className='inline-block w-full py-2 px-3 border rounded-md empty:hidden'>
            <Suspense fallback={null}>
              <PhaseDetail
                studentId={studentId}
                coursePhaseId={coursePhaseEnrollment.coursePhaseId}
                courseId={courseId}
                courseParticipationId={courseEnrollment.courseParticipationId}
              />
            </Suspense>
          </div>
        )}
      </div>
    </div>
  )
}
