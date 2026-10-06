import { type CoursePhaseWithType, useCourseStore } from '@tumaet/prompt-shared-state'
import { useMemo } from 'react'

export const useCoursePhases = (courseId: string | undefined): CoursePhaseWithType[] => {
  const { courses } = useCourseStore()
  return useMemo(
    () =>
      (courses.find((course) => course.id === courseId)?.coursePhases ?? [])
        .filter((phase) => phase.sequenceOrder !== -1)
        .sort((a, b) => a.sequenceOrder - b.sequenceOrder),
    [courseId, courses],
  )
}
