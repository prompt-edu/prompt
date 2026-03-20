import { StudentCoursePhaseEnrollment } from './StudentCoursePhaseEnrollment'
import { PassStatus } from '@tumaet/prompt-shared-state'
import { useState } from 'react'
import { Button } from '@tumaet/prompt-ui-components'
import { CourseDetail } from './CourseDetail'
import { CourseEnrollment } from '@core/managementConsole/shared/interfaces/StudentEnrollment'

export function StudentCourseEnrollment({
  courseEnrollment,
  studentId,
}: {
  courseEnrollment: CourseEnrollment
  studentId: string
}) {
  const [showFuturePhases, setShowFuturePhases] = useState(false)

  const phases = courseEnrollment.coursePhases ?? []

  const failedIndex = phases.findIndex((cp) => cp.passStatus === PassStatus.FAILED)

  const currentIndex = phases.findIndex((cp) => cp.passStatus === PassStatus.NOT_ASSESSED)

  const cutoffIndex =
    failedIndex !== -1 ? failedIndex : currentIndex !== -1 ? currentIndex : phases.length - 1

  const visiblePhases = phases.slice(0, cutoffIndex + 1)
  const futurePhases = phases.slice(cutoffIndex + 1)

  const hasFailed = failedIndex !== -1
  const canShowMore = !hasFailed && futurePhases.length > 0

  return (
    <div className='w-full'>
      <CourseDetail studentCourseEnrollment={courseEnrollment} />
      <div className='mt-4 ml-3 w-full'>
        {visiblePhases.map((courseParticipation, index) => (
          <StudentCoursePhaseEnrollment
            key={courseParticipation.coursePhaseId}
            coursePhaseEnrollment={courseParticipation}
            courseEnrollment={courseEnrollment}
            studentId={studentId}
            courseId={courseEnrollment.courseId}
            current={index === currentIndex}
            showLine={
              index === phases.length - 1 || courseParticipation.passStatus === PassStatus.FAILED
            }
          />
        ))}

        {canShowMore && !showFuturePhases && (
          <Button
            variant='ghost'
            className='transform -translate-x-4 -translate-y-3 text-blue-600 text-xs'
            onClick={() => setShowFuturePhases(true)}
          >
            More
          </Button>
        )}

        {canShowMore &&
          showFuturePhases &&
          futurePhases.map((courseParticipation, index) => (
            <StudentCoursePhaseEnrollment
              key={courseParticipation.coursePhaseId}
              coursePhaseEnrollment={courseParticipation}
              courseEnrollment={courseEnrollment}
              studentId={studentId}
              courseId={courseEnrollment.courseId}
              current={false}
              showLine={
                cutoffIndex + index + 1 === phases.length - 1 ||
                courseParticipation.passStatus === PassStatus.FAILED
              }
            />
          ))}
      </div>
    </div>
  )
}
